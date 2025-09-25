package db

import (
	"capitalWhCalendar/config"
	"database/sql"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

var DB *sql.DB

func init() {
	var err error

	server := config.Config["driverName"].(string)
	conn := config.Config["connectionStr"].(string)

	DB, err = sql.Open(server, conn)
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}
}
