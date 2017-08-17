package main

import (
	"log"
)

// Config to hold the database server and ddn server configuration
type Config struct {
	DBProvider  string   `toml:"db-provider"`
	DBAddress   string   `toml:"db-addr"`
	DBPort      string   `toml:"db-port"`
	DBUser      string   `toml:"db-username"`
	DBPass      string   `toml:"db-userpass"`
	DBName      string   `toml:"db-name"`
	ServerHost  string   `toml:"server-host"`
	ServerPort  string   `toml:"server-port"`
	SMTPAddr    string   `toml:"smtp-host"`
	SMTPPort    int      `toml:"smtp-port"`
	SMTPUser    string   `toml:"smtp-user"`
	SMTPPass    string   `toml:"smtp-password"`
	EmailSender string   `toml:"email-sender"`
	AdminEmail  []string `toml:"admin-emails"`
	MountLoc    string   `toml:"mount-loc"`
}

// Print prints the configuration to the log.
func (c Config) Print() {
	log.Printf("Database Provider:\t\t%s", c.DBProvider)

	if c.DBProvider == "mysql" {
		log.Printf("Database Address:\t\t%s", c.DBAddress)
		log.Printf("Database Port:\t\t%s", c.DBPort)
		log.Printf("Database User:\t\t%s", c.DBUser)
		log.Printf("Database Name:\t\t%s", c.DBName)
	} else if c.DBProvider == "sqlite" {
		log.Printf("Database file location: %s", c.DBAddress)
	}

	log.Printf("Server Host:\t\t%s", c.ServerHost)
	log.Printf("Server Port:\t\t%s", c.ServerPort)

	if c.SMTPAddr != "" && c.SMTPPort != 0 && c.EmailSender != "" {
		log.Printf("Admin email:\t\t%s", c.AdminEmail)
		log.Printf("Server configured to send emails.")
	}
}
