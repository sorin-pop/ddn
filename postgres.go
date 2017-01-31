package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	"strings"

	"regexp"

	_ "github.com/lib/pq"
)

type postgres struct {
	conn *sql.DB
}

func (db *postgres) Connect(c Config) error {
	var err error

	if ok := present(c.User, c.Password, c.DBAddress, c.DBPort); !ok {
		return fmt.Errorf("Missing parameters. Need-Got: {user: %s}, {password: %s}, {dbAddress: %s}, {dbPort: %s}", c.User, c.Password, c.DBAddress, c.DBPort)
	}

	datasource := fmt.Sprintf("postgres://%s:%s@%s:%s", c.User, c.Password, c.DBAddress, c.DBPort)
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

func (db *postgres) CreateDatabase(dbRequest DBRequest) error {
	err := db.Alive()
	if err != nil {
		return err
	}

	exists, err := db.dbExists(dbRequest.DatabaseName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Database '%s' already exists", dbRequest.DatabaseName)
	}

	exists, err = db.userExists(dbRequest.Username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("User '%s' already exists", dbRequest.Username)
	}

	// Begin transaction so that we can roll it back at any point something goes wrong.
	tx, err := db.conn.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("CREATE DATABASE %s ENCODING 'utf-8';", dbRequest.DatabaseName))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", dbRequest.Username, dbRequest.Password))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", dbRequest.DatabaseName, dbRequest.Username))
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
func (db *postgres) DropDatabase(dbRequest DBRequest) error {
	var err error

	err = db.Alive()
	if err != nil {
		return err
	}

	tx, err := db.conn.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbRequest.DatabaseName))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("DROP USER IF EXISTS %s", dbRequest.Username))
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (db *postgres) ImportDatabase(dbreq DBRequest) error {
	userArg := fmt.Sprintf("-U%s", conf.User)

	cmd := exec.Command(conf.Exec, userArg, dbreq.DatabaseName)

	file, err := os.Open(dbreq.DumpLocation)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd.Stdin = file

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(errBuf.String())
	}

	return nil
}

func (db *postgres) ListDatabase() ([]string, error) {
	var err error

	err = db.Alive()
	if err != nil {
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

func (db *postgres) Version() (string, error) {
	var buf bytes.Buffer

	cmd := exec.Command(conf.Exec, "--version")
	cmd.Stdout = &buf

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("[0-9.]+")

	return re.FindString(buf.String()), nil
}

func (db *postgres) RequiredFields(dbreq DBRequest, reqType int) []string {
	req := []string{dbreq.DatabaseName, dbreq.Username}

	switch reqType {
	case createDB:
		req = append(req, dbreq.Password)
	case importDB:
		req = append(req, dbreq.Password, dbreq.DumpLocation)
	}

	return req
}

func (db *postgres) userExists(user string) (bool, error) {
	var count int

	query := fmt.Sprintf("SELECT count(1) FROM pg_roles WHERE rolname='%s'", strings.ToLower(user))

	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return true, err
	}
	if count != 0 {
		return true, nil
	}

	return false, nil
}

func (db *postgres) dbExists(database string) (bool, error) {
	var count int

	query := fmt.Sprintf("SELECT count(*) FROM pg_database WHERE datistemplate = false AND datname = '%s'", strings.ToLower(database))

	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return true, err
	}
	if count != 0 {
		return true, nil
	}

	return false, nil
}
