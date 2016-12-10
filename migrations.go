package tsq

import (
	"database/sql"
	"errors"
	"log"
)

type MigrationFn func(*sql.DB) error

type Migration struct {
	version string
	migrate MigrationFn
}

func (m *Migration) hasRun(db *sql.DB) (err error, status bool) {
	err = db.QueryRow("select status from schema_version where version = ?", m.version).Scan(&status)
	switch {
	case err == sql.ErrNoRows:
		return nil, false
	case err != nil:
		return
	}
	if status == false {
		err = errors.New("Migration " + m.version + " failed before. Manual cleanup required")
	}
	return
}

func (m *Migration) setStatus(db *sql.DB, status bool) (err error) {
	_, err = db.Exec("insert into schema_version(version, status) values (?, ?)", m.version, status)
	return
}

type Migrations struct {
	db         *sql.DB
	migrations []Migration
}

func NewMigrations(db *sql.DB) *Migrations {
	return &Migrations{db: db}
}

func (ms *Migrations) Run() (err error) {
	_, err = ms.db.Exec(`
		create table if not exists schema_version (
			version string not null primary key,
			status boolean not null
		)`)
	if err != nil {
		return err
	}
	for _, migration := range ms.migrations {
		m_err, ok := migration.hasRun(ms.db)
		if m_err != nil {
			return m_err
		}
		if !ok {
			log.Println("Running migration: " + migration.version)
			m_err = migration.migrate(ms.db)
			if m_err != nil {
				migration.setStatus(ms.db, false)
				return m_err
			}
			migration.setStatus(ms.db, true)
		}
	}
	return
}

func (ms *Migrations) Register(version string, migration MigrationFn) {
	ms.migrations = append(ms.migrations, Migration{version: version, migrate: migration})
}
