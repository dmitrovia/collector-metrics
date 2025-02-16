// Package migrator
// implements database migrations.
package migrator

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrator - describing the struct migrator.
type Migrator struct {
	srcDriver source.Driver
}

// MustGetNewMigrator - to create an instance
// of a migrator object.
func MustGetNewMigrator(
	sqlFiles embed.FS,
	dirName string,
) (*Migrator, error) {
	driver, err := iofs.New(sqlFiles, dirName)
	if err != nil {
		return nil,
			fmt.Errorf("MustGetNewMigrator->iofs.New %w", err)
	}

	return &Migrator{
		srcDriver: driver,
	}, nil
}

// ApplyMigrations - applies migrations to the database.
func (m *Migrator) ApplyMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(
		db,
		&postgres.Config{})
	if err != nil {
		return fmt.Errorf("unable to create db instance: %w", err)
	}

	migrator, err := migrate.NewWithInstance(
		"migration_embedded_sql_files",
		m.srcDriver,
		"psql_db",
		driver)
	if err != nil {
		return fmt.Errorf("unable to create migration: %w", err)
	}

	defer migrator.Close()

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("unable to apply migrations %w", err)
	}

	return nil
}
