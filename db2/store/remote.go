package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	routeGetRowsRange = "/rowRange"
	routeGetRow       = "/row"
)

type RemoteServer struct {
	store Store
}

func NewRemoteStore(store Store) RemoteServer {
	return RemoteServer{store: store}
}

func (api RemoteServer) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(routeGetRowsRange, api.GetRowsRange)
	mux.HandleFunc(routeGetRow, api.GetRow)

	return mux
}

type ParamsGetRowsRange struct {
	High string `json:"high"`
	Low  string `json:"low"`
}

func (api RemoteServer) GetRowsRange(w http.ResponseWriter, r *http.Request) {
	var args ParamsGetRowsRange
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	rows, err := api.store.GetRowsRange(args.High, args.Low)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	data, _ := json.Marshal(rows)
	_, _ = w.Write(data)
}

type ParamsGetRow struct {
	Key string `json:"key"`
}

func (api RemoteServer) GetRow(w http.ResponseWriter, r *http.Request) {
	var args ParamsGetRow
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	row, err := api.store.GetRow(args.Key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	data, _ := json.Marshal(row)
	_, _ = w.Write(data)
}

type RemoteClient struct {
	url string
}

func NewRemoteClient(url string) *RemoteClient {
	return &RemoteClient{url: url}
}

func (r RemoteClient) Add(key, column string, data []byte, allowDuplicate bool) error {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) BulkAdd(itemsByKey map[string][]Item) error {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) Read(prefix string) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) GetRow(key string) (map[string][]byte, error) {
	b, err := json.Marshal(ParamsGetRow{Key: key})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", r.url, routeGetRow), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, b)
	}
	var row map[string][]byte
	if err := json.NewDecoder(resp.Body).Decode(&row); err != nil {
		return nil, err
	}
	return row, nil
}

func (r RemoteClient) GetRowKeys(prefix string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) GetLatestValue(key string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) GetRowsRange(high, low string) (map[string]map[string][]byte, error) {
	b, err := json.Marshal(ParamsGetRowsRange{
		High: high,
		Low:  low,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", r.url, routeGetRowsRange), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, b)
	}
	var rows map[string]map[string][]byte
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r RemoteClient) Close() error {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) Clear() error {
	//TODO implement me
	panic("implement me")
}
