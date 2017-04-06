package main

import (
	"bytes"
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

var panicked bool

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

	err = db.initTables()
	if err != nil {
		return fmt.Errorf("initializing tables failed: %s", err.Error())
	}

	return nil
}

type dbUpdate struct {
	Query   string
	Comment string
}

func (db *mysql) initTables() error {
	queries := []dbUpdate{
		dbUpdate{
			Query:   "CREATE TABLE `ddn`.`version` (`queryId` INT NOT NULL AUTO_INCREMENT, `query` LONGTEXT NULL, `comment` TEXT NULL, `date` DATETIME NULL, PRIMARY KEY (`queryId`));",
			Comment: "Create the version table",
		},
		dbUpdate{
			Query:   "CREATE TABLE IF NOT EXISTS `databases` ( `id` INT NOT NULL AUTO_INCREMENT, `dbname` VARCHAR(255) NULL, `dbuser` VARCHAR(255) NULL, `dbpass` VARCHAR(255) NULL, `dbsid` VARCHAR(45) NULL, `dumpfile` LONGTEXT NULL, `createDate` DATETIME NULL, `expiryDate` DATETIME NULL, `creator` VARCHAR(255) NULL, `connectorName` VARCHAR(255) NULL, `dbAddress` VARCHAR(255) NULL, `dbPort` VARCHAR(45) NULL, `dbvendor` VARCHAR(255) NULL, `status` INT,  PRIMARY KEY (`id`));",
			Comment: "Create the databases table",
		},
	}

	var (
		err      error
		startLoc int
	)
	db.conn.QueryRow("SELECT count(*) FROM `version`").Scan(&startLoc)

	for _, q := range queries[startLoc:] {
		log.Printf("Updating database %q", q.Comment)
		_, err = db.conn.Exec(q.Query)
		if err != nil {
			return fmt.Errorf("executing query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}

		_, err = db.conn.Exec("INSERT INTO `version` (query, comment, date) VALUES (?, ?, NOW())", q.Query, q.Comment)
		if err != nil {
			return fmt.Errorf("updating version table with query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}
	}

	return nil
}

func (db *mysql) connectDS(datasource string) error {
	var err error

	db.conn, err = sql.Open("mysql", datasource+"?parseTime=true")
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

	_, err := db.conn.Exec("select * from `databases` WHERE 1 = 0")
	if err != nil {
		if !panicked && config.AdminEmail != "" {
			sendMail(config.AdminEmail, "[Cloud DB] Local database down", fmt.Sprintf("<p>Something wrong:</p>%s", err.Error()))
			panicked = true
		}

		return fmt.Errorf("executing stayalive query failed: %s", sutils.TrimNL(err.Error()))
	}

	if panicked {
		sendMail(config.AdminEmail, "[Cloud DB] Local database back online", "Yay")
		panicked = false
	}

	return nil
}

func (db *mysql) persist(dbentry model.DBEntry) (int64, error) {
	if err := db.Alive(); err != nil {
		return 0, fmt.Errorf("database down: %s", err.Error())
	}

	query := fmt.Sprintf("INSERT INTO `databases` (`dbname`, `dbuser`, `dbpass`, `dbsid`, `dumpfile`, `createDate`, `expiryDate`, `creator`, `connectorName`, `dbAddress`, `dbPort`, `dbvendor`, `status`) VALUES ('%s', '%s', '%s', '%s', '%s', NOW(), NOW() + INTERVAL 30 DAY, '%s', '%s', '%s','%s', '%s', %d)",
		dbentry.DBName,
		dbentry.DBUser,
		dbentry.DBPass,
		dbentry.DBSID,
		dbentry.Dumpfile,
		dbentry.Creator,
		dbentry.ConnectorName,
		dbentry.DBAddress,
		dbentry.DBPort,
		dbentry.DBVendor,
		dbentry.Status,
	)

	res, err := db.conn.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("executing insert query failed: %s", err.Error())
	}

	return res.LastInsertId()
}

func (db *mysql) delete(id int64) {
	query := fmt.Sprintf("DELETE FROM `databases` WHERE id = %d", id)

	db.conn.Exec(query)
}

func (db *mysql) list() ([]model.DBEntry, error) {
	if err := db.Alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []model.DBEntry

	rows, err := db.conn.Query("SELECT id, dbname, dbuser, dbpass, dbsid, dumpfile, createDate, expiryDate, creator, connectorName, dbAddress, dbPort, dbVendor, status FROM `databases`")
	if err != nil {
		return nil, fmt.Errorf("couldn't execute query: %s", err.Error())
	}

	for rows.Next() {
		var row model.DBEntry

		err = rows.Scan(
			&row.ID,
			&row.DBName,
			&row.DBUser,
			&row.DBPass,
			&row.DBSID,
			&row.Dumpfile,
			&row.CreateDate,
			&row.ExpiryDate,
			&row.Creator,
			&row.ConnectorName,
			&row.DBAddress,
			&row.DBPort,
			&row.DBVendor,
			&row.Status)
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

type clause struct {
	Column string
	Value  interface{}
}

func (db *mysql) listWhere(clauses ...clause) ([]model.DBEntry, error) {
	if err := db.Alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var buf bytes.Buffer

	buf.WriteString("SELECT id, dbname, dbuser, dbpass, dbsid, dumpfile, createDate, expiryDate, creator, connectorName, dbAddress, dbPort, dbVendor, status FROM `databases` WHERE 1=1")

	for _, clause := range clauses {
		buf.WriteString(" AND ")
		buf.WriteString(clause.Column)
		buf.WriteString("='")
		buf.WriteString(fmt.Sprintf("%v", clause.Value))
		buf.WriteString("'")
	}

	buf.WriteString(" ORDER BY id DESC")

	rows, err := db.conn.Query(buf.String())
	if err != nil {
		return nil, fmt.Errorf("couldn't execute query: %s", err.Error())
	}

	var entries []model.DBEntry

	for rows.Next() {
		var row model.DBEntry

		err = rows.Scan(
			&row.ID,
			&row.DBName,
			&row.DBUser,
			&row.DBPass,
			&row.DBSID,
			&row.Dumpfile,
			&row.CreateDate,
			&row.ExpiryDate,
			&row.Creator,
			&row.ConnectorName,
			&row.DBAddress,
			&row.DBPort,
			&row.DBVendor,
			&row.Status)
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

func (db *mysql) entryByID(ID int64) (model.DBEntry, error) {
	if err := db.Alive(); err != nil {
		return model.DBEntry{}, fmt.Errorf("database down: %s", err.Error())
	}

	var entry model.DBEntry

	row := db.conn.QueryRow("SELECT id, dbname, dbuser, dbpass, dbsid, dumpfile, createDate, expiryDate, creator, connectorName, dbAddress, dbPort, dbVendor, status FROM `databases` WHERE id = ?", ID)

	row.Scan(
		&entry.ID,
		&entry.DBName,
		&entry.DBUser,
		&entry.DBPass,
		&entry.DBSID,
		&entry.Dumpfile,
		&entry.CreateDate,
		&entry.ExpiryDate,
		&entry.Creator,
		&entry.ConnectorName,
		&entry.DBAddress,
		&entry.DBPort,
		&entry.DBVendor,
		&entry.Status)

	return entry, nil
}

func (db *mysql) updateColumn(ID int, column string, value interface{}) error {
	if err := db.Alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	q := fmt.Sprintf("UPDATE `databases` SET %s=%v WHERE id=%d", column, value, ID)

	_, err := db.conn.Exec(q)
	if err != nil {
		return fmt.Errorf("failed to update: %s", err.Error())
	}

	return nil
}
