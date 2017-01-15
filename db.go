package main

// Database interface to be used when running queries. All DB implementations
// should implement all its methods.
type Database interface {
	Connect(server, user, password, DBPort string) error
	Close()
	Alive() error
	CreateDatabase(dbRequest DBRequest) error
	DropDatabase(dbRequest DBRequest) error
	ImportDatabase(dbRequest DBRequest) error
	ListDatabase() ([]string, error)
}
