package mysql

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/sutils"

	// Db
	_ "github.com/go-sql-driver/mysql"
)

var (
	dbname   string
	panicked bool
	conn     *sql.DB
)

// FetchByID returns the entry associated with that ID, or
// an error if it does not exist
func FetchByID(ID int) (data.Row, error) {
	if err := alive(); err != nil {
		return data.Row{}, fmt.Errorf("database down: %s", err.Error())
	}

	row := conn.QueryRow("SELECT * FROM `databases` WHERE id = ?", ID)
	res, err := readRow(row)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading result: %v", err)
	}

	return res, nil
}

// FetchByCreator returns public entries that were created by the
// specified user, an empty list if it's not the user does
// not have any entries, or an error if something went
// wrong
func FetchByCreator(creator string) ([]data.Row, error) {
	if err := alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := conn.Query("SELECT * FROM `databases` WHERE creator = ? AND visibility = 0 ORDER BY id DESC", creator)
	if err != nil {
		return nil, fmt.Errorf("couldn't execute query: %s", err.Error())
	}

	for rows.Next() {
		row, err := readRows(rows)
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

// Insert adds an entry to the database, returning its ID
func Insert(entry *data.Row) error {
	if err := alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	query := "INSERT INTO `databases` (`dbname`, `dbuser`, `dbpass`, `dbsid`, `dumpfile`, `createDate`, `expiryDate`, `creator`, `connectorName`, `dbAddress`, `dbPort`, `dbvendor`, `status`, `message`, `visibility`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?, ?, ?, ?, ?)"

	res, err := conn.Exec(query,
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
func Update(entry *data.Row) error {
	if err := alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	var count int

	err := conn.QueryRow("SELECT count(*) FROM `databases` WHERE id = ?", entry.ID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed existence check: %v", err)
	}

	if count == 0 {
		return Insert(entry)
	}

	query := "UPDATE `databases` SET `dbname`= ?, `dbuser`= ?, `dbpass`= ?, `dbsid`= ?, `dumpfile`= ?, `createDate`= ?, `expiryDate`= ?, `creator`= ?, `connectorName`= ?, `dbAddress`= ?, `dbPort`= ?, `dbvendor`= ?, `status`= ?, `message`= ?, `visibility`= ? WHERE id = ?"

	_, err = conn.Exec(query,
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
func Delete(entry data.Row) error {
	if err := alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	_, err := conn.Exec("DELETE FROM `databases` WHERE id = ?", entry.ID)

	return err
}

// FetchPublic returns all entries that have "Public" set to true
func FetchPublic() ([]data.Row, error) {
	if err := alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := conn.Query("SELECT * FROM `databases` WHERE visibility = 1 ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed running query: %v", err)
	}

	for rows.Next() {
		row, err := readRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	return entries, nil
}

// FetchAll returns all entries.
func FetchAll() ([]data.Row, error) {
	if err := alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := conn.Query("SELECT * FROM `databases` ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed running query: %v", err)
	}

	for rows.Next() {
		row, err := readRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	return entries, nil
}

func readRow(result *sql.Row) (data.Row, error) {
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

func readRows(rows *sql.Rows) (data.Row, error) {
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
func alive() error {
	defer func() {
		if p := recover(); p != nil {
			log.Println("Panic Attack! Database seems to be down.")
		}
	}()

	_, err := conn.Exec("select * from `databases` WHERE 1 = 0")
	if err != nil {
		return fmt.Errorf("executing stayalive query failed: %s", sutils.TrimNL(err.Error()))
	}

	return nil
}

// Close closes the database connection
func Close() error {
	return conn.Close()
}
