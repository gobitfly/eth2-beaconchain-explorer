package store

type Store interface {
	Add(key, column string, data []byte, allowDuplicate bool) error
	BulkAdd(itemsByKey map[string][]Item) error
	Read(prefix string) ([][]byte, error)
	GetRow(key string) (map[string][]byte, error)
	GetRowKeys(prefix string) ([]string, error)
	GetLatestValue(key string) ([]byte, error)
	GetRowsRange(high, low string) (map[string]map[string][]byte, error)
	Close() error
	Clear() error
}

var (
	_ Store = (*TableWrapper)(nil)
	_ Store = (*RemoteClient)(nil)
)
