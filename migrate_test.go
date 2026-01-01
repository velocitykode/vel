package vel

import (
	"testing"
)

func TestRunMigrate_FailsWithoutDatabase(t *testing.T) {
	// runMigrate requires database connection
	// Without proper env setup, it should return error
	err := runMigrate(nil, nil)
	if err == nil {
		t.Error("runMigrate() should error when database not configured")
	}
}

func TestRunMigrateFresh_FailsWithoutDatabase(t *testing.T) {
	// runMigrateFresh requires database connection
	err := runMigrateFresh(nil, nil)
	if err == nil {
		t.Error("runMigrateFresh() should error when database not configured")
	}
}

// Note: Full integration tests for migrate require a real database
// Those should be in a separate integration test file with build tag
