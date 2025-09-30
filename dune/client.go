package dune

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// logger follows the repository logging guideline with a module field.
var logger = logrus.StandardLogger().WithField("module", "dune")

// Client is a minimal Dune API v1 client.
// Initialize with NewClient and use FetchCSV to execute a query and stream its CSV results.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewClient creates a new Dune API client. apiKey is required.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.dune.com",
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// newRequest creates an HTTP request with the Dune API key header pre-set.
func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Dune-Api-Key", c.apiKey)
	return req, nil
}

// FetchCSV triggers execution of the given query and, once completed, returns a ReadCloser
// streaming the CSV results. The caller must Close the returned ReadCloser.
func (c *Client) FetchCSV(ctx context.Context, queryID int, timeout time.Duration) (io.ReadCloser, error) {
	return c.FetchCSVWithLimit(ctx, queryID, timeout, 0)
}

// FetchCSVWithLimit behaves like FetchCSV, but if limitRows > 0 it will request the
// server-side row limit using the `limit` query parameter on the CSV endpoint.
func (c *Client) FetchCSVWithLimit(ctx context.Context, queryID int, timeout time.Duration, limitRows int) (io.ReadCloser, error) {
	if c == nil || c.apiKey == "" {
		return nil, fmt.Errorf("dune client not initialized or api key missing")
	}
	if timeout == 0 {
		timeout = 30 * time.Minute
	}

	// 1) Trigger execution
	execURL := fmt.Sprintf("%s/api/v1/query/%d/execute", c.baseURL, queryID)
	req, err := c.newRequest(ctx, http.MethodPost, execURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("execute query: status %d: %s", resp.StatusCode, string(b))
	}
	var execResp struct {
		ExecutionID string `json:"execution_id"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<10)).Decode(&execResp); err != nil {
		return nil, fmt.Errorf("parse execution id: %w", err)
	}
	if execResp.ExecutionID == "" {
		return nil, fmt.Errorf("execution_id empty in response")
	}

	// 2) Poll status until complete
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for dune execution %s", execResp.ExecutionID)
		}
		statusURL := fmt.Sprintf("%s/api/v1/execution/%s/status", c.baseURL, execResp.ExecutionID)
		req, _ := c.newRequest(ctx, http.MethodGet, statusURL, nil)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		var statusBody struct {
			State string `json:"state"`
		}
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 32<<10))
		resp.Body.Close()
		_ = json.Unmarshal(b, &statusBody) // tolerate schema changes; we only need .State
		if resp.StatusCode == 200 && statusBody.State == "QUERY_STATE_COMPLETED" {
			break
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("execution status %d: %s", resp.StatusCode, string(b))
		}
		time.Sleep(5 * time.Second)
	}

	// 3) Download CSV results (optionally limited server-side)
	csvURL := fmt.Sprintf("%s/api/v1/execution/%s/results/csv", c.baseURL, execResp.ExecutionID)
	if limitRows > 0 {
		csvURL = fmt.Sprintf("%s?limit=%d", csvURL, limitRows)
	}
	req, _ = c.newRequest(ctx, http.MethodGet, csvURL, nil)
	req.Header.Set("Accept", "text/csv")
	resp, err = c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		resp.Body.Close()
		return nil, fmt.Errorf("download csv: status %d: %s", resp.StatusCode, string(b))
	}
	logger.WithFields(logrus.Fields{"execution_id": execResp.ExecutionID, "limit": limitRows}).Debug("fetched dune csv results")

	return resp.Body, nil
}
