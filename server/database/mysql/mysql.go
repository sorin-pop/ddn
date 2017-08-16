package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/sutils"

	// Db
	_ "github.com/go-sql-driver/mysql"
)

// DB implements the BackendConnection
type DB struct {
	Address, Port, User, Pass, Database string
	conn                                *sql.DB
}

// ConnectAndPrepare establishes a database connection and initializes the tables, if needed
func (mys *DB) ConnectAndPrepare() error {
	datasource := fmt.Sprintf("%s:%s@tcp(%s:%s)/", mys.User, mys.Pass, mys.Address, mys.Port)
	err := mys.connect(datasource)
	if err != nil {
		return fmt.Errorf("couldn't connect to the database: %s", err.Error())
	}

	_, err = mys.conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARSET utf8;", mys.Database))
	if err != nil {
		return fmt.Errorf("executing create database query failed: %s", sutils.TrimNL(err.Error()))
	}

	mys.Close()

	datasource = datasource + mys.Database

	err = mys.connect(datasource)
	if err != nil {
		return fmt.Errorf("couldn't connect to the database: %s", err.Error())
	}

	err = mys.initTables()
	if err != nil {
		return fmt.Errorf("initializing tables failed: %s", err.Error())
	}

	return nil
}

// Close closes the database connection
func (mys *DB) Close() error {
	return mys.conn.Close()
}

// FetchByID returns the entry associated with that ID, or
// an error if it does not exist
func (mys *DB) FetchByID(ID int) (data.Row, error) {
	if err := mys.alive(); err != nil {
		return data.Row{}, fmt.Errorf("database down: %s", err.Error())
	}

	row := mys.conn.QueryRow("SELECT * FROM `databases` WHERE id = ?", ID)
	res, err := mys.readRow(row)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading result: %v", err)
	}

	return res, nil
}

// FetchByCreator returns public entries that were created by the
// specified user, an empty list if it's not the user does
// not have any entries, or an error if something went
// wrong
func (mys *DB) FetchByCreator(creator string) ([]data.Row, error) {
	if err := mys.alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := mys.conn.Query("SELECT * FROM `databases` WHERE creator = ? AND visibility = 0 ORDER BY id DESC", creator)
	if err != nil {
		return nil, fmt.Errorf("couldn't execute query: %s", err.Error())
	}

	for rows.Next() {
		row, err := mys.readRows(rows)
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

// FetchPublic returns all entries that have "Public" set to true
func (mys *DB) FetchPublic() ([]data.Row, error) {
	if err := mys.alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := mys.conn.Query("SELECT * FROM `databases` WHERE visibility = 1 ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed running query: %v", err)
	}

	for rows.Next() {
		row, err := mys.readRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	return entries, nil
}

// FetchAll returns all entries.
func (mys *DB) FetchAll() ([]data.Row, error) {
	if err := mys.alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := mys.conn.Query("SELECT * FROM `databases` ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed running query: %v", err)
	}

	for rows.Next() {
		row, err := mys.readRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	return entries, nil
}

// Insert adds an entry to the database, returning its ID
func (mys *DB) Insert(entry *data.Row) error {
	if err := mys.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	query := "INSERT INTO `databases` (`dbname`, `dbuser`, `dbpass`, `dbsid`, `dumpfile`, `createDate`, `expiryDate`, `creator`, `connectorName`, `dbAddress`, `dbPort`, `dbvendor`, `status`, `message`, `visibility`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?, ?, ?, ?, ?)"

	res, err := mys.conn.Exec(query,
		entry.DBName,
		entry.DBUser,
		entry.DBPass,
		entry.DBSID,
		entry.Dumpfile,
		entry.CreateDate,
		entry.ExpiryDate,
		entry.Creator,
		entry.ConnectorName,
		entry.DBAddress,
		entry.DBPort,
		entry.DBVendor,
		entry.Status,
		entry.Message,
		entry.Public,
	)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed getting new ID: %v", err)
	}

	entry.ID = int(id)

	return nil
}

// Update updates an already existing entry
func (mys *DB) Update(entry *data.Row) error {
	if err := mys.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	var count int

	err := mys.conn.QueryRow("SELECT count(*) FROM `databases` WHERE id = ?", entry.ID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed existence check: %v", err)
	}

	if count == 0 {
		return mys.Insert(entry)
	}

	query := "UPDATE `databases` SET `dbname`= ?, `dbuser`= ?, `dbpass`= ?, `dbsid`= ?, `dumpfile`= ?, `createDate`= ?, `expiryDate`= ?, `creator`= ?, `connectorName`= ?, `dbAddress`= ?, `dbPort`= ?, `dbvendor`= ?, `status`= ?, `message`= ?, `visibility`= ? WHERE id = ?"

	_, err = mys.conn.Exec(query,
		entry.DBName,
		entry.DBUser,
		entry.DBPass,
		entry.DBSID,
		entry.Dumpfile,
		entry.CreateDate,
		entry.ExpiryDate,
		entry.Creator,
		entry.ConnectorName,
		entry.DBAddress,
		entry.DBPort,
		entry.DBVendor,
		entry.Status,
		entry.Message,
		entry.Public,
		entry.ID,
	)
	if err != nil {
		return fmt.Errorf("failed update: %v", err)
	}

	return nil
}

// Delete removes the entry from the database
func (mys *DB) Delete(entry data.Row) error {
	if err := mys.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	_, err := mys.conn.Exec("DELETE FROM `databases` WHERE id = ?", entry.ID)

	return err
}

func (mys *DB) readRow(result *sql.Row) (data.Row, error) {
	var row data.Row

	err := result.Scan(
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
		&row.Status,
		&row.Message,
		&row.Public)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading row: %v", err)
	}

	return row, nil
}

func (mys *DB) readRows(rows *sql.Rows) (data.Row, error) {
	var row data.Row

	err := rows.Scan(
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
		&row.Status,
		&row.Message,
		&row.Public)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading row: %v", err)
	}

	return row, nil
}

// Alive checks whether the connection is alive. Returns error if not.
func (mys *DB) alive() error {
	defer func() {
		if p := recover(); p != nil {
			log.Println("Panic Attack! Database seems to be down.")
		}
	}()

	_, err := mys.conn.Exec("select * from `databases` WHERE 1 = 0")
	if err != nil {
		return fmt.Errorf("executing stayalive query failed: %s", sutils.TrimNL(err.Error()))
	}

	return nil
}

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

func (mys *DB) connect(datasource string) error {
	db, err := sql.Open("mysql", datasource+"?parseTime=true")
	if err != nil {
		return fmt.Errorf("creating connection pool failed: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return fmt.Errorf("database ping failed: %s", sutils.TrimNL(err.Error()))
	}
	mys.conn = db

	return nil
}

func (mys *DB) initTables() error {
	var (
		err      error
		startLoc int
	)

	mys.conn.QueryRow("SELECT count(*) FROM `version`").Scan(&startLoc)

	for _, q := range queries[startLoc:] {
		log.Printf("Updating database %q", q.Comment)
		_, err = mys.conn.Exec(q.Query)
		if err != nil {
			return fmt.Errorf("executing query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}

		_, err = mys.conn.Exec("INSERT INTO `version` (query, comment, date) VALUES (?, ?, ?)", q.Query, q.Comment, time.Now())
		if err != nil {
			return fmt.Errorf("updating version table with query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}
	}

	return nil
}
