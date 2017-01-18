package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-oci8"
)

type oracle struct {
	conn *sql.DB
}

func (db *oracle) Connect(c Config) error {
	var err error

	if ok := present(c.User, c.Password, c.DBAddress, c.DBPort, c.SID); !ok {
		return fmt.Errorf("Missing parameters. Need-Got: {user: %s}, {password: %s}, {dbAddress: %s}, {dbPort: %s}, {oracle-sid: %s}", c.User, c.Password, c.DBAddress, c.DBPort, c.SID)
	}

	datasource := fmt.Sprintf("%s/%s@%s:%s/%s", c.User, c.Password, c.DBAddress, c.DBPort, c.SID)
	db.conn, err = sql.Open("oci8", datasource)
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

func (db *oracle) Close() {
	db.conn.Close()
}

func (db *oracle) Alive() error {
	return nil
}

func (db *oracle) CreateDatabase(dbRequest DBRequest) error {
	return nil
}

func (db *oracle) DropDatabase(dbRequest DBRequest) error {
	return nil
}

func (db *oracle) ImportDatabase(dbRequest DBRequest) error {
	return nil
}

func (db *oracle) ListDatabase() ([]string, error) {
	return nil, nil
}

func (db *oracle) Version() (string, error) {
	return "", nil
}
