package db

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/parquet-go/parquet-go"
)

// FetchClickhouseParquet executes the provided SQL (must include `FORMAT Parquet`) against
// ClickHouse over HTTP(S) and streams Parquet rows of type T to the provided yield callback.
// The full HTTP response is buffered in memory to satisfy parquet-go's reader requirements.
// The yield callback should return true to continue or false to stop early.
func FetchClickhouseParquet[T any](ctx context.Context, sql string, yield func(T) bool) error {
	cfg := utils.Config
	if !cfg.ClickHouseEnabled {
		return fmt.Errorf("clickhouse is not enabled in config")
	}
	chCfg := cfg.ClickHouse.ReaderDatabase
	if chCfg.Host == "" || chCfg.Port == "" || chCfg.Name == "" || chCfg.Username == "" {
		return fmt.Errorf("incomplete ClickHouse HTTP config (host/port/name/username)")
	}

	// Use HTTPS by default for ClickHouse HTTP interface
	scheme := "https"
	url := fmt.Sprintf("%s://%s:%d/?database=%s&enable_http_compression=1", scheme, chCfg.Host, 8443, chCfg.Name)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(strings.NewReader(sql)))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	// Basic auth
	req.SetBasicAuth(chCfg.Username, chCfg.Password)
	// We'll expect raw Parquet bytes in the body
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")

	client := &http.Client{Timeout: time.Second * 120}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return fmt.Errorf("clickhouse http status %d: %s", resp.StatusCode, string(b))
	}
	logger.Info("fetched parquet http response")
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read parquet http body: %w", err)
	}
	// print the body size in megabytes
	logger.Infof("fetched %d Mb of parquet data", len(data)/1024/1024)
	reader := parquet.NewReader(bytes.NewReader(data))
	defer reader.Close()
	logger.Info("initialized parquet reader from memory buffer")

	for {
		var v T
		if err := reader.Read(&v); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read parquet row: %w", err)
		}
		if !yield(v) {
			break
		}
	}
	return nil
}
