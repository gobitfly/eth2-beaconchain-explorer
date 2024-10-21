package store

type Store interface {
	Add(key, column string, data []byte, allowDuplicate bool) error
	BulkAdd(itemsByKey map[string][]Item) error
	Read(prefix string) ([][]byte, error)
	GetRow(key string) (map[string][]byte, error)
	GetRowKeys(prefix string) ([]string, error)
	GetRows(table string, keys []string) (map[string]map[string][]byte, error)
	GetLatestValue(key string) ([]byte, error)
	Close() error
	Clear() error
}

var (
	_ Store = (*Mem)(nil)
	_ Store = (*TableWrapper)(nil)
)
