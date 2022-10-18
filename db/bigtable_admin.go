package db

import (
	"context"
	"eth2-exporter/cache"
	"eth2-exporter/utils"
	"log"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

type BigtableAdmin struct {
	client *gcp_bigtable.AdminClient
}

type CreateTables struct {
	Name    string
	ColFams []CreateFamily
}

type CreateFamily struct {
	Name   string
	Policy gcp_bigtable.GCPolicy
}

var CacheTable CreateTables = CreateTables{
	cache.TABLE_CACHE,
	[]CreateFamily{
		{
			Name:   cache.FAMILY_TEN_MINUTES,
			Policy: gcp_bigtable.IntersectionPolicy(gcp_bigtable.MaxVersionsPolicy(1), gcp_bigtable.MaxAgePolicy(time.Minute*10)),
		},
		{
			Name:   cache.FAMILY_ONE_HOUR,
			Policy: gcp_bigtable.IntersectionPolicy(gcp_bigtable.MaxVersionsPolicy(1), gcp_bigtable.MaxAgePolicy(time.Hour)),
		},
		{
			Name:   cache.FAMILY_ONE_DAY,
			Policy: gcp_bigtable.IntersectionPolicy(gcp_bigtable.MaxVersionsPolicy(1), gcp_bigtable.MaxAgePolicy(time.Hour*24)),
		},
	},
}

var BigAdminClient *BigtableAdmin

func MustInitBigtableAdmin(ctx context.Context, project, instance string) {
	admin, err := gcp_bigtable.NewAdminClient(ctx, project, instance)
	if err != nil {
		log.Fatalf("Could not create admin client: %v", err)
	}

	bta := &BigtableAdmin{
		client: admin,
	}

	BigAdminClient = bta
}

func (admin *BigtableAdmin) SetupBigtableCache() error {

	if err := admin.createTables([]CreateTables{CacheTable}); err != nil {
		log.Fatal("Error occurred trying to create tables", err)
	}
	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	for _, cf := range CacheTable.ColFams {
		if err := admin.client.SetGCPolicy(ctx, CacheTable.Name, cf.Name, cf.Policy); err != nil {
			return err
		}
	}

	return nil
}

func (admin *BigtableAdmin) TearDownCache() error {
	if err := admin.deleteTables([]CreateTables{CacheTable}); err != nil {
		return err
	}
	return nil
}

func (admin *BigtableAdmin) createTables(tables []CreateTables) error {
	ctx := context.Background()

	tableList, err := admin.client.Tables(ctx)
	if err != nil {
		log.Printf("Could not fetch table list")
		return err
	}

	for _, table := range tables {
		if !utils.SliceContains(tableList, table.Name) {
			log.Printf("Creating table %s", table)
			if err := admin.client.CreateTable(ctx, table.Name); err != nil {
				log.Printf("Could not create table %s", table.Name)
				return err
			}
		}

		tblInfo, err := admin.client.TableInfo(ctx, table.Name)
		if err != nil {
			log.Printf("Could not read info for table %s", table.Name)
			return err
		}
		for _, colfam := range table.ColFams {
			if !utils.SliceContains(tblInfo.Families, colfam.Name) {
				if err := admin.client.CreateColumnFamily(ctx, table.Name, colfam.Name); err != nil {
					log.Printf("Could not create column family %s: %v", colfam.Name, err)
					return err
				}
			}
		}
	}
	return nil
}

func (admin *BigtableAdmin) deleteTables(tables []CreateTables) error {
	ctx := context.Background()
	for _, table := range tables {
		if err := admin.client.DeleteTable(ctx, table.Name); err != nil {
			log.Printf("Could not delete table %s err %s", table, err)
			return err
		} else {
			log.Printf("Deleted Table: %v", table.Name)
		}
	}
	return nil
}
