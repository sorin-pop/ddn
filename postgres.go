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

func (db *postgres) ListDatabase() ([]string, error) {
	var err error

	err = db.Alive()
	if err != nil {
		log.Println("Died:", err)
		return nil, err
	}

	rows, err := db.conn.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")
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

		switch database {
		case "postgres":
			continue
		}

		list = append(list, database)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return list, nil
}

func (db *postgres) userExists(user string) (bool, error) {
	var count int

	query := fmt.Sprintf("SELECT count(1) FROM pg_roles WHERE rolname='%s'", user)

	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return true, err
	}
	if count == 0 {
		return true, nil
	}

	return false, nil
}

func (db *postgres) dbExists(database string) (bool, error) {
	var count int

	query := fmt.Sprintf("SELECT count(*) FROM pg_database WHERE datistemplate = false AND datname = '%s'", database)

	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return true, err
	}
	if count == 0 {
		return true, nil
	}

	return false, nil
}
