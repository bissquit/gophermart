package migrations

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

// 1) Go will pack sql files into binary during compilation
// 'go:embed *.sql' - is a template which files should be packed
//
//go:embed *.sql
var embeddedMigrations embed.FS

func InitializeDB(databaseURL string) error {
	// 2) iofs.New says to golang-migrate:
	// you should look for a migrations in that filesystem (embeddedMigrations)
	// by the following path - "."
	d, err := iofs.New(embeddedMigrations, ".")
	if err != nil {
		return err
	}

	// 3) databaseURL — DSN to connect to DB,
	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	// 4) Up() applies one by one all migration that haven't been applied yet
	// if all migrations are already applied — return migrate.ErrNoChange (not an error)
	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}
