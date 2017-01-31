package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type oracle struct {
	conn *sql.DB
}

func (db *oracle) Connect(c Config) error {
	return nil
}

func (db *oracle) Close() {
	db.conn.Close()
}

func (db *oracle) Alive() error {
	return nil
}

func (db *oracle) CreateDatabase(dbRequest DBRequest) error {

	err := db.Alive()
	if err != nil {
		log.Println("Died:", err)
		return fmt.Errorf("Unable to complete request as the underlying database is down")
	}

	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@create_schema.sql", dbRequest.Username, dbRequest.Password, conf.DefaultTablespace}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode == 1920 {
		return fmt.Errorf("User/schema %s already exists!", dbRequest.Username)
	}

	if res.exitCode != 0 {
		return fmt.Errorf(res.stdout, res.stderr)
	}

	return nil
}

func (db *oracle) DropDatabase(dbRequest DBRequest) error {

	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@drop_schema.sql", dbRequest.Username}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return fmt.Errorf(res.stdout, res.stderr)
	}

	return nil
}

func (db *oracle) ImportDatabase(dbRequest DBRequest) error {
	return nil
}

func (db *oracle) ListDatabase() ([]string, error) {
	return nil, nil
}

func (db *oracle) Version() (string, error) {

	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@get_db_version.sql"}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return "", fmt.Errorf(res.stdout, res.stderr)
	}

	return strings.TrimSpace(res.stdout), nil
}
