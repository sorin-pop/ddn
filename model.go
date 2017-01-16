package main

// DBRequest is used to represent JSON call to create a database
type DBRequest struct {
	DatabaseName string `json:"database_name"`
	DumpLocation string `json:"dumpfile_location"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}
