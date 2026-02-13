package vel

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/velocitykode/vel/internal/ui"
	"github.com/velocitykode/velocity/pkg/orm"
	"github.com/velocitykode/velocity/pkg/orm/migrate"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run all pending database migrations for your application.`,
	RunE:  runMigrate,
}

var migrateFreshCmd = &cobra.Command{
	Use:   "migrate:fresh",
	Short: "Drop all tables and re-run migrations",
	Long:  `Drop all database tables and re-run all migrations from scratch.`,
	RunE:  runMigrateFresh,
}

var migrateRollbackCmd = &cobra.Command{
	Use:   "migrate:rollback",
	Short: "Rollback the last database migration batch",
	Long:  `Rollback the last batch of database migrations. Use --step to rollback multiple batches.`,
	RunE:  runMigrateRollback,
}

func init() {
	migrateRollbackCmd.Flags().IntP("step", "s", 1, "Number of batches to rollback")
}

// initDB creates an ORM manager from environment variables.
// Returns (nil, nil) if no DB_CONNECTION is configured.
func initDB() (*orm.Manager, error) {
	driver := os.Getenv("DB_CONNECTION")
	if driver == "" {
		return nil, nil
	}
	return orm.NewManager(orm.ManagerConfig{
		Driver:   driver,
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Database: os.Getenv("DB_DATABASE"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	})
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ui.Header("migrate")

	// Initialize database from environment
	manager, err := initDB()
	if err != nil {
		ui.Error(fmt.Sprintf("Database connection failed: %v", err))
		return err
	}

	if manager == nil {
		ui.Warning("No database configured (DB_CONNECTION not set), skipping migrations")
		return nil
	}
	defer manager.Close()

	// Get all registered migrations (via init() imports in user's cmd/velocity/main.go)
	migrations := migrate.All()
	if len(migrations) == 0 {
		ui.Warning("No migrations found")
		return nil
	}

	// Create migrator
	migrator := migrate.NewMigrator(manager.DB(), manager.DriverName())

	// Get pending migrations
	pending, err := getPendingMigrations(manager.DB(), migrations)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get pending migrations: %v", err))
		return err
	}

	if len(pending) == 0 {
		ui.Info("Nothing to migrate")
		return nil
	}

	ui.Info("Running migrations")

	if err := migrator.Up(); err != nil {
		ui.Error(fmt.Sprintf("Migration failed: %v", err))
		return err
	}

	for _, m := range pending {
		ui.Success(fmt.Sprintf("%s_%s", m.Version, m.Description))
	}

	ui.Newline()
	ui.Success("Done")
	return nil
}

func runMigrateFresh(cmd *cobra.Command, args []string) error {
	ui.Header("migrate:fresh")

	// Initialize database from environment
	manager, err := initDB()
	if err != nil {
		ui.Error(fmt.Sprintf("Database connection failed: %v", err))
		return err
	}

	if manager == nil {
		ui.Warning("No database configured (DB_CONNECTION not set), skipping migrations")
		return nil
	}
	defer manager.Close()

	// Get all registered migrations
	migrations := migrate.All()
	if len(migrations) == 0 {
		ui.Warning("No migrations found")
		return nil
	}

	// Create migrator
	migrator := migrate.NewMigrator(manager.DB(), manager.DriverName())

	ui.Info("Dropping all tables")

	if err := migrator.Fresh(); err != nil {
		ui.Error(fmt.Sprintf("Fresh migration failed: %v", err))
		return err
	}

	ui.Info("Running migrations")

	for _, m := range migrations {
		ui.Success(fmt.Sprintf("%s_%s", m.Version, m.Description))
	}

	ui.Newline()
	ui.Success("Done")
	return nil
}

func runMigrateRollback(cmd *cobra.Command, args []string) error {
	ui.Header("migrate:rollback")

	// Initialize database from environment
	manager, err := initDB()
	if err != nil {
		ui.Error(fmt.Sprintf("Database connection failed: %v", err))
		return err
	}

	if manager == nil {
		ui.Warning("No database configured (DB_CONNECTION not set), skipping rollback")
		return nil
	}
	defer manager.Close()

	// Get all registered migrations
	migrations := migrate.All()
	if len(migrations) == 0 {
		ui.Warning("No migrations found")
		return nil
	}

	// Create migrator
	migrator := migrate.NewMigrator(manager.DB(), manager.DriverName())

	steps, _ := cmd.Flags().GetInt("step")

	// Get migrations that will be rolled back (for display)
	rollbackVersions, err := getRollbackMigrations(manager.DB(), steps)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get rollback migrations: %v", err))
		return err
	}

	if len(rollbackVersions) == 0 {
		ui.Info("Nothing to rollback")
		return nil
	}

	ui.Info("Rolling back migrations")

	if err := migrator.Down(steps); err != nil {
		ui.Error(fmt.Sprintf("Rollback failed: %v", err))
		return err
	}

	for _, version := range rollbackVersions {
		ui.Success(version)
	}

	ui.Newline()
	ui.Success("Done")
	return nil
}

func getRollbackMigrations(db *sql.DB, steps int) ([]string, error) {
	// Query all migrations with batch info (no parameters, works on all drivers)
	rows, err := db.Query("SELECT version, batch FROM migrations ORDER BY version DESC")
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	// Find the max batch and collect version/batch pairs
	type migrationRecord struct {
		version string
		batch   int
	}
	var records []migrationRecord
	maxBatch := 0
	for rows.Next() {
		var r migrationRecord
		if err := rows.Scan(&r.version, &r.batch); err != nil {
			continue
		}
		records = append(records, r)
		if r.batch > maxBatch {
			maxBatch = r.batch
		}
	}

	if maxBatch == 0 {
		return nil, nil
	}

	// Filter to versions in the last N batches
	cutoff := maxBatch - steps
	var versions []string
	for _, r := range records {
		if r.batch > cutoff {
			versions = append(versions, r.version)
		}
	}

	return versions, nil
}

func getPendingMigrations(db *sql.DB, all []migrate.Migration) ([]migrate.Migration, error) {
	// Get applied migrations from database
	appliedVersions := make(map[string]bool)

	rows, err := db.Query("SELECT version FROM migrations")
	if err != nil {
		// Table might not exist yet, all migrations are pending
		return all, nil
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			continue
		}
		appliedVersions[version] = true
	}

	// Find pending
	var pending []migrate.Migration
	for _, m := range all {
		if !appliedVersions[m.Version] {
			pending = append(pending, m)
		}
	}

	return pending, nil
}
