package sqlite

import "github.com/djavorszky/ddn/server/database/data"

// DB implements the BackendConnection
type DB struct{}

// ConnectAndPrepare establishes a database connection and initializes the tables, if needed
func (db *DB) ConnectAndPrepare() error {
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return nil
}

// FetchByID returns the entry associated with that ID, or
// an error if it does not exist
func (db *DB) FetchByID(ID int) (data.Row, error) {
	return data.Row{}, nil
}

// FetchByCreator returns public entries that were created by the
// specified user, an empty list if it's not the user does
// not have any entries, or an error if something went
// wrong
func (db *DB) FetchByCreator(creator string) ([]data.Row, error) {
	return []data.Row{}, nil
}

// FetchPublic returns all entries that have "Public" set to true
func (db *DB) FetchPublic() ([]data.Row, error) {
	return []data.Row{}, nil
}

// FetchAll returns all entries.
func (db *DB) FetchAll() ([]data.Row, error) {
	return []data.Row{}, nil
}

// Insert adds an entry to the database, returning its ID
func (db *DB) Insert(row *data.Row) error {
	return nil
}

// Update updates an already existing entry
func (db *DB) Update(row *data.Row) error {
	return nil
}

// Delete removes the entry from the database
func (db *DB) Delete(row data.Row) error {
	return nil
}
