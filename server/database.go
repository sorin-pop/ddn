package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/djavorszky/sutils"

	_ "github.com/go-sql-driver/mysql"
)

type mysql struct {
	conn *sql.DB
}

func (db *mysql) connect(c Config) error {
	var err error

	datasource := fmt.Sprintf("%s:%s@tcp(%s:%s)/", c.DBUser, c.DBPass, c.DBAddress, c.DBPort)
	err = db.connectDS(datasource)
	if err != nil {
		return fmt.Errorf("couldn't connect to the database: %s", err.Error())
	}

	_, err = db.conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARSET utf8;", c.DBName))
	if err != nil {
		return fmt.Errorf("executing create database query failed: %s", sutils.TrimNL(err.Error()))
	}

	db.conn.Close()

	datasource = datasource + c.DBName

	err = db.connectDS(datasource)
	if err != nil {
		return fmt.Errorf("couldn't connect to the database: %s", err.Error())
	}

	_, err = db.conn.Exec("CREATE TABLE IF NOT EXISTS `databases` ( `id` INT NOT NULL AUTO_INCREMENT, `dbname` VARCHAR(255) NULL, `dbuser` VARCHAR(255) NULL, `dbpass` VARCHAR(255) NULL, `dumpfile` LONGTEXT NULL, `createDate` DATETIME NULL, `creator` VARCHAR(255) NULL, `connectorName` VARCHAR(255) NULL,  PRIMARY KEY (`id`));")
	if err != nil {
		return fmt.Errorf("executing create table query failed: %s", sutils.TrimNL(err.Error()))
	}

	return nil
}

func (db *mysql) connectDS(datasource string) error {
	var err error

	db.conn, err = sql.Open("mysql", datasource)
	if err != nil {
		return fmt.Errorf("creating connection pool failed: %s", err.Error())
	}

	err = db.conn.Ping()
	if err != nil {
		db.conn.Close()
		return fmt.Errorf("database ping failed: %s", sutils.TrimNL(err.Error()))
	}

	return nil
}

// Close closes the connection to the database
func (db *mysql) close() {
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
		return fmt.Errorf("executing stayalive query failed: %s", sutils.TrimNL(err.Error()))
	}

	return nil

}

func (db *mysql) persist(dbentry DBEntry) error {
	if err := db.Alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	query := fmt.Sprintf("INSERT INTO `databases` (`dbname`, `dbuser`, `dbpass`, `dumpfile`, `createDate`, `creator`, `connectorName`) VALUES ('%s', '%s', '%s', '%s', NOW(), '%s', '%s')",
		dbentry.DBName,
		dbentry.DBUser,
		dbentry.DBPass,
		dbentry.Dumpfile,
		dbentry.Creator,
		dbentry.ConnectorName,
	)

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("executing insert query failed: %s", err.Error())
	}

	return nil
}

func (db *mysql) list() ([]DBEntry, error) {
	var entries []DBEntry

	rows, err := db.conn.Query("SELECT id, dbname, dbuser, dbpass, dumpfile, createDate, creator, connectorName FROM `databases`")
	if err != nil {
		return nil, fmt.Errorf("couldn't execute query: %s", err.Error())
	}

	for rows.Next() {
		var row DBEntry

		err = rows.Scan(&row.ID, &row.DBName, &row.DBUser, &row.DBPass, &row.Dumpfile, &row.CreateDate, &row.Creator, &row.ConnectorName)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error reading result from query: %s", err.Error())
	}

	return entries, nil
}

// DBEntry represents a row in the "databases" table.
type DBEntry struct {
	ID            int
	DBVendor      string
	DBName        string
	DBUser        string
	DBPass        string
	Dumpfile      string
	CreateDate    string
	Creator       string
	ConnectorName string
}
