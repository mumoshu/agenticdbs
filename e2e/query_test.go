package e2e

import (
	"strings"
	"testing"
)

// =============================================================
// Goal 1: Query DB Tests
// UC1.1–UC1.6 tested on both backends, UC1.7 auto-recommend, UC1.8 debug
// =============================================================

// UC1.1: Relational query on SurrealDB
func TestQuery_Relational_Surreal(t *testing.T) {
	output := runClaude(t, `Using the /db skill, query all users with their orders from SurrealDB. Execute the query using agenticdbs query --backend surreal and show me the results.`)
	lowerOutput := strings.ToLower(output)

	// Verify: must contain actual user data from seed
	assertContainsAny(t, output, []string{"Alice Johnson", "Alice", "alice"}, "should return user names")
	// Verify SurrealDB was targeted (Claude may mention it in explanation)
	assertContainsAny(t, lowerOutput, []string{"surrealdb", "surreal", "orders:1", "users:alice"}, "should target SurrealDB backend")
	// Verify it's actually a relational query result (has order data)
	assertContainsAny(t, output, []string{"orders", "order", "Headphones", "Keyboard", "79.99", "129.99"}, "should show order data")
}

// UC1.1: Relational query on PostgreSQL
func TestQuery_Relational_Postgres(t *testing.T) {
	output := runClaude(t, `Using the /db skill, query all users with their orders from PostgreSQL. Execute the query using agenticdbs query --backend postgres and show me the results.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, output, []string{"Alice Johnson", "Alice", "alice"}, "should return user names")
	assertContainsAny(t, lowerOutput, []string{"postgresql", "postgres", "left join", "join", "sql"}, "should target PostgreSQL backend")
	assertContainsAny(t, output, []string{"orders", "order", "Headphones", "Keyboard", "79.99", "129.99"}, "should show order data")
}

// UC1.2: Graph traversal on SurrealDB
func TestQuery_Graph_Surreal(t *testing.T) {
	output := runClaude(t, `Using the /db skill, find all friends of Alice using graph traversal in SurrealDB. Execute the query with agenticdbs query --backend surreal.`)
	lowerOutput := strings.ToLower(output)

	// Alice knows Bob and Carol in the seed data
	assertContainsAny(t, output, []string{"Bob", "Carol", "bob", "carol"}, "should find Alice's friends Bob and/or Carol")
	assertContainsAny(t, lowerOutput, []string{"knows", "graph", "traversal", "friend", "->", "relate"}, "should reference graph concepts")
}

// UC1.2: Graph traversal on PostgreSQL
func TestQuery_Graph_Postgres(t *testing.T) {
	output := runClaude(t, `Using the /db skill, find all friends of Alice using graph traversal in PostgreSQL. Execute the query with agenticdbs query --backend postgres.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, output, []string{"Bob", "Carol", "bob", "carol"}, "should find Alice's friends")
	assertContainsAny(t, lowerOutput, []string{"cypher", "match", "knows", "graph", "age"}, "should use AGE cypher or graph query")
}

// UC1.3: Vector similarity search on SurrealDB
func TestQuery_Vector_Surreal(t *testing.T) {
	output := runClaude(t, `Using the /db skill, find products similar to the 'Wireless Bluetooth Headphones' using vector search in SurrealDB. Execute with agenticdbs query --backend surreal.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, lowerOutput, []string{"mouse", "keyboard", "ergonomic", "mechanical", "backpack", "novel", "similar"}, "should find similar products")
	assertContainsAny(t, lowerOutput, []string{"vector", "similarity", "cosine", "embedding", "hnsw", "distance"}, "should use vector search")
}

// UC1.3: Vector similarity search on PostgreSQL
func TestQuery_Vector_Postgres(t *testing.T) {
	output := runClaude(t, `Using the /db skill, find products similar to the 'Wireless Bluetooth Headphones' (product id 1) using vector search in PostgreSQL. Execute with agenticdbs query --backend postgres.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, lowerOutput, []string{"mouse", "keyboard", "ergonomic", "mechanical", "backpack", "novel", "similar"}, "should find similar products")
	assertContainsAny(t, lowerOutput, []string{"<=>", "vector", "embedding", "cosine", "pgvector", "distance"}, "should use pgvector search")
}

// UC1.4: Timeseries aggregation on SurrealDB
func TestQuery_Timeseries_Surreal(t *testing.T) {
	output := runClaude(t, `Using the /db skill, show hourly page view metrics from SurrealDB for 2024-06-01. Execute with agenticdbs query --backend surreal.`)
	lowerOutput := strings.ToLower(output)

	// Verify actual metric values from seed data
	assertContainsAny(t, lowerOutput, []string{"views", "clicks", "150", "200", "180", "metrics"}, "should return actual metric values")
	assertContainsAny(t, lowerOutput, []string{"page_metrics", "2024-06", "hourly", "hour"}, "should query page_metrics with timestamps")
}

// UC1.4: Timeseries aggregation on PostgreSQL
func TestQuery_Timeseries_Postgres(t *testing.T) {
	output := runClaude(t, `Using the /db skill, show hourly aggregated page view metrics from PostgreSQL for 2024-06-01. Execute with agenticdbs query --backend postgres.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, lowerOutput, []string{"views", "clicks", "time_bucket", "sum", "metrics", "aggregate"}, "should use TimescaleDB or aggregation")
	assertContainsAny(t, lowerOutput, []string{"page_metrics", "2024-06", "hourly", "hour"}, "should query page_metrics")
}

// UC1.5: JSON/flexible object query on SurrealDB
func TestQuery_JSON_Surreal(t *testing.T) {
	output := runClaude(t, `Using the /db skill, get user preferences for Alice from SurrealDB, showing her theme and notification settings. Execute with agenticdbs query --backend surreal.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, lowerOutput, []string{"dark", "theme"}, "should return Alice's theme preference (dark)")
	assertContainsAny(t, lowerOutput, []string{"notification", "email", "push"}, "should return notification settings")
}

// UC1.5: JSON/flexible object query on PostgreSQL
func TestQuery_JSON_Postgres(t *testing.T) {
	output := runClaude(t, `Using the /db skill, get user preferences for Alice (user_id=1) from PostgreSQL, showing her theme and notification settings. Execute with agenticdbs query --backend postgres.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, lowerOutput, []string{"dark", "theme"}, "should return Alice's theme preference")
	assertContainsAny(t, lowerOutput, []string{"notification", "email", "push", "preference"}, "should return notification settings")
}

// UC1.6: Full-text search on SurrealDB
func TestQuery_FullText_Surreal(t *testing.T) {
	output := runClaude(t, `Using the /db skill, search for products about 'wireless' in SurrealDB using full-text search. Execute with agenticdbs query --backend surreal.`)
	lowerOutput := strings.ToLower(output)

	// Headphones and mouse both have "wireless" in description
	assertContainsAny(t, lowerOutput, []string{"headphones", "mouse", "wireless"}, "should find wireless products")
	assertContainsAny(t, lowerOutput, []string{"search", "@@", "score", "full-text", "fulltext", "bm25", "text search"}, "should use full-text search")
}

// UC1.6: Full-text search on PostgreSQL
func TestQuery_FullText_Postgres(t *testing.T) {
	output := runClaude(t, `Using the /db skill, search for products about 'wireless' in PostgreSQL using full-text search. Execute with agenticdbs query --backend postgres.`)
	lowerOutput := strings.ToLower(output)

	assertContainsAny(t, lowerOutput, []string{"headphones", "mouse", "wireless"}, "should find wireless products")
	assertContainsAny(t, lowerOutput, []string{"tsvector", "tsquery", "@@", "ts_rank", "to_tsquery", "full-text", "fulltext", "text search"}, "should use PostgreSQL full-text search")
}

// UC1.7: Auto-recommend backend when not specified
func TestQuery_AutoRecommend(t *testing.T) {
	output := runClaude(t, `Using the /db skill, I want to find friends of a user and explore their social connections. Which backend should I use and why? Don't specify a backend - let the system recommend one.`)

	lowerOutput := strings.ToLower(output)
	assertContainsAny(t, lowerOutput, []string{"surrealdb", "surreal"}, "should recommend SurrealDB for graph queries")
	assertContainsAny(t, lowerOutput, []string{"graph", "traversal", "native", "arrow", "relate", "relationship"}, "should explain graph capability as reason")
	// Should also mention PostgreSQL AGE as alternative
	assertContainsAny(t, lowerOutput, []string{"postgres", "age", "cypher", "alternative", "also"}, "should mention PostgreSQL/AGE as alternative")
}

// UC1.8: Debug a failing query
func TestQuery_Debug(t *testing.T) {
	output := runClaude(t, `Using the /db skill, I tried to run this SurrealDB query but it's not working: "SELCT * FORM users". Help me debug and fix it. Execute the corrected query with agenticdbs query --backend surreal.`)

	lowerOutput := strings.ToLower(output)
	assertContainsAny(t, lowerOutput, []string{"select", "from"}, "should suggest corrected SELECT FROM syntax")
	assertContainsAny(t, lowerOutput, []string{"typo", "syntax", "error", "incorrect", "misspell", "spell", "selct", "form"}, "should identify the error type")
	// Should show actual results after fixing
	assertContainsAny(t, output, []string{"Alice", "Bob", "alice", "bob", "users"}, "should return actual user data after correction")
}
