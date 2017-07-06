package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/djavorszky/prompter"
)

// Config to hold the database server and connector information
type Config struct {
	Vendor          string `toml:"db-vendor"`
	Version         string `toml:"db-version"`
	Exec            string `toml:"db-executable"`
	User            string `toml:"db-username"`
	Password        string `toml:"db-userpass"`
	SID             string `toml:"oracle-sid"`
	Tablespace      string `toml:"oracle-tablespace"`
	LocalDBAddr     string `toml:"db-local-addr"`
	LocalDBPort     string `toml:"db-local-port"`
	ConnectorDBHost string `toml:"db-remote-addr"`
	ConnectorDBPort string `toml:"db-remote-port"`
	ConnectorAddr   string `toml:"connector-addr"`
	ConnectorPort   string `toml:"connector-port"`
	ShortName       string `toml:"connector-shortname"`
	ConnectorName   string `toml:"connector-longname"`
	MasterAddress   string `toml:"server-address"`
}

// Print prints the Config object to the log.
func (c Config) Print() {
	log.Printf("Vendor:\t\t%s\n", conf.Vendor)
	log.Printf("Version:\t\t%s\n", conf.Version)
	log.Printf("Executable:\t\t%s\n", conf.Exec)

	log.Printf("Username:\t\t%s\n", conf.User)
	log.Printf("Password:\t\t****\n")

	if conf.Vendor == "oracle" {
		log.Printf("SID:\t\t%s", conf.SID)
		log.Printf("Tablespace:\t\t%s", conf.Tablespace)
	}

	log.Printf("Local DB addr:\t%s\n", conf.LocalDBAddr)
	log.Printf("Local DB port:\t%s\n", conf.LocalDBPort)

	log.Printf("Remote DB addr:\t%s\n", conf.ConnectorDBHost)
	log.Printf("Remote DB port:\t%s\n", conf.ConnectorDBPort)

	log.Printf("Connector addr:\t%s\n", conf.ConnectorAddr)
	log.Printf("Connector port:\t%s\n", conf.ConnectorPort)

	log.Printf("Short name:\t\t%s\n", conf.ShortName)
	log.Printf("Connector name:\t%s\n", conf.ConnectorName)

	log.Printf("Master address:\t%s\n", conf.MasterAddress)
}

// NewConfig returns a configuration file based on the vendor
func NewConfig(vendor string) Config {
	var conf Config

	switch vendor {
	case "mysql":
		conf = Config{
			Vendor:          "mysql",
			Version:         "5.5.53",
			ShortName:       "mysql-55",
			LocalDBPort:     "3306",
			LocalDBAddr:     "localhost",
			ConnectorPort:   "7000",
			ConnectorAddr:   "http://localhost",
			ConnectorDBPort: "3306",
			ConnectorDBHost: "localhost",
			User:            "root",
			Password:        "root",
			MasterAddress:   "http://localhost:7010",
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
			Vendor:          "postgres",
			Version:         "9.4.9",
			ShortName:       "postgre-94",
			LocalDBPort:     "5432",
			LocalDBAddr:     "localhost",
			ConnectorPort:   "7000",
			ConnectorAddr:   "http://localhost",
			ConnectorDBPort: "5432",
			ConnectorDBHost: "localhost",
			User:            "postgres",
			Password:        "password",
			MasterAddress:   "http://localhost:7010",
		}

		switch runtime.GOOS {
		case "windows":
			conf.Exec = "C:\\path\\to\\psql.exe"
		case "darwin":
			conf.Exec = "/Library/PostgreSQL/9.4/bin/psql"
		default:
			conf.Exec = "/usr/bin/psql"
		}
	case "oracle":
		conf = Config{
			Vendor:          "oracle",
			Version:         "11g",
			ShortName:       "oracle-11g",
			LocalDBPort:     "1521",
			LocalDBAddr:     "localhost",
			ConnectorPort:   "7000",
			ConnectorAddr:   "http://localhost",
			ConnectorDBPort: "1521",
			ConnectorDBHost: "localhost",
			User:            "system",
			Password:        "password",
			SID:             "orcl",
			Tablespace:      "USERS",
			MasterAddress:   "http://localhost:7010",
		}
		switch runtime.GOOS {
		case "windows":
			conf.Exec = "C:\\path\\to\\sqlplus.exe"
		case "darwin":
			conf.Exec = "/path/to/sqlplus"
		default:
			conf.Exec = "/path/to/sqlplus"
		}
	}

	conf.ConnectorName = fmt.Sprintf("%s-%s", hostname, conf.ShortName)

	return conf
}

func generateInteractive(filename string) (string, Config) {
	var (
		ok    = false
		vType = 0
	)

	for !ok {
		vType, ok = prompter.AskSelectionDef("What is the database vendor?", 0, vendors)
	}

	vendor := vendors[vType]
	def := NewConfig(vendor)

	var config Config

	config.Vendor = vendor
	config.Version = prompter.Ask("What is the database version?")

	config.User = prompter.AskDef("Who is the database user?", def.User)
	config.Password = prompter.AskDef("What is the user's password?", def.Password)

	if vendor == "oracle" {
		config.Exec = prompter.AskDef("Where is the sqlplus executable?", def.Exec)
		config.SID = prompter.AskDef("What is the SID?", def.SID)
		config.Tablespace = prompter.AskDef("What is the default tablespace?", def.Tablespace)
	} else if vendor == "mysql" {
		config.Exec = prompter.AskDef("Where is the mysql executable?", def.Exec)
	} else if vendor == "postgres" {
		config.Exec = prompter.AskDef("Where is the psql executable?", def.Exec)
	}

	config.LocalDBAddr = prompter.AskDef("What is the database address when connecting locally?", def.LocalDBAddr)
	config.LocalDBPort = prompter.AskDef("What is the database port when connecting locally?", def.LocalDBPort)

	config.ConnectorDBHost = prompter.AskDef("What is the database address when connecting remotely?", def.ConnectorDBHost)
	config.ConnectorDBPort = prompter.AskDef("What is the database port when connecting remotely?", def.ConnectorDBPort)

	config.ConnectorAddr = prompter.AskDef("What is the connector's remote address?", def.ConnectorAddr)
	config.ConnectorPort = prompter.AskDef("What should the connector's remote port be?", def.ConnectorPort)

	config.ShortName = prompter.AskDef("What should the connector's short name be?", def.ShortName)
	config.ConnectorName = prompter.AskDef("What should the connector's identifier be?", def.ConnectorName)

	config.MasterAddress = prompter.AskDef("What is the address of the Master server?", def.MasterAddress)

	fname := prompter.AskDef("What should we name the configuration file?", filename)

	return fname, config
}
