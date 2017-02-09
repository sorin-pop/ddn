package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/djavorszky/ddn/common/model"
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

func (db *mysql) persist(req model.ClientRequest) error {
	if err := db.Alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	query := fmt.Sprintf("INSERT INTO `databases` (`dbname`, `dbuser`, `dbpass`, `dumpfile`, `createDate`, `creator`, `connectorName`) VALUES ('%s', '%s', '%s', '%s', NOW(), '%s', '%s')",
		req.DatabaseName,
		req.Username,
		req.Password,
		req.DumpLocation,
		req.Requester,
		req.ConnectorIdentifier,
	)

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("executing insert query failed: %s", err.Error())
	}

	return nil
}
