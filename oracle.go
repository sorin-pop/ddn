package main

import (
	"database/sql"
	"fmt"
	"log"
	//"os/exec"
	//"bytes"
	//_ "github.com/mattn/go-oci8"

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


	args := []string{"-L", "-S", conf.User + "/" + conf.Password, "@create_schema.sql", dbRequest.Username, dbRequest.Password, conf.DefaultTablespace}
	
	stdout, stderr, exitCode := RunCommand(conf.Exec, args...)
	
	if exitCode == 1920 {
		return fmt.Errorf("User/schema " + dbRequest.Username + " already exists!")
	}
	
	if exitCode != 0 {
		return fmt.Errorf(stdout + " " + stderr)
	}
	
	return nil
}

func (db *oracle) DropDatabase(dbRequest DBRequest) error {

	args := []string{"-L", "-S", conf.User + "/" + conf.Password, "@drop_schema.sql", dbRequest.Username}

	stdout, stderr, exitCode := RunCommand(conf.Exec, args...)
	
	if exitCode != 0 {
		return fmt.Errorf(stdout + " " + stderr)
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
	return "", nil
}
