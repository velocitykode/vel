//go:build integration

// Package testutil provides reusable test utilities for the vel CLI test suite.
package testutil

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestDB represents a test database connection with cleanup functionality.
type TestDB struct {
	DB      *sql.DB
	Path    string
	Cleanup func()
}

// SetupTestDB creates a temporary SQLite database for testing.
// It returns a TestDB with the database connection, path, and cleanup function.
// The cleanup function removes the temporary database file and closes the connection.
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Create temp directory for database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("SetupTestDB: failed to open database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("SetupTestDB: failed to ping database: %v", err)
	}

	// Set environment variables for orm.InitFromEnv()
	os.Setenv("DB_CONNECTION", "sqlite")
	os.Setenv("DB_DATABASE", dbPath)

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
		os.Unsetenv("DB_CONNECTION")
		os.Unsetenv("DB_DATABASE")
	}

	return &TestDB{
		DB:      db,
		Path:    dbPath,
		Cleanup: cleanup,
	}
}

// CreateMigrationsTable creates the migrations table in the test database.
// This simulates the table that the migrator creates to track applied migrations.
func (tdb *TestDB) CreateMigrationsTable(t *testing.T) {
	t.Helper()

	_, err := tdb.DB.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL,
			batch INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("CreateMigrationsTable: failed to create table: %v", err)
	}
}

// MarkMigrationApplied records a migration as applied in the migrations table.
// This is useful for setting up test scenarios where some migrations are already applied.
func (tdb *TestDB) MarkMigrationApplied(t *testing.T, version string, batch int) {
	t.Helper()

	_, err := tdb.DB.Exec(
		"INSERT INTO migrations (version, batch, created_at) VALUES (?, ?, ?)",
		version, batch, time.Now(),
	)
	if err != nil {
		t.Fatalf("MarkMigrationApplied: failed to insert migration %s: %v", version, err)
	}
}

// GetAppliedMigrations returns a list of version strings for all applied migrations.
func (tdb *TestDB) GetAppliedMigrations(t *testing.T) []string {
	t.Helper()

	rows, err := tdb.DB.Query("SELECT version FROM migrations ORDER BY id")
	if err != nil {
		t.Fatalf("GetAppliedMigrations: failed to query migrations: %v", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			t.Fatalf("GetAppliedMigrations: failed to scan row: %v", err)
		}
		versions = append(versions, version)
	}

	return versions
}

// TableExists checks if a table exists in the test database.
func (tdb *TestDB) TableExists(t *testing.T, tableName string) bool {
	t.Helper()

	var count int
	err := tdb.DB.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
		tableName,
	).Scan(&count)
	if err != nil {
		t.Fatalf("TableExists: failed to check table %s: %v", tableName, err)
	}

	return count > 0
}

// ExecSQL executes a SQL statement on the test database.
// This is useful for setting up test fixtures or verifying database state.
func (tdb *TestDB) ExecSQL(t *testing.T, query string, args ...interface{}) sql.Result {
	t.Helper()

	result, err := tdb.DB.Exec(query, args...)
	if err != nil {
		t.Fatalf("ExecSQL: failed to execute query: %v", err)
	}

	return result
}

// QueryRow executes a query that returns a single row.
func (tdb *TestDB) QueryRow(t *testing.T, query string, args ...interface{}) *sql.Row {
	t.Helper()

	return tdb.DB.QueryRow(query, args...)
}
