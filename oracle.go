package main

import (
	"database/sql"
)

type oracle struct {
	conn *sql.DB
}

func (db *oracle) Connect(user, password, DBPort string) error {
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
