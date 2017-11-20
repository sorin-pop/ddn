package main

import "github.com/djavorszky/ddn/common/logger"

// Config to hold the database server and ddn server configuration
type Config struct {
	DBProvider      string   `toml:"db-provider"`
	DBAddress       string   `toml:"db-addr"`
	DBPort          string   `toml:"db-port"`
	DBUser          string   `toml:"db-username"`
	DBPass          string   `toml:"db-userpass"`
	DBName          string   `toml:"db-name"`
	ServerHost      string   `toml:"server-host"`
	ServerPort      string   `toml:"server-port"`
	SMTPAddr        string   `toml:"smtp-host"`
	SMTPPort        int      `toml:"smtp-port"`
	SMTPUser        string   `toml:"smtp-user"`
	SMTPPass        string   `toml:"smtp-password"`
	EmailSender     string   `toml:"email-sender"`
	AdminEmail      []string `toml:"admin-emails"`
	MountLoc        string   `toml:"mount-loc"`
	VAPIDPrivateKey string   `toml:"vapid-private-key"`
}

// Print prints the configuration to the log.
func (c Config) Print() {
	logger.Info("Database Provider:\t\t%s", c.DBProvider)

	if c.DBProvider == "mysql" {
		logger.Info("Database Address:\t\t%s", c.DBAddress)
		logger.Info("Database Port:\t\t%s", c.DBPort)
		logger.Info("Database User:\t\t%s", c.DBUser)
		logger.Info("Database Name:\t\t%s", c.DBName)
	} else if c.DBProvider == "sqlite" {
		logger.Info("Database file location:\t%s", c.DBAddress)
	}

	logger.Info("Server Host:\t\t%s", c.ServerHost)
	logger.Info("Server Port:\t\t%s", c.ServerPort)

	if c.SMTPAddr != "" && c.SMTPPort != 0 && c.EmailSender != "" {
		logger.Info("Admin email:\t\t%s", c.AdminEmail)
		logger.Info("Server configured to send emails.")
	}
}
