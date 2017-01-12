package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type database struct {
	datasource string
	conn       *sql.DB
}

// Connect creates and initialises a Database struct
func (db *database) Connect(server, user, password, DBPort string) error {
	var err error

	db.datasource = fmt.Sprintf("%s:%s@/", user, password)
	db.conn, err = sql.Open(server, db.datasource)
	if err != nil {
		log.Fatal(err)
	}

	err = db.conn.Ping()
	if err != nil {
		db.conn.Close()
		return err
	}

	return nil
}

func (db *database) Close() {
	db.conn.Close()
}

func (db *database) Ping() error {
	return db.conn.Ping()
}

func (db *database) listDatabase() []string {
	var err error

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.conn.Query("show databases")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	list := make([]string, 0, 10)

	var database string
	for rows.Next() {
		err = rows.Scan(&database)
		if err != nil {
			log.Fatal(err)
		}

		if database == "information_schema" || database == "mysql" ||
			database == "performance_schema" || database == "nbinfo" {
			continue
		}

		list = append(list, database)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return list
}
