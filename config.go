package main

import (
	"fmt"
	"runtime"
)

// Config to hold the database server information
type Config struct {
	Vendor        string `toml:"vendor"`
	Version       string `toml:"version"`
	Exec          string `toml:"executable"`
	DBPort        string `toml:"dbport"`
	ConnectorPort string `toml:"connectorPort"`
	User          string `toml:"username"`
	Password      string `toml:"password"`
	MasterAddress string `toml:"masterAddress"`
}

// NewConfig returns a configuration file based on the vendor
func NewConfig(vendor string) Config {
	var conf Config

	switch vendor {
	case "mysql":
		conf = Config{
			Vendor:        "mysql",
			Version:       "5.5.53",
			DBPort:        "3306",
			ConnectorPort: "7000",
			User:          "root",
			Password:      "root",
			MasterAddress: "127.0.0.1",
		}

		switch runtime.GOOS {
		case "windows":
			conf.Exec = "C:\\path\\to\\mysql.exe"
		case "darwin":
			conf.Exec = "/usr/local/mysql/bin/mysql"
		default:
			conf.Exec = "/usr/bin/mysql"
		}
	case "postgres":
		conf = Config{
			Vendor:        "postgres",
			Version:       "9.4.9",
			DBPort:        "5432",
			ConnectorPort: "7000",
			User:          "postgres",
			Password:      "password",
			MasterAddress: "127.0.0.1",
		}

		switch runtime.GOOS {
		case "windows":
			conf.Exec = "C:\\path\\to\\psql.exe"
		case "darwin":
			conf.Exec = "/Library/PostgreSQL/9.4/bin/createdb"
		default:
			conf.Exec = "/usr/bin/psql"
		}
	}

	return conf
}

// VendorSupported returns an error if the specified vendor is not supported.
func VendorSupported(vendor string) error {
	switch vendor {
	case "mysql", "postgres":
		return nil
	}
	return fmt.Errorf("Vendor %s not supported.", vendor)
}
