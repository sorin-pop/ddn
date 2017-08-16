package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/djavorszky/sutils"
)

type dbUpdate struct {
	Query   string
	Comment string
}

var queries = []dbUpdate{
	dbUpdate{
		Query:   "CREATE TABLE `version` (`queryId` INT NOT NULL AUTO_INCREMENT, `query` LONGTEXT NULL, `comment` TEXT NULL, `date` DATETIME NULL, PRIMARY KEY (`queryId`));",
		Comment: "Create the version table",
	},
	dbUpdate{
		Query:   "CREATE TABLE IF NOT EXISTS `databases` ( `id` INT NOT NULL AUTO_INCREMENT, `dbname` VARCHAR(255) NULL, `dbuser` VARCHAR(255) NULL, `dbpass` VARCHAR(255) NULL, `dbsid` VARCHAR(45) NULL, `dumpfile` LONGTEXT NULL, `createDate` DATETIME NULL, `expiryDate` DATETIME NULL, `creator` VARCHAR(255) NULL, `connectorName` VARCHAR(255) NULL, `dbAddress` VARCHAR(255) NULL, `dbPort` VARCHAR(45) NULL, `dbvendor` VARCHAR(255) NULL, `status` INT,  PRIMARY KEY (`id`));",
		Comment: "Create the databases table",
	},
	dbUpdate{
		Query:   "ALTER TABLE `databases` ADD COLUMN `visibility` INT(11) NULL DEFAULT 0 AFTER `status`;",
		Comment: "Add 'visibility' to databases, default 0",
	},
	dbUpdate{
		Query:   "ALTER TABLE `databases` ADD COLUMN `message` LONGTEXT AFTER `status`;",
		Comment: "Add 'message' column",
	},
	dbUpdate{
		Query:   "UPDATE `databases` SET `message` = '' WHERE `message` IS NULL;",
		Comment: "Update 'message' columns to empty where null",
	},
}

// ConnectAndPrepare establishes a database connection and initializes the tables, if needed
func ConnectAndPrepare(address, port, user, pass, database string) error {
	var err error

	datasource := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, pass, address, port)
	err = connect(datasource)
	if err != nil {
		return fmt.Errorf("couldn't connect to the database: %s", err.Error())
	}

	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARSET utf8;", database))
	if err != nil {
		return fmt.Errorf("executing create database query failed: %s", sutils.TrimNL(err.Error()))
	}

	conn.Close()

	datasource = datasource + database

	err = connect(datasource)
	if err != nil {
		return fmt.Errorf("couldn't connect to the database: %s", err.Error())
	}

	err = initTables()
	if err != nil {
		return fmt.Errorf("initializing tables failed: %s", err.Error())
	}

	return nil
}

func connect(datasource string) error {
	var err error

	conn, err = sql.Open("mysql", datasource+"?parseTime=true")
	if err != nil {
		return fmt.Errorf("creating connection pool failed: %s", err.Error())
	}

	err = conn.Ping()
	if err != nil {
		conn.Close()
		return fmt.Errorf("database ping failed: %s", sutils.TrimNL(err.Error()))
	}

	return nil
}

func initTables() error {
	var (
		err      error
		startLoc int
	)

	conn.QueryRow("SELECT count(*) FROM `version`").Scan(&startLoc)

	for _, q := range queries[startLoc:] {
		log.Printf("Updating database %q", q.Comment)
		_, err = conn.Exec(q.Query)
		if err != nil {
			return fmt.Errorf("executing query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}

		_, err = conn.Exec("INSERT INTO `version` (query, comment, date) VALUES (?, ?, ?)", q.Query, q.Comment, time.Now())
		if err != nil {
			return fmt.Errorf("updating version table with query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}
	}

	return nil
}
