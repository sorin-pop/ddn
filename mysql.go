package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var (
	conn string
	db   *sql.DB
	err  error

	dbname    = "ddnc"
	tablename = "info"
)

func prepDatabase() {
	conn = fmt.Sprintf("%s:%s@/", conf.User, conf.Password)
	db, err = sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	validateConn()
	validateMetaDB()
}

func validateConn() {

	err := db.Ping()
	if err != nil {
		log.Fatalf("Could not validate database connection:\n\t%s", err.Error())
	}

	log.Println("MySQL connection validated")
}

func validateMetaDB() {
	var count int

	err := db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbname).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 1 {
		return
	}

	log.Printf("Database '%s' does not exist", dbname)

	createDB()

	log.Printf("Database '%s' and table '%s' created", dbname, tablename)
}

func createDB() {
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s CHARSET UTF8", dbname))
	if err != nil {
		log.Fatal(err)
	}

	createDBStatement := `CREATE TABLE %s.%s (
  ID INT NOT NULL AUTO_INCREMENT,
  databaseName VARCHAR(255) NULL,
  tablespaceName VARCHAR(255) NULL,
  tablespaceFileLocation MEDIUMTEXT NULL,
  dbUser VARCHAR(255) NULL,
  dbPass VARCHAR(255) NULL,
  createDate TIMESTAMP NULL,
  requestedBy VARCHAR(255) NULL,
  importFileLocation MEDIUMTEXT NULL,
  hidden INT NULL,
  PRIMARY KEY (ID));`

	_, err = db.Exec(fmt.Sprintf(createDBStatement, dbname, tablename))
	if err != nil {
		log.Fatal(err)
	}
	// TODO continue
}
