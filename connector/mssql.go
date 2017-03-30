package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/djavorszky/ddn/common/model"
)

type mssql struct {
	conn *sql.DB
}

func (db *mssql) Connect(c Config) error {
	return nil
}

func (db *mssql) Close() {
	db.conn.Close()
}

func (db *mssql) Alive() error {
	return nil
}

func (db *mssql) CreateDatabase(dbRequest model.DBRequest) error {

	args := []string{"-b", "-U", conf.User, "-P", conf.Password, "-Q", fmt.Sprintf("CREATE DATABASE %s", dbRequest.DatabaseName)}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		if strings.Contains(res.stderr, "already exists") {
			return fmt.Errorf("Database %s already exists", dbRequest.Username)
		}

		return fmt.Errorf("Unable to create database:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return nil
}

func (db *mssql) DropDatabase(dbRequest model.DBRequest) error {

	args := []string{"-b", " -U", conf.User, "-P", conf.Password, "-Q", fmt.Sprintf("DROP DATABASE %s", dbRequest.DatabaseName)}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		if !(strings.Contains(res.stderr, "it does not exist")) {
			return fmt.Errorf("Unable to drop database:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
		}
	}

	return nil
}

func (db *mssql) ImportDatabase(dbRequest model.DBRequest) error {

	dumpDir, fileName := filepath.Split(dbRequest.DumpLocation)

	// Start the import
	args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@./sql/oracle/import_dump.sql", dumpDir, fileName, dbRequest.Username, dbRequest.Password, conf.Tablespace}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return fmt.Errorf("Dump import seems to have failed:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return nil
}

func (db *mssql) ListDatabase() ([]string, error) {
	return nil, nil
}

func (db *mssql) Version() (string, error) {

	/*args := []string{"-L", "-S", fmt.Sprintf("%s/%s", conf.User, conf.Password), "@./sql/oracle/get_db_version.sql"}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return "", fmt.Errorf("Unable to get Oracle version:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return strings.TrimSpace(res.stdout), nil*/
	return "SQL Server 2012", nil
}

func (db *mssql) RequiredFields(dbreq model.DBRequest, reqType int) []string {
	req := []string{dbreq.DatabaseName}

	switch reqType {
	case createDB:
		req = append(req, dbreq.Password)
	case importDB:
		req = append(req, strconv.Itoa(dbreq.ID), dbreq.Password, dbreq.DumpLocation)
	}

	return req
}
