package common

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDb(f string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", f)
	return db, err
}
