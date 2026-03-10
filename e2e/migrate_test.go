package e2e

import (
	"strings"
	"testing"
)

// =============================================================
// Goal 2: Migrate Tests
// UC2.1–UC2.4: Schema export, data migration, comparison, env config
// =============================================================

// UC2.1: Export schema from SurrealDB, generate equivalent PostgreSQL schema
func TestMigrate_SchemaExport(t *testing.T) {
	output := runClaude(t, `Using the /db skill, export the users table schema from SurrealDB and generate the equivalent PostgreSQL CREATE TABLE statement. First inspect the SurrealDB schema using agenticdbs query --backend surreal "INFO FOR TABLE users", then produce the PostgreSQL equivalent.`)

	lowerOutput := strings.ToLower(output)

	// Should produce valid PostgreSQL CREATE TABLE
	assertContainsAny(t, lowerOutput, []string{"create table", "postgresql", "postgres"}, "should generate PostgreSQL schema")
	assertContainsAny(t, lowerOutput, []string{"name", "email"}, "should include column definitions")

	// Should reference both backends in the translation
	assertContainsAny(t, lowerOutput, []string{"surrealdb", "surreal"}, "should reference source backend")
}

// UC2.2: Migrate user data from SurrealDB to PostgreSQL
func TestMigrate_DataExport(t *testing.T) {
	output := runClaude(t, `Using the /db skill, migrate user data: first query all users from SurrealDB using agenticdbs query --backend surreal, then show me how to insert that data into PostgreSQL. Actually execute both the export query from SurrealDB and verify the data exists in PostgreSQL with agenticdbs query --backend postgres "SELECT * FROM users".`)

	lowerOutput := strings.ToLower(output)

	// Should show actual user data
	assertContainsAny(t, output, []string{"Alice", "Bob", "Carol", "alice", "bob", "carol"}, "should show exported user data")

	// Should reference both backends
	assertContainsAny(t, lowerOutput, []string{"surrealdb", "surreal"}, "should reference SurrealDB for export")
	assertContainsAny(t, lowerOutput, []string{"postgresql", "postgres", "insert"}, "should reference PostgreSQL insertion")
}

// UC2.3: Compare data between SurrealDB and PostgreSQL
func TestMigrate_DataCompare(t *testing.T) {
	output := runClaude(t, `Using the /db skill, compare the user data between SurrealDB and PostgreSQL to verify they have matching records. Query both backends using agenticdbs query and report whether the data matches.`)

	lowerOutput := strings.ToLower(output)

	// Should show comparison results
	assertContainsAny(t, lowerOutput, []string{"match", "same", "identical", "consistent", "both", "5 user", "5 record", "comparing"}, "should report comparison results")

	// Should include actual data evidence
	assertContainsAny(t, output, []string{"Alice", "Bob", "alice", "bob"}, "should show actual user names from both backends")
}

// UC2.4: Generate env-specific config for staging environment
func TestMigrate_EnvConfig(t *testing.T) {
	output := runClaude(t, `Using the /db skill, generate connection configuration for the PostgreSQL backend for a staging environment at host db.staging.example.com, port 5432, database agenticdbs_staging, user staging_user, password staging_pass_123.`)

	// Should include the specified staging host
	assertContains(t, output, "db.staging.example.com", "should include staging hostname")

	// Should include the specified database name
	assertContainsAny(t, output, []string{"agenticdbs_staging"}, "should include staging database name")

	// Should include the specified credentials
	assertContainsAny(t, output, []string{"staging_user"}, "should include staging username")
}
