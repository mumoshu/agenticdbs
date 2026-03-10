package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultSurrealHost = "localhost:18000"
	DefaultSurrealNS   = "agenticdbs"
	DefaultSurrealDB   = "agenticdbs"
	DefaultSurrealUser = "root"
	DefaultSurrealPass = "root"
)

// SurrealClient is an HTTP client for SurrealDB.
type SurrealClient struct {
	Host      string
	Namespace string
	Database  string
	User      string
	Pass      string
	Client    *http.Client
}

// NewSurrealClient creates a new SurrealDB client with defaults.
func NewSurrealClient() *SurrealClient {
	return &SurrealClient{
		Host:      DefaultSurrealHost,
		Namespace: DefaultSurrealNS,
		Database:  DefaultSurrealDB,
		User:      DefaultSurrealUser,
		Pass:      DefaultSurrealPass,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SurrealResponse represents a SurrealDB query response.
type SurrealResponse struct {
	Result json.RawMessage `json:"result"`
	Status string          `json:"status"`
	Time   string          `json:"time"`
}

// Query executes a SurrealQL query and returns the raw JSON response.
func (c *SurrealClient) Query(ctx context.Context, query string) (string, error) {
	url := fmt.Sprintf("http://%s/sql", c.Host)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(query))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("surreal-ns", c.Namespace)
	req.Header.Set("surreal-db", c.Database)
	req.SetBasicAuth(c.User, c.Pass)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SurrealDB returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse to check for query-level errors
	var responses []SurrealResponse
	if err := json.Unmarshal(body, &responses); err != nil {
		// May be a non-array response
		return string(body), nil
	}

	for _, r := range responses {
		if r.Status == "ERR" {
			return "", fmt.Errorf("SurrealDB query error: %s", string(r.Result))
		}
	}

	// Pretty-print
	var pretty any
	if err := json.Unmarshal(body, &pretty); err != nil {
		return string(body), nil
	}
	prettyJSON, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return string(body), nil
	}

	return string(prettyJSON), nil
}

// Health checks if SurrealDB is responding.
func (c *SurrealClient) Health(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/health", c.Host)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("SurrealDB health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SurrealDB health check returned HTTP %d", resp.StatusCode)
	}
	return nil
}
