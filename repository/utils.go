package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

const (
	ENVVAR_DBUSER = "MERKNERA_DBUSER"
	ENVVAR_DBPASS = "MERKNERA_DBPASS"
	ENVVAR_DBHOST = "MERKNERA_DBHOST"
	ENVVAR_DBNAME = "MERKNERA_DBNAME"
)

var DB *sql.DB

func init() {
	dbuser := os.Getenv(ENVVAR_DBUSER)
	dbpass := os.Getenv(ENVVAR_DBPASS)
	dbhost := os.Getenv(ENVVAR_DBHOST)
	dbname := os.Getenv(ENVVAR_DBNAME)

	if DB == nil {
		connstring := fmt.Sprintf("postgres://%s:%s@%s/%s", dbuser, dbpass, dbhost, dbname)
		db, err := sql.Open("postgres", connstring)
		if err != nil {
			log.Fatalf("Error opening database connection: %s", err)
		}
		DB = db
	}
}

func GetDB() *sql.DB {
	return DB
}
