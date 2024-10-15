package store

import (
	"fmt"
	"strings"
)

// TODO probably drop that type, we have storetest for mem bigtable
type Mem struct {
	data map[string]map[string][]byte
}

func NewMem() *Mem {
	return &Mem{make(map[string]map[string][]byte)}
}

func (m *Mem) Add(key string, column string, data []byte, allowDuplicate bool) error {
	if m.data[key] == nil {
		m.data[key] = make(map[string][]byte)
	}
	if _, exist := m.data[key][column]; exist && !allowDuplicate {
		return nil
	}
	m.data[key][column] = data
	return nil
}

func (m *Mem) BulkAdd(itemsByKey map[string][]Item) error {
	for key, items := range itemsByKey {
		for _, item := range items {
			if err := m.Add(key, item.Column, item.Data, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Mem) Read(prefix string) ([][]byte, error) {
	var data [][]byte
	for _, columns := range m.data {
		for row, value := range columns {
			if !strings.Contains(row, prefix) {
				continue
			}
			data = append(data, value)
		}
	}
	return data, nil
}

func (m *Mem) GetRowKeys(prefix string) ([]string, error) {
	var keys []string

	for key := range m.data {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (m *Mem) GetRow(key string) (map[string][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Mem) GetLatestValue(key string) ([]byte, error) {
	for _, v := range m.data[key] {
		return v, nil
	}

	return nil, fmt.Errorf("not found")
}

func (m *Mem) Clear() error {
	m.data = make(map[string]map[string][]byte)
	return nil
}

func (m *Mem) Close() error {
	m.data = nil
	return nil
}
