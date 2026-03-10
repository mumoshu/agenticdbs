package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

const (
	DefaultPGHost = "localhost"
	DefaultPGPort = "15432"
	DefaultPGUser = "agenticdbs"
	DefaultPGPass = "agenticdbs"
	DefaultPGDB   = "agenticdbs"
)

// PostgresClient wraps psql for query execution.
// It can run psql locally or via docker compose exec.
type PostgresClient struct {
	Host       string
	Port       string
	User       string
	Password   string
	Database   string
	ComposeDir string // If set, runs psql via docker compose exec
}

// NewPostgresClient creates a new PostgreSQL client with defaults.
func NewPostgresClient() *PostgresClient {
	return &PostgresClient{
		Host:     DefaultPGHost,
		Port:     DefaultPGPort,
		User:     DefaultPGUser,
		Password: DefaultPGPass,
		Database: DefaultPGDB,
	}
}

// psqlArgs returns the base psql arguments for connecting.
// When using docker compose exec, we connect to localhost:5432 inside the container.
func (c *PostgresClient) psqlCommand(ctx context.Context, extraArgs ...string) *exec.Cmd {
	if c.ComposeDir != "" {
		// Run psql inside the postgres container
		args := []string{"compose", "exec", "-T", "postgres",
			"psql", "-U", c.User, "-d", c.Database}
		args = append(args, extraArgs...)
		cmd := exec.CommandContext(ctx, "docker", args...)
		cmd.Dir = c.ComposeDir
		return cmd
	}
	// Run psql locally
	args := []string{"-h", c.Host, "-p", c.Port, "-U", c.User, "-d", c.Database}
	args = append(args, extraArgs...)
	cmd := exec.CommandContext(ctx, "psql", args...)
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("PGPASSWORD=%s", c.Password))
	return cmd
}

// Query executes a SQL query via psql and returns the output.
func (c *PostgresClient) Query(ctx context.Context, query string) (string, error) {
	// Prepend AGE setup for graph queries
	fullQuery := query
	if strings.Contains(strings.ToLower(query), "cypher(") {
		fullQuery = fmt.Sprintf("LOAD 'age';\nSET search_path = ag_catalog, \"$user\", public;\n%s", query)
	}

	cmd := c.psqlCommand(ctx, "-t", "-A", "-F", "|", "-c", fullQuery)
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		if output != "" {
			return "", fmt.Errorf("PostgreSQL query error: %s\n%s", err, output)
		}
		return "", fmt.Errorf("PostgreSQL query error: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// QueryJSON executes a query and returns results as JSON.
func (c *PostgresClient) QueryJSON(ctx context.Context, query string) (string, error) {
	// Wrap in a JSON-producing query
	jsonQuery := fmt.Sprintf("SELECT json_agg(t) FROM (%s) t", query)

	// Prepend AGE setup for graph queries
	if strings.Contains(strings.ToLower(query), "cypher(") {
		jsonQuery = fmt.Sprintf("LOAD 'age';\nSET search_path = ag_catalog, \"$user\", public;\n%s", jsonQuery)
	}

	cmd := c.psqlCommand(ctx, "-t", "-A", "-c", jsonQuery)
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		if output != "" {
			return "", fmt.Errorf("PostgreSQL query error: %s\n%s", err, output)
		}
		return "", fmt.Errorf("PostgreSQL query error: %w", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "" || result == "(null)" {
		return "[]", nil
	}

	// Pretty-print JSON
	var pretty any
	if err := json.Unmarshal([]byte(result), &pretty); err != nil {
		return result, nil
	}
	prettyJSON, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return result, nil
	}

	return string(prettyJSON), nil
}

// Health checks if PostgreSQL is accepting connections.
func (c *PostgresClient) Health(ctx context.Context) error {
	if c.ComposeDir != "" {
		// Use docker compose exec to check health
		cmd := exec.CommandContext(ctx, "docker", "compose", "exec", "-T", "postgres",
			"pg_isready", "-U", c.User)
		cmd.Dir = c.ComposeDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("PostgreSQL health check failed: %s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}

	// Try pg_isready locally first
	cmd := exec.CommandContext(ctx, "pg_isready",
		"-h", c.Host,
		"-p", c.Port,
		"-U", c.User,
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// If pg_isready is not found, try a TCP connection check
	if isExecNotFound(err) {
		return checkTCPHealth(c.Host, c.Port)
	}

	return fmt.Errorf("PostgreSQL health check failed: %s: %s", err, strings.TrimSpace(string(out)))
}

func isExecNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "executable file not found")
}

func checkTCPHealth(host, port string) error {
	addr := host + ":" + port
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("PostgreSQL health check failed: cannot connect to %s: %w", addr, err)
	}
	_ = conn.Close()
	return nil
}

// ExecFile executes SQL content via psql.
func (c *PostgresClient) ExecFile(ctx context.Context, sqlContent string) error {
	cmd := c.psqlCommand(ctx, "-v", "ON_ERROR_STOP=1")
	cmd.Stdin = strings.NewReader(sqlContent)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("psql execution error: %s\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
