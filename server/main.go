package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/server/brwsr"
	"github.com/djavorszky/ddn/server/database"
	"github.com/djavorszky/ddn/server/database/mysql"
	"github.com/djavorszky/ddn/server/database/sqlite"
	"github.com/djavorszky/ddn/server/mail"
	"github.com/djavorszky/sutils"
)

var (
	workdir string
	config  Config
	db      database.BackendConnection
)

var version = ""

func main() {
	path, _ := filepath.Abs(os.Args[0])
	workdir = filepath.Dir(path)

	defer func() {
		if p := recover(); p != nil {
			if len(config.AdminEmail) != 0 {
				for _, addr := range config.AdminEmail {
					mail.Send(addr, "[FATAL] Cloud DB server panicked", fmt.Sprintf("%v", p))
				}
			}
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c
		// Received kill
		logger.Fatal("Received signal to terminate.")
	}()

	var err error
	filename := flag.String("p", "server.conf", "Specify the configuration file's name")
	logname := flag.String("l", "std", "Specify the log's filename. By default, logs to the terminal.")

	flag.Parse()

	if *logname != "std" {
		if _, err = os.Stat(*logname); err == nil {
			rotated := fmt.Sprintf("%s.%s", *logname, time.Now().Format("2006-01-02_03:04"))

			logger.Debug("Rotated logfile to %s", rotated)

			os.Rename(*logname, rotated)
		}

		logOut, err := os.OpenFile(*logname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("error opening file %s, will continue logging to stderr: %s", *logname, err.Error())
			logOut = os.Stderr
		}
		defer logOut.Close()

		log.SetOutput(logOut)
	}

	loadProperties(*filename)

	logger.Info("Version: %s", version)

	logger.Info("Starting with properties:")

	config.Print()

	if config.MountLoc != "" {
		err = brwsr.Mount(config.MountLoc)
		if err != nil {
			logger.Warn("Couldn't mount folder: %s", err)

			config.MountLoc = ""
		} else {
			logger.Info("Mounted folder %q", config.MountLoc)
		}
	}

	switch config.DBProvider {
	case "mysql":
		db = &mysql.DB{
			Address:  config.DBAddress,
			Port:     config.DBPort,
			User:     config.DBUser,
			Pass:     config.DBPass,
			Database: config.DBName,
		}
	case "sqlite":
		db = &sqlite.DB{DBLocation: config.DBAddress}
	default:
		logger.Fatal("Unknown database provider: %s", config.DBProvider)
	}

	err = db.ConnectAndPrepare()
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Database connection established")

	if config.SMTPAddr != "" {
		if config.SMTPUser != "" {
			err = mail.Init(config.SMTPAddr, config.SMTPPort, config.SMTPUser, config.SMTPPass, config.EmailSender)
		} else {
			err = mail.InitNoAuth(config.SMTPAddr, config.SMTPPort, config.EmailSender)
		}

		if err != nil {
			logger.Warn("Mail failed to initialize: %v", err)
		} else {
			logger.Info("Mail initialized")
		}
	}

	if config.WebPushEnabled {
		if !sutils.Present(config.VAPIDPrivateKey, config.WebPushSubscriber) {
			logger.Error("WebPush is enabled but no private key / subscriber email specified! Disabling WebPush.")
			config.WebPushEnabled = false
		} else {
			logger.Info("WebPush is enabled.")
		}
	}

	// Start maintenance goroutine
	go maintain()

	// Start agent checker goroutine
	go checkAgents()

	logger.Info("Starting to listen on port %s", config.ServerPort)

	port := fmt.Sprintf(":%s", config.ServerPort)

	logger.Error("%v", http.ListenAndServe(port, Router()))

	if len(config.AdminEmail) != 0 {
		for _, addr := range config.AdminEmail {
			mail.Send(addr, "[Cloud DB] Server went down", fmt.Sprintf(`<p>Cloud DB down for some reason.</p>`))
		}
	}
}

func loadProperties(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Warn("Couldn't find properties file, trying to download one.")

		tmpConfig, err := inet.DownloadFile(".", "https://raw.githubusercontent.com/djavorszky/ddn/master/server/srv.conf")
		if err != nil {
			logger.Fatal("Could not fetch configuration file, please download it manually from https://github.com/djavorszky/ddn")
		}

		os.Rename(tmpConfig, filename)

		logger.Info("Continuing with default configuration...")
	}

	if _, err := toml.DecodeFile(filename, &config); err != nil {
		logger.Fatal("couldn't read configuration file: %v", err)
	}

}
