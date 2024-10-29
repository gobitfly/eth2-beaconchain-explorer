package store

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"cloud.google.com/go/bigtable"
	"golang.org/x/exp/maps"
)

var ErrNotFound = fmt.Errorf("not found")

const (
	timeout = time.Minute // Timeout duration for Bigtable operations
)

type TableWrapper struct {
	*BigTableStore
	table  string
	family string
}

func Wrap(db *BigTableStore, table string, family string) TableWrapper {
	return TableWrapper{
		BigTableStore: db,
		table:         table,
		family:        family,
	}
}

func (w TableWrapper) Add(key, column string, data []byte, allowDuplicate bool) error {
	return w.BigTableStore.Add(w.table, w.family, key, column, data, allowDuplicate)
}

func (w TableWrapper) Read(prefix string) ([][]byte, error) {
	return w.BigTableStore.Read(w.table, w.family, prefix)
}

func (w TableWrapper) GetLatestValue(key string) ([]byte, error) {
	return w.BigTableStore.GetLatestValue(w.table, w.family, key)
}

func (w TableWrapper) GetRow(key string) (map[string][]byte, error) {
	return w.BigTableStore.GetRow(w.table, key)
}

func (w TableWrapper) GetRowKeys(prefix string) ([]string, error) {
	return w.BigTableStore.GetRowKeys(w.table, prefix)
}

func (w TableWrapper) BulkAdd(itemsByKey map[string][]Item) error {
	return w.BigTableStore.BulkAdd(w.table, itemsByKey)
}

func (w TableWrapper) GetRowsRange(high, low string) (map[string]map[string][]byte, error) {
	return w.BigTableStore.GetRowsRange(w.table, high, low)
}

// BigTableStore is a wrapper around Google Cloud Bigtable for storing and retrieving data
type BigTableStore struct {
	client *bigtable.Client
	admin  *bigtable.AdminClient
}

func NewBigTableWithClient(ctx context.Context, client *bigtable.Client, adminClient *bigtable.AdminClient, tablesAndFamilies map[string][]string) (*BigTableStore, error) {
	// Initialize the Bigtable table and column family
	if err := initTable(ctx, adminClient, tablesAndFamilies); err != nil {
		return nil, err
	}

	return &BigTableStore{client: client, admin: adminClient}, nil
}

// NewBigTable initializes a new BigTableStore
// It returns a BigTableStore and an error if any part of the setup fails
func NewBigTable(project, instance string, tablesAndFamilies map[string][]string) (*BigTableStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create an admin client to manage Bigtable tables
	adminClient, err := bigtable.NewAdminClient(ctx, project, instance)
	if err != nil {
		return nil, fmt.Errorf("could not create admin client: %v", err)
	}

	// Create a Bigtable client for performing data operations
	client, err := bigtable.NewClient(ctx, project, instance)
	if err != nil {
		return nil, fmt.Errorf("could not create data operations client: %v", err)
	}

	return NewBigTableWithClient(ctx, client, adminClient, tablesAndFamilies)
}

// initTable creates the tables and column family in the Bigtable
func initTable(ctx context.Context, adminClient *bigtable.AdminClient, tablesAndFamilies map[string][]string) error {
	for table, families := range tablesAndFamilies {
		if err := createTableAndFamilies(ctx, adminClient, table, families...); err != nil {
			return err
		}
	}
	return nil
}

func createTableAndFamilies(ctx context.Context, admin *bigtable.AdminClient, tableName string, familyNames ...string) error {
	// Get the list of existing tables
	tables, err := admin.Tables(ctx)
	if err != nil {
		return fmt.Errorf("could not fetch table list: %v", err)
	}

	// Create the table if it doesn't exist
	if !slices.Contains(tables, tableName) {
		if err := admin.CreateTable(ctx, tableName); err != nil {
			return fmt.Errorf("could not create table %s: %v", tableName, err)
		}
	}

	// Retrieve information about the table
	tblInfo, err := admin.TableInfo(ctx, tableName)
	if err != nil {
		return fmt.Errorf("could not read info for table %s: %v", tableName, err)
	}

	for _, familyName := range familyNames {
		// Create the column family if it doesn't exist
		if !slices.Contains(tblInfo.Families, familyName) {
			if err := admin.CreateColumnFamily(ctx, tableName, familyName); err != nil {
				return fmt.Errorf("could not create column family %s: %v", familyName, err)
			}
		}
	}
	return nil
}

type Item struct {
	Family string
	Column string
	Data   []byte
}

func (b BigTableStore) BulkAdd(table string, itemsByKey map[string][]Item) error {
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var muts []*bigtable.Mutation
	for _, items := range itemsByKey {
		mut := bigtable.NewMutation()
		for _, item := range items {
			mut.Set(item.Family, item.Column, bigtable.Timestamp(0), item.Data)
		}
		muts = append(muts, mut)
	}
	errs, err := tbl.ApplyBulk(ctx, maps.Keys(itemsByKey), muts)
	if err != nil {
		return fmt.Errorf("cannot ApplyBulk err: %w", err)
	}
	// TODO aggregate errs
	for _, e := range errs {
		return fmt.Errorf("cannot ApplyBulk elem err: %w", e)
	}
	return nil
}

// Add inserts a new row with the given key, column, and data into the Bigtable
// It applies a mutation that stores data in the receiver column family
// It returns error if the operation fails
func (b BigTableStore) Add(table, family string, key string, column string, data []byte, allowDuplicate bool) error {
	// Open the transfer table for data operations
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a new mutation to store data in the given column
	mut := bigtable.NewMutation()
	mut.Set(family, column, bigtable.Now(), data)

	if !allowDuplicate {
		mut = bigtable.NewCondMutation(bigtable.RowKeyFilter(key), nil, mut)
	}
	// Apply the mutation to the table using the given key
	if err := tbl.Apply(ctx, key, mut); err != nil {
		return fmt.Errorf("could not apply row mutation: %v", err)
	}
	return nil
}

// Read retrieves all rows from the Bigtable's receiver column family
// It returns the data in the form of a 2D byte slice and an error if the operation fails
func (b BigTableStore) Read(table, family, prefix string) ([][]byte, error) {
	// Open the transfer table for reading
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var data [][]byte
	// Read all rows from the table and collect values from the receiver column family
	err := tbl.ReadRows(ctx, bigtable.PrefixRange(prefix), func(row bigtable.Row) bool {
		for _, item := range row[family] {
			// Append each value from the receiver family to the data slice
			data = append(data, item.Value)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}

	return data, nil
}

func (b BigTableStore) GetLatestValue(table, family, key string) ([]byte, error) {
	// Open the transfer table for reading
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var data []byte
	err := tbl.ReadRows(ctx, bigtable.PrefixRange(key), func(row bigtable.Row) bool {
		data = row[family][0].Value
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}

	return data, nil
}

func (b BigTableStore) GetRow(table, key string) (map[string][]byte, error) {
	// Open the transfer table for reading
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data := make(map[string][]byte)
	err := tbl.ReadRows(ctx, bigtable.PrefixRange(key), func(row bigtable.Row) bool {
		for _, family := range row {
			for _, item := range family {
				data[item.Column] = item.Value
			}
		}
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}
	if len(data) == 0 {
		return nil, ErrNotFound
	}

	return data, nil
}

func (b BigTableStore) GetRowsRange(table, high, low string) (map[string]map[string][]byte, error) {
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rowRange := bigtable.NewRange(low, high)
	data := make(map[string]map[string][]byte)
	err := tbl.ReadRows(ctx, rowRange, func(row bigtable.Row) bool {
		data[row.Key()] = make(map[string][]byte)
		for _, family := range row {
			for _, item := range family {
				data[row.Key()][item.Column] = item.Value
			}
		}
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}
	if len(data) == 0 {
		return nil, ErrNotFound
	}

	return data, nil
}

func (b BigTableStore) GetRowKeys(table, prefix string) ([]string, error) {
	// Open the transfer table for reading
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var data []string
	// Read all rows from the table and collect all the row keys
	err := tbl.ReadRows(ctx, bigtable.PrefixRange(prefix), func(row bigtable.Row) bool {
		data = append(data, row.Key())
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}

	return data, nil
}

func (b BigTableStore) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tables, err := b.admin.Tables(ctx)
	if err != nil {
		return err
	}
	for _, table := range tables {
		if err := b.admin.DropAllRows(ctx, table); err != nil {
			return fmt.Errorf("could not drop all rows: %v", err)
		}
	}
	return nil
}

// Close shuts down the BigTableStore by closing the Bigtable client connection
// It returns an error if the operation fails
func (b BigTableStore) Close() error {
	if err := b.client.Close(); err != nil {
		return fmt.Errorf("could not close client: %v", err)
	}
	if err := b.admin.Close(); err != nil {
		if !strings.Contains(err.Error(), "the client connection is closing") {
			return fmt.Errorf("could not close admin client: %v", err)
		}
	}

	return nil
}
