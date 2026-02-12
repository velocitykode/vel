package vel

import (
	"testing"
)

func TestRunMigrate_FailsWithoutDatabase(t *testing.T) {
	// runMigrate without DB_CONNECTION set should return nil (graceful skip)
	err := runMigrate(nil, nil)
	if err != nil {
		t.Errorf("runMigrate() should return nil when database not configured, got: %v", err)
	}
}

func TestRunMigrateFresh_FailsWithoutDatabase(t *testing.T) {
	// runMigrateFresh without DB_CONNECTION set should return nil (graceful skip)
	err := runMigrateFresh(nil, nil)
	if err != nil {
		t.Errorf("runMigrateFresh() should return nil when database not configured, got: %v", err)
	}
}

func TestRunMigrateRollback_FailsWithoutDatabase(t *testing.T) {
	// runMigrateRollback without DB_CONNECTION set should return nil (graceful skip)
	err := runMigrateRollback(migrateRollbackCmd, nil)
	if err != nil {
		t.Errorf("runMigrateRollback() should return nil when database not configured, got: %v", err)
	}
}

// Note: Full integration tests for migrate require a real database
// Those should be in a separate integration test file with build tag
