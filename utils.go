package main

import "database/sql"


func InitDB() {
    var err error
    db, err = sql.Open("sqlite", "./streaming_service.db")
    if err != nil {
        log.Fatal(err)
    }

    createTable := `CREATE TABLE IF NOT EXISTS tokens (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        token TEXT NOT NULL,
        expires_at TIMESTAMP NOT NULL
    );`
    _, err = db.Exec(createTable)
    if err != nil {
        log.Fatal(err)
    }
}