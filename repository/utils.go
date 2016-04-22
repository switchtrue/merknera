package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

const (
	ENVVAR_DBUSER = "MERKNERA_DBUSER"
	ENVVAR_DBPASS = "MERKNERA_DBPASS"
	ENVVAR_DBHOST = "MERKNERA_DBHOST"
	ENVVAR_DBNAME = "MERKNERA_DBNAME"
)

var dbuser = os.Getenv(ENVVAR_DBUSER)
var dbpass = os.Getenv(ENVVAR_DBPASS)
var dbhost = os.Getenv(ENVVAR_DBHOST)
var dbname = os.Getenv(ENVVAR_DBNAME)

var DB *sql.DB

func InitializeDatabasePool() *sql.DB {
	connstring := fmt.Sprintf("postgres://%s:%s@%s/%s", dbuser, dbpass, dbhost, dbname)
	db, err := sql.Open("postgres", connstring)
	if err != nil {
		log.Fatal(err)
	}

	DB = db

	return DB
}

func NewTransaction() *sql.DB {
	//db, err := DB.Begin()
	//if err != nil {
	//	log.Println("error beginning transaction")
	//	log.Fatal(err)
	//}

	return DB
}

func GetDB() *sql.DB {
	return DB
}
