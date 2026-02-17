package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version   string
	Name      string
	SQL       string
	AppliedAt *time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// Run executes all pending migrations
func (m *Migrator) Run() error {
	log.Printf("[Migrator] Starting migrations...")

	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get all migration files
	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Apply pending migrations
	pendingCount := 0
	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			continue // Already applied
		}

		log.Printf("[Migrator] Applying migration: %s - %s", migration.Version, migration.Name)

		if err := m.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		pendingCount++
		log.Printf("[Migrator] Applied migration: %s", migration.Version)
	}

	if pendingCount == 0 {
		log.Printf("[Migrator] No pending migrations")
	} else {
		log.Printf("[Migrator] Applied %d migrations", pendingCount)
	}

	return nil
}

// ensureMigrationsTable creates the migrations tracking table if it doesn't exist
func (m *Migrator) ensureMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := m.db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of already applied migrations
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	query := `SELECT version FROM schema_migrations`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// loadMigrations loads all migration files from the embedded filesystem
func (m *Migrator) loadMigrations() ([]*Migration, error) {
	var migrations []*Migration

	err := fs.WalkDir(migrationsFS, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		filename := filepath.Base(path)
		parts := strings.SplitN(strings.TrimSuffix(filename, ".sql"), "_", 2)
		
		version := parts[0]
		name := filename
		if len(parts) > 1 {
			name = parts[1]
		}

		migrations = append(migrations, &Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(migration *Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Apply migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	query := `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`
	if _, err := tx.Exec(query, migration.Version, migration.Name); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// Status returns the status of all migrations
func (m *Migrator) Status() ([]*Migration, error) {
	if err := m.ensureMigrationsTable(); err != nil {
		return nil, err
	}

	// Get applied migrations with timestamps
	query := `SELECT version, name, applied_at FROM schema_migrations ORDER BY version`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedMap := make(map[string]*Migration)
	for rows.Next() {
		var version, name string
		var appliedAt time.Time
		if err := rows.Scan(&version, &name, &appliedAt); err != nil {
			return nil, err
		}
		appliedMap[version] = &Migration{
			Version:   version,
			Name:      name,
			AppliedAt: &appliedAt,
		}
	}

	// Get all migrations
	migrations, err := m.loadMigrations()
	if err != nil {
		return nil, err
	}

	// Merge with applied info
	for _, migration := range migrations {
		if applied, ok := appliedMap[migration.Version]; ok {
			migration.AppliedAt = applied.AppliedAt
		}
	}

	return migrations, nil
}

// Rollback rolls back the last migration (not implemented yet)
func (m *Migrator) Rollback() error {
	return fmt.Errorf("rollback not implemented - please create a new migration to undo changes")
}
