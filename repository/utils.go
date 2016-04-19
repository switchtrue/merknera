package repository

import (
	"database/sql"
	"log"
)

func GetDatabaseConnection() *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost/postgres")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
