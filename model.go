package main

// CreateRequest is used to represent JSON call to create a database
type CreateRequest struct {
	DatabaseName   string `json:"database_name"`
	TablespaceName string `json:"tablespace_name"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}
