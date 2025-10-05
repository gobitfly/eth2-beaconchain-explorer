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

	client := &http.Client{Timeout: time.Second * 120}

	var data []byte
	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(strings.NewReader(sql)))
		if err != nil {
			return fmt.Errorf("new request: %w", err)
		}
		// Basic auth
		req.SetBasicAuth(chCfg.Username, chCfg.Password)
		// We'll expect raw Parquet bytes in the body
		req.Header.Set("Content-Type", "text/plain; charset=UTF-8")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http do (attempt %d/5): %w", attempt, err)
			logger.WithField("attempt", attempt).Warnf("clickhouse parquet http request failed: %v", err)
		} else if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("clickhouse http status %d (attempt %d/5): %s", resp.StatusCode, attempt, string(b))
			logger.WithFields(map[string]interface{}{"attempt": attempt, "status": resp.StatusCode}).Warnf("clickhouse parquet http non-200: %s", strings.TrimSpace(string(b)))
		} else {
			logger.Info("fetched parquet http response")
			// Read body fully
			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = fmt.Errorf("read parquet http body (attempt %d/5): %w", attempt, err)
				logger.WithField("attempt", attempt).Warnf("failed reading parquet http body: %v", err)
			} else {
				data = bodyBytes
				break
			}
		}

		if attempt < 5 {
			// If context is cancelled, abort early
			select {
			case <-ctx.Done():
				return fmt.Errorf("context canceled during clickhouse fetch: %w", ctx.Err())
			case <-time.After(1 * time.Second):
			}
		}
	}
	if data == nil {
		if lastErr != nil {
			return lastErr
		}
		return fmt.Errorf("failed to fetch parquet data from clickhouse after 5 attempts")
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
