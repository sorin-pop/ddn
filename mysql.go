package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func validateConn() {

	conn := fmt.Sprintf("%s:%s@/", conf.User, conf.Password)

	db, err := sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not validate database connection:\n\t%s", err.Error())
	}

	log.Println("MySQL connection validated")
}
