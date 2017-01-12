package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type database struct {
	conn string
	db   *sql.DB
}

// Connect creates and initialises a Database struct
func (d *database) Connect(server, user, password, DBPort string) error {
	var err error

	d.conn = fmt.Sprintf("%s:%s@/", user, password)
	d.db, err = sql.Open(server, d.conn)
	if err != nil {
		log.Fatal(err)
	}

	err = d.db.Ping()
	if err != nil {
		d.db.Close()
		return err
	}

	return nil
}

func (d *database) Close() {
	d.db.Close()
}

func (d *database) listDatabase() []string {
	var err error

	rows, err := d.db.Query("show databases")
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

		if database == "information_schema" || database == "mysql" || database == "performance_schema" {
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
