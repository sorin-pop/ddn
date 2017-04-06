package main

import (
	"database/sql"
	"fmt"
	"os"
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

	args := []string{"-b", "-U", conf.User, "-P", conf.Password, "-v", fmt.Sprintf("DROP DATABASE %s", dbRequest.DatabaseName)}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		if !(strings.Contains(res.stderr, "it does not exist")) {
			return fmt.Errorf("Unable to drop database:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
		}
	}

	return nil
}

func (db *mssql) ImportDatabase(dbRequest model.DBRequest) error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not determine current exe directory.")
	}

	args := []string{"-b",
		"-U", conf.User,
		"-P", conf.Password,
		"-v", fmt.Sprintf("dumpFile=\"%s\"", dbRequest.DumpLocation), "targetDatabaseName=" + "\"" + dbRequest.DatabaseName + "\"",
		"-i", "\"" + curDir + "sql\\mssql\\import_dump.sql" + "\""}

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

	args := []string{"-b", "-h", "-1", "-W", "-U", conf.User, "-P", conf.Password, "-Q",
		"SET NOCOUNT ON; SELECT (CAST(SERVERPROPERTY('productversion') AS nvarchar(128)) + SPACE(1) + CAST(SERVERPROPERTY('productlevel') AS nvarchar(128)) + SPACE(1) + CAST(SERVERPROPERTY('edition') AS nvarchar(128)))"}

	res := RunCommand(conf.Exec, args...)

	if res.exitCode != 0 {
		return "", fmt.Errorf("Unable to get SQL Server version:\n> stdout:\n'%s'\n> stderr:\n'%s'\n> exitCode: %d", res.stdout, res.stderr, res.exitCode)
	}

	return strings.TrimSpace(res.stdout), nil
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
