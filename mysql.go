package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
)

type mysql struct {
	conn *sql.DB
}

// Connect creates and initialises a Database struct and connects to the database
func (db *mysql) Connect(user, password, dbPort string) error {
	var err error

	datasource := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/", user, password, dbPort)
	db.conn, err = sql.Open("mysql", datasource)
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

// Close closes the connection to the database
func (db *mysql) Close() {
	db.conn.Close()
}

// Alive checks whether the connection is alive. Returns error if not.
func (db *mysql) Alive() error {
	defer func() {
		if p := recover(); p != nil {
			log.Println("Panic Attack! Database seems to be down.")
		}
	}()

	_, err := db.conn.Exec("select * from mysql.user WHERE 1 = 0")
	if err != nil {
		return err
	}

	return nil
}

// ListDatabase returns a list of strings - the names of the databases in the server
// All system tables are omitted from the returned list. If there's an error, it is returned.
func (db *mysql) ListDatabase() ([]string, error) {
	var err error

	err = db.Alive()
	if err != nil {
		return nil, err
	}

	rows, err := db.conn.Query("show databases")
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
		case "information_schema", "performance_schema", "mysql", "nbinfo", "sys":
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

// CreateDatabase creates a Database along with a user, to which all privileges
// are granted on the created database. Fails if database or user already exists.
func (db *mysql) CreateDatabase(dbRequest DBRequest) error {

	err := db.Alive()
	if err != nil {
		log.Println("Died:", err)
		return fmt.Errorf("Unable to complete request as the underlying database is down")
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

	_, err = db.conn.Exec(fmt.Sprintf("CREATE DATABASE %s CHARSET utf8;", dbRequest.DatabaseName))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("CREATE USER '%s' IDENTIFIED BY '%s';", dbRequest.Username, dbRequest.Password))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%s';", dbRequest.DatabaseName, dbRequest.Username, "%"))
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

// DropDatabase drops a database and a user. Always succeeds, even if droppable database or
// user does not exist
func (db *mysql) DropDatabase(dbRequest DBRequest) error {
	err := db.Alive()
	if err != nil {
		log.Println("Died:", err)
		return fmt.Errorf("Unable to complete request as the underlying database is down")
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

	exists, err := db.userExists(dbRequest.Username)
	if err != nil {
		tx.Rollback()
		return err
	}

	if exists {
		_, err = db.conn.Exec(fmt.Sprintf("DROP USER %s", dbRequest.Username))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// ImportDatabase imports the dumpfile to the database or returns an error
// if it failed for some reason.
func (db *mysql) ImportDatabase(dbreq DBRequest) error {
	userArg, pwArg, dbnameArg := fmt.Sprintf("-u%s", conf.User), fmt.Sprintf("-p%s", conf.Password), dbreq.DatabaseName

	cmd := exec.Command(conf.Exec, userArg, pwArg, dbnameArg)

	file, err := os.Open(dbreq.DumpLocation)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd.Stdin = file

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (db *mysql) dbExists(databasename string) (bool, error) {
	var count int

	err := db.conn.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", databasename).Scan(&count)
	if err != nil {
		return true, err
	}
	if count != 0 {
		return true, nil
	}

	return false, nil
}

func (db *mysql) userExists(username string) (bool, error) {
	var count int

	err := db.conn.QueryRow("SELECT count(*) FROM mysql.user WHERE user = ?", username).Scan(&count)
	if err != nil {
		return true, err
	}
	if count != 0 {
		return true, nil
	}

	return false, nil
}
