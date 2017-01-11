package main

// Config to hold the database server information
type Config struct {
	Vendor        string `toml:"vendor"`
	Version       string `toml:"version"`
	Exec          string `toml:"executable"`
	User          string `toml:"username"`
	Password      string `toml:"password"`
	MasterAddress string `toml:"masterAddress"`
}
