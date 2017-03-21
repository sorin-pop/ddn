package main

import (
	"log"

	"github.com/djavorszky/prompter"
)

// Config to hold the database server and ddn server configuration
type Config struct {
	DBAddress  string `toml:"db.address"`
	DBPort     string `toml:"db.port"`
	DBUser     string `toml:"db.user.name"`
	DBPass     string `toml:"db.user.password"`
	DBName     string `toml:"db.name"`
	ServerHost string `toml:"server.host"`
	ServerPort string `toml:"server.port"`
}

// Print prints the configuration to the log.
func (c Config) Print() {
	log.Printf("Database Address:\t\t%s", c.DBAddress)
	log.Printf("Database Port:\t\t%s", c.DBPort)
	log.Printf("Database User:\t\t%s", c.DBUser)
	log.Printf("Database Password:\t\t****")
	log.Printf("Database Name:\t\t%s", c.DBName)
	log.Printf("Server Host:\t\t%s", c.ServerHost)
	log.Printf("Server Port:\t\t%s", c.ServerPort)
}

func newConfig() Config {
	return Config{
		DBAddress:  "localhost",
		DBPort:     "3306",
		DBUser:     "root",
		DBPass:     "root",
		DBName:     "ddn",
		ServerHost: "localhost",
		ServerPort: "7010",
	}
}

func setup(filename string) (*string, Config) {
	var config Config

	def := newConfig()

	config.DBPort = prompter.AskDef("What is the database port?", def.DBPort)
	config.DBAddress = prompter.AskDef("What is the database address?", def.DBAddress)
	config.DBUser = prompter.AskDef("Who is the database user?", def.DBUser)
	config.DBPass = prompter.AskDef("What is the database password?", def.DBPass)
	config.DBName = prompter.AskDef("What should the database's name be?", def.DBName)
	config.ServerHost = prompter.AskDef("What is the server's hostname?", def.ServerHost)
	config.ServerPort = prompter.AskDef("What should the server's port be?", def.ServerPort)

	fname := prompter.AskDef("What should we name the configuration file?", filename)

	return &fname, config
}
