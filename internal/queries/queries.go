package queries

import (
	"embed"

	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/shurco/litecart/pkg/fsutil"
	_ "modernc.org/sqlite"
)

var db *Base

type Base struct {
	AuthQueries
	InstallQueries
	SettingQueries
}

func InitDB(dbPath string, migrations embed.FS) (*sql.DB, error) {
	if !fsutil.IsFile(dbPath) {
		// create db
		if _, err := fsutil.OpenFile(dbPath, fsutil.FsCWFlags, 0666); err != nil {
			return nil, err
		}

		// first migrate db
		if err := Migrate(dbPath, migrations); err != nil {
			return nil, err
		}
	}

	// connect to database
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)
	sqlite, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return sqlite, nil
}

func Migrate(dbPath string, migrations embed.FS) error {
	goose.SetBaseFS(migrations)
	db, err := goose.OpenDBWithDriver("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	return nil
}

func InitQueries(embed embed.FS) error {
	// init database
	sqlite, err := InitDB("./lc_base/data.db", embed)
	if err != nil {
		return err
	}

	db = &Base{
		AuthQueries:    AuthQueries{DB: sqlite},
		InstallQueries: InstallQueries{DB: sqlite},
		SettingQueries: SettingQueries{DB: sqlite},
	}
	return nil
}

func DB() *Base {
	return db
}