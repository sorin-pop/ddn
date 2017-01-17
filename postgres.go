package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type postgres struct {
	conn *sql.DB
}

func (db *postgres) Connect(user, password, DBPort string) error {
	var err error

	datasource := fmt.Sprintf("postgres://%s:%s@127.0.0.1:%s", user, password, DBPort)
	db.conn, err = sql.Open("postgres", datasource)
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

func (db *postgres) Close() {
	db.conn.Close()
}

func (db *postgres) Alive() error {
	defer func() {
		if p := recover(); p != nil {
			log.Println("Panic Attack! Database seems to be down.")
		}
	}()

	_, err := db.conn.Exec("select 1 from pg_roles WHERE 1 = 0")
	if err != nil {
		return err
	}

	return nil
}

func (db *postgres) CreateDatabase(dbRequest DBRequest) error { return nil }
func (db *postgres) DropDatabase(dbRequest DBRequest) error   { return nil }
func (db *postgres) ImportDatabase(dbRequest DBRequest) error { return nil }
func (db *postgres) ListDatabase() ([]string, error)          { return nil, nil }

func (db *postgres) dbExists(database string) error {
	return nil
}
