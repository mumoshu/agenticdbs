package e2e

import (
	"strings"
	"testing"
)

// =============================================================
// Goal 3: Develop Apps Tests
// UC3.1–UC3.3: Direct access code, agentic access code, pattern recommendation
// =============================================================

// UC3.1: Generate direct-access code for PostgreSQL with vector search
func TestAppDev_DirectAccess(t *testing.T) {
	output := runClaude(t, `Using the /db skill, generate Python code to connect directly to the PostgreSQL backend and perform a vector similarity search for products. The code should use psycopg and pgvector to find products similar to a given embedding. Include the connection parameters: host=localhost, port=5432, dbname=agenticdbs, user=agenticdbs, password=agenticdbs.`)

	lowerOutput := strings.ToLower(output)

	// Should contain Python code with psycopg
	assertContainsAny(t, lowerOutput, []string{"psycopg", "psycopg2", "import psycopg"}, "should use psycopg library")

	// Should include connection to PostgreSQL
	assertContainsAny(t, output, []string{"localhost", "5432", "agenticdbs"}, "should include connection parameters")

	// Should include vector search query with pgvector operator
	assertContainsAny(t, output, []string{"<=>", "vector", "embedding", "ORDER BY"}, "should include pgvector similarity query")

	// Should be actual code (has function/variable definitions or SQL)
	assertContainsAny(t, output, []string{"def ", "conn", "cursor", "execute", "SELECT"}, "should contain actual Python code")

	// Should reference the products table
	assertContainsAny(t, lowerOutput, []string{"products", "product"}, "should reference products table")
}

// UC3.2: Generate agentic-access code using agenticdbs CLI
func TestAppDev_AgenticAccess(t *testing.T) {
	output := runClaude(t, `Using the /db skill, generate code for an application that uses the agenticdbs CLI to flexibly query the database. Show how the app can shell out to 'agenticdbs query --backend surreal' or 'agenticdbs query --backend postgres' to execute natural-language-driven queries. Use Python as the language.`)

	lowerOutput := strings.ToLower(output)

	// Should reference the agenticdbs CLI
	assertContainsAny(t, output, []string{"agenticdbs query", "agenticdbs"}, "should reference agenticdbs CLI")

	// Should show shell execution
	assertContainsAny(t, lowerOutput, []string{"subprocess", "os.system", "popen", "shell", "exec"}, "should use subprocess or shell execution")

	// Should reference --backend flag
	assertContainsAny(t, output, []string{"--backend surreal", "--backend postgres", "--backend"}, "should use --backend flag")

	// Should be actual code
	assertContainsAny(t, output, []string{"def ", "import ", "class ", "subprocess"}, "should contain actual Python code")

	// Should demonstrate flexibility (natural language or dynamic queries)
	assertContainsAny(t, lowerOutput, []string{"flexible", "dynamic", "natural language", "query", "agentic"}, "should demonstrate flexible query pattern")
}

// UC3.3: Recommend access pattern based on latency requirements
func TestAppDev_RecommendPattern(t *testing.T) {
	output := runClaude(t, `Using the /db skill, I need to build a real-time dashboard that displays live metrics with sub-100ms query latency. Should I use the agentic access pattern (via agenticdbs CLI / Claude) or direct backend access? Explain the trade-offs.`)

	lowerOutput := strings.ToLower(output)

	// Should recommend direct access for low-latency scenarios
	assertContainsAny(t, lowerOutput, []string{"direct", "direct access", "directly", "native client"}, "should recommend direct access for low latency")

	// Should discuss latency trade-off
	assertContainsAny(t, lowerOutput, []string{"latency", "performance", "speed", "overhead", "millisecond", "100ms"}, "should discuss latency considerations")

	// Should mention that agentic adds overhead
	assertContainsAny(t, lowerOutput, []string{"overhead", "slower", "additional", "ai", "llm", "agent"}, "should explain agentic overhead")

	// Should mention TimescaleDB as good fit for metrics/dashboard
	assertContainsAny(t, lowerOutput, []string{"timescaledb", "timescale", "time_bucket", "hypertable", "postgresql", "postgres"}, "should recommend PostgreSQL/TimescaleDB for metrics dashboard")

	// Should provide both options with clear recommendation
	assertContainsAny(t, lowerOutput, []string{"recommend", "suggest", "best", "should use", "ideal"}, "should make a clear recommendation")
}
