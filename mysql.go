package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type database struct {
	datasource string
	conn       *sql.DB
}

// Connect creates and initialises a Database struct
func (db *database) Connect(server, user, password, DBPort string) error {
	var err error

	db.datasource = fmt.Sprintf("%s:%s@/", user, password)
	db.conn, err = sql.Open(server, db.datasource)
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

func (db *database) Close() {
	db.conn.Close()
}

func (db *database) Ping() error {
	return db.conn.Ping()
}

func (db *database) listDatabase() []string {
	var err error

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
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

		if database == "information_schema" || database == "mysql" ||
			database == "performance_schema" || database == "nbinfo" {
			continue
		}

		list = append(list, database)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return list
}

func (db *database) createDatabase(cr CreateRequest) error {
	err := db.dbExists(cr.DatabaseName)
	if err != nil {
		return err
	}

	err = db.userExists(cr.Username)
	if err != nil {
		return err
	}

	// Begin transaction so that we can roll it back at any point something goes wrong.
	tx, err := db.conn.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("CREATE DATABASE %s CHARSET utf8;", cr.DatabaseName))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("CREATE USER '%s' IDENTIFIED BY '%s';", cr.Username, cr.Password))
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = db.conn.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%s';", cr.DatabaseName, cr.Username, "%"))
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

func (db *database) dbExists(databasename string) error {
	var count int

	err := db.conn.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", databasename).Scan(&count)
	if err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("Database '%s' already exists", databasename)
	}

	return nil
}

func (db *database) userExists(username string) error {
	var count int

	err := db.conn.QueryRow("SELECT count(*) FROM mysql.user WHERE user = ?", username).Scan(&count)
	if err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("User '%s' already exists", username)
	}

	return nil
}
