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
func (d *database) Connect(server, user, password, DBPort string) error {
	var err error

	d.datasource = fmt.Sprintf("%s:%s@/", user, password)
	d.conn, err = sql.Open(server, d.datasource)
	if err != nil {
		log.Fatal(err)
	}

	err = d.conn.Ping()
	if err != nil {
		d.conn.Close()
		return err
	}

	return nil
}

func (d *database) Close() {
	d.conn.Close()
}

func (d *database) Ping() error {
	return d.conn.Ping()
}

func (d *database) listDatabase() []string {
	var err error

	err = d.Ping()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := d.conn.Query("show databases")
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
