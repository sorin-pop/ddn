package main

import (
	"log"
	"strconv"

	"github.com/djavorszky/prompter"
)

// Config to hold the database server and ddn server configuration
type Config struct {
	DBAddress   string `toml:"db.address"`
	DBPort      string `toml:"db.port"`
	DBUser      string `toml:"db.user.name"`
	DBPass      string `toml:"db.user.password"`
	DBName      string `toml:"db.name"`
	ServerHost  string `toml:"server.host"`
	ServerPort  string `toml:"server.port"`
	SMTPAddr    string `toml:"smtp.host"`
	SMTPPort    int    `toml:"smtp.port"`
	SMTPUser    string `toml:"smtp.user"`
	SMTPPass    string `toml:"smtp.password"`
	EmailSender string `toml:"email.sender"`
	AdminEmail  string `toml:"admin.email"`
	UseCDN      bool   `toml:"use.cdn"`
	MountLoc    string `toml:"mount.loc"`
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
	log.Printf("Using CDN:\t\t\t%t", c.UseCDN)

	if c.MountLoc != "" {
		log.Printf("Mounted folder:\t\t%s", c.MountLoc)
	}

	if c.SMTPAddr != "" && c.SMTPPort != 0 && c.EmailSender != "" {
		log.Printf("Admin email:\t\t%s", c.AdminEmail)
		log.Printf("Server configured to send emails.")
	}
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
		AdminEmail: "webmaster@example.com",
		UseCDN:     true,
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

	config.SMTPAddr = prompter.Ask("What is the SMTP address?")
	config.SMTPPort, _ = strconv.Atoi(prompter.Ask("What is the SMTP port?"))
	config.SMTPUser = prompter.Ask("Who is the SMTP user?")
	config.SMTPPass = prompter.Ask("What is the password of the SMTP user?")
	config.EmailSender = prompter.Ask("What address should be used to send the emails from?")
	config.AdminEmail = prompter.AskDef("Who should be notified if something goes horribly wrong?", def.AdminEmail)
	config.UseCDN = prompter.AskBoolDef("Should we serve third party javascript, css and fonts from CDN?", true)
	config.MountLoc = prompter.Ask("If you want to mount a folder for browsing dumps, please specify now")

	fname := prompter.AskDef("What should we name the configuration file?", filename)

	return &fname, config
}
