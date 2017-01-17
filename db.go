package main

// Database interface to be used when running queries. All DB implementations
// should implement all its methods.
type Database interface {
	// Connect creates and initialises a Database struct and connects to the database
	Connect(user, password, DBPort string) error

	// Close closes the connection to the database
	Close()

	// Alive checks whether the connection is alive. Returns error if not.
	Alive() error

	// CreateDatabase creates a Database along with a user, to which all privileges
	// are granted on the created database. Fails if database or user already exists.
	CreateDatabase(dbRequest DBRequest) error

	// DropDatabase drops a database and a user. Always succeeds, even if droppable database or
	// user does not exist
	DropDatabase(dbRequest DBRequest) error

	// ImportDatabase imports the dumpfile to the database or returns an error
	// if it failed for some reason.
	ImportDatabase(dbRequest DBRequest) error

	// ListDatabase returns a list of strings - the names of the databases in the server
	// All system tables are omitted from the returned list. If there's an error, it is returned.
	ListDatabase() ([]string, error)

	// Version returns the database server's version.
	Version() (string, error)
}
