package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var (
	conn string
	db   *sql.DB
	err  error
)

func prepDatabase() {
	conn = fmt.Sprintf("%s:%s@/", conf.User, conf.Password)
	db, err = sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	validateConn()
	validateMetaDB()
}

func validateConn() {

	err := db.Ping()
	if err != nil {
		log.Fatalf("Could not validate database connection:\n\t%s", err.Error())
	}

	log.Println("MySQL connection validated")
}

func validateMetaDB() {
	var count int

	err := db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", "ddnc_info").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 1 {
		return
	}

}
