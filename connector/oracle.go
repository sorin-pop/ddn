package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/djavorszky/ddn/common/model"
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

func (db *oracle) CreateDatabase(dbRequest model.DBRequest) error {

	err := db.Alive()
	if err != nil {
		return fmt.Errorf("alive check failed: %s", err.Error())
	}

	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@./sql/oracle/create_schema.sql", dbRequest.Username, dbRequest.Password, conf.Tablespace}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode == 1920 {
		return fmt.Errorf("user/schema %s already exists", dbRequest.Username)
	}

	if res.exitCode != 0 {
		return fmt.Errorf("unable to create database:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return nil
}

func (db *oracle) DropDatabase(dbRequest model.DBRequest) error {

	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@./sql/oracle/drop_schema.sql", dbRequest.Username}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return fmt.Errorf("Unable to drop database:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return nil
}

func (db *oracle) ImportDatabase(dbRequest model.DBRequest) error {
	//var errBuf bytes.Buffer

	/*file, err := os.Open(dbreq.DumpLocation)
	if err != nil {
		db.DropDatabase(dbreq)
		return fmt.Errorf("could not open dumpfile '%s': %s", dbreq.DumpLocation, err.Error())
	}
	defer file.Close()*/

	log.Println("Dump location: ", dbRequest.DumpLocation)

	dumpDir, fileName := filepath.Split(dbRequest.DumpLocation)

	// Start the import
	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@./sql/oracle/import_dump.sql", dumpDir, fileName, dbRequest.Username, dbRequest.Password, conf.Tablespace}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return fmt.Errorf("Dump import seems to have failed:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return nil
}

func (db *oracle) ListDatabase() ([]string, error) {
	return nil, nil
}

func (db *oracle) Version() (string, error) {

	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@./sql/oracle/get_db_version.sql"}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return "", fmt.Errorf("Unable to get Oracle version:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return strings.TrimSpace(res.stdout), nil
}

func (db *oracle) RequiredFields(dbreq model.DBRequest, reqType int) []string {
	req := []string{dbreq.Username}

	switch reqType {
	case createDB:
		req = append(req, dbreq.Password)
	case importDB:
		req = append(req, strconv.Itoa(dbreq.ID), dbreq.Password, dbreq.DumpLocation)
	}

	return req
}
