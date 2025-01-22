package store

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/gobitfly/eth2-beaconchain-explorer/db2/storetest"
)

func TestBigTableStore(t *testing.T) {
	type item struct {
		key    string
		column string
		data   string
	}
	tests := []struct {
		name     string
		bulk     bool
		items    []item
		expected []string
	}{
		{
			name: "simple add",
			items: []item{{
				key:    "foo",
				column: "bar",
				data:   "foobar",
			}},
			expected: []string{"foobar"},
		},
		{
			name: "bulk add",
			bulk: true,
			items: []item{{
				key:    "key1",
				column: "col1",
				data:   "foobar",
			}, {
				key:    "key2",
				column: "col2",
				data:   "foobar",
			}, {
				key:    "key3",
				column: "col3",
				data:   "foobar",
			}},
			expected: []string{"foobar", "foobar", "foobar"},
		},
		{
			name: "dont duplicate",
			items: []item{{
				key:    "foo",
				column: "bar",
				data:   "foobar",
			}, {
				key:    "foo",
				column: "bar",
				data:   "foobar",
			}},
			expected: []string{"foobar"},
		},
		{
			name: "with a prefix",
			items: []item{{
				key: "foo",
			}, {
				key: "foofoo",
			}, {
				key: "foofoofoo",
			}, {
				key: "bar",
			}},
			expected: []string{"", "", "", ""},
		},
	}
	tables := map[string][]string{"testTable": {"testFamily"}}
	client, admin := storetest.NewBigTable(t)
	store, err := NewBigTableWithClient(context.Background(), client, admin, tables)
	if err != nil {
		t.Fatal(err)
	}
	db := Wrap(store, "testTable", "testFamily")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				_ = db.Clear()
			}()

			if tt.bulk {
				itemsByKey := make(map[string][]Item)
				for _, item := range tt.items {
					itemsByKey[item.key] = append(itemsByKey[item.key], Item{
						Family: "testFamily",
						Column: item.column,
						Data:   []byte(item.data),
					})
				}
				if err := db.BulkAdd(itemsByKey); err != nil {
					t.Error(err)
				}
			} else {
				for _, it := range tt.items {
					if err := db.Add(it.key, it.column, []byte(it.data), false); err != nil {
						t.Error(err)
					}
				}
			}

			t.Run("Read", func(t *testing.T) {
				res, err := db.Read("")
				if err != nil {
					t.Error(err)
				}
				if got, want := len(res), len(tt.expected); got != want {
					t.Errorf("got %v want %v", got, want)
				}
				for _, data := range res {
					if !slices.Contains(tt.expected, string(data)) {
						t.Errorf("wrong data %s", data)
					}
				}
			})

			t.Run("GetLatestValue", func(t *testing.T) {
				for _, it := range tt.items {
					v, err := db.GetLatestValue(it.key)
					if err != nil {
						t.Error(err)
					}
					if got, want := string(v), it.data; got != want {
						t.Errorf("got %v want %v", got, want)
					}
				}
			})

			t.Run("GetRowKeys", func(t *testing.T) {
				for _, it := range tt.items {
					keys, err := db.GetRowKeys(it.key)
					if err != nil {
						t.Error(err)
					}
					count, found := 0, false
					for _, expected := range tt.items {
						if !strings.HasPrefix(expected.key, it.key) {
							continue
						}
						// don't count duplicate inputs since the add prevent duplicate keys
						if expected.key == it.key && found {
							continue
						}
						found = expected.key == it.key
						count++
						if !slices.Contains(keys, expected.key) {
							t.Errorf("missing %v in %v", expected.key, keys)
						}
					}
					if got, want := len(keys), count; got != want {
						t.Errorf("got %v want %v", got, want)
					}
				}
			})
		})
	}

	if err := db.Close(); err != nil {
		t.Errorf("cannot close db: %v", err)
	}
}

func TestRangeIncludeLimits(t *testing.T) {
	tables := map[string][]string{"testTable": {"testFamily"}}
	client, admin := storetest.NewBigTable(t)
	store, err := NewBigTableWithClient(context.Background(), client, admin, tables)
	if err != nil {
		t.Fatal(err)
	}
	db := Wrap(store, "testTable", "testFamily")

	db.Add("1:999999999999", "", []byte("0"), false)
	db.Add("1:999999999998", "", []byte("1"), false)

	rows, err := db.GetRowsRange("1:999999999999", "1:999999999998")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(rows), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}
}
