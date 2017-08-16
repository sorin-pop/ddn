package database

import (
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/database/mysql"
)

// Vendor represents a database backend to which the server can connect
type Vendor int

// Available vendors
const (
	MySQL Vendor = iota
	SQLite
)

// Choose sets the database vendor to be used.
func Choose(v Vendor) {
}

// ConnectAndPrepare establishes a database connection and initializes the tables, if needed
func ConnectAndPrepare(address, port, user, pass, database string) error {
	return mysql.ConnectAndPrepare(address, port, user, pass, database)
}

// FetchByID returns the entry associated with that ID, or
// an error if it does not exist
func FetchByID(ID int) (data.Row, error) {
	return mysql.FetchByID(ID)
}

// FetchByCreator returns public entries that were created by the
// specified user, an empty list if it's not the user does
// not have any entries, or an error if something went
// wrong
func FetchByCreator(creator string) ([]data.Row, error) {
	return mysql.FetchByCreator(creator)
}

// Insert adds an entry to the database, returning its ID
func Insert(entry *data.Row) error {
	return mysql.Insert(entry)
}

// Update updates an already existing entry
func Update(entry *data.Row) error {
	return mysql.Update(entry)
}

// Delete removes the entry from the database
func Delete(entry data.Row) error {
	return mysql.Delete(entry)
}

// FetchPublic returns all entries that have "Public" set to true
func FetchPublic() ([]data.Row, error) {
	return mysql.FetchPublic()
}

// FetchAll returns all entries.
func FetchAll() ([]data.Row, error) {
	return mysql.FetchPublic()
}

// Close closes the database connection
func Close() error {
	return mysql.Close()
}
