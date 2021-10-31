package main

import (
	"database/sql"
	"embed"
	"log"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var migrationsFolder embed.FS

// MigrateUp start the migrations
func MigrateUp(dbDriver string, dbString string) error {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFolder,
		Root:       "migrations",
	}

	db, err := sql.Open(dbDriver, dbString)
	if err != nil {
		return err
	}

	n, err := migrate.Exec(db, dbDriver, migrations, migrate.Up)
	if err != nil {
		return err
	}

	log.Printf("Applied %d migrations!\n", n)
	return nil
}

func MigrateDown(dbDriver string, dbString string) error {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFolder,
		Root:       ".",
	}

	db, err := sql.Open(dbDriver, dbString)
	if err != nil {
		return err
	}

	n, err := migrate.Exec(db, dbDriver, migrations, migrate.Down)
	if err != nil {
		return err
	}

	log.Printf("Down %d migrations!\n", n)
	return nil
}
