package db

import (
	"database/sql"
	"os"
)


var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbName = os.Getenv("DB_NAME")
var DB *sql.DB

func InitDB() (Db *sql.DB, err error) {
    fdbFile := dbName + "?_auth&_auth_user=" + dbUser + "&_auth_pass=" + dbPassword + "&_auth_crypt=sha1"
	Db, err = sql.Open("sqlite3", fdbFile)
	if err != nil {
		return nil, err
	}

	createTable := `CREATE TABLE IF NOT EXISTS tokens (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        token TEXT NOT NULL,
        expires_at TIMESTAMP NOT NULL
    );`
	_, err = Db.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return Db, nil
}