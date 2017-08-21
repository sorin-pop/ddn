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

	"github.com/djavorszky/ddn/common/inet"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/ddn/server/brwsr"
	"github.com/djavorszky/ddn/server/database"
	"github.com/djavorszky/ddn/server/database/mysql"
	"github.com/djavorszky/ddn/server/database/sqlite"
	"github.com/djavorszky/ddn/server/mail"
)

var (
	workdir string
	config  Config
	db      database.BackendConnection
)

const version = "2.0.0"

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
		log.Println("Received signal to terminate.")
		os.Exit(1)
	}()

	var err error
	filename := flag.String("p", "server.conf", "Specify the configuration file's name")
	logname := flag.String("l", "std", "Specify the log's filename. By default, logs to the terminal.")

	flag.Parse()

	if *logname != "std" {
		if _, err = os.Stat(*logname); err == nil {
			rotated := fmt.Sprintf("%s.%d", *logname, time.Now().Unix())

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

	if _, err = os.Stat(*filename); os.IsNotExist(err) {
		log.Println("Couldn't find properties file, trying to download one.")

		tmpConfig, err := inet.DownloadFile(".", "https://raw.githubusercontent.com/djavorszky/ddn/master/server/srv.conf")
		if err != nil {
			log.Fatal("Could not fetch configuration file, please download it manually from https://github.com/djavorszky/ddn")
		}

		os.Rename(tmpConfig, *filename)

		log.Println("Continuing with default configuration...")
	}

	if _, err := toml.DecodeFile(*filename, &config); err != nil {
		log.Fatal("couldn't read configuration file: ", err.Error())
	}

	log.Println("Starting with properties:")

	config.Print()

	if config.MountLoc != "" {
		err = brwsr.Mount(config.MountLoc)
		if err != nil {
			log.Printf("Couldn't mount folder: %s", err.Error())

			config.MountLoc = ""
		} else {
			log.Printf("Mounted folder %q", config.MountLoc)
		}
	}

	if config.MountLoc != "" {
		if _, err = os.Stat(config.MountLoc); os.IsNotExist(err) {
			log.Printf("Mounted folder does not exist, unsetting it")
			config.MountLoc = ""
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
		log.Fatalf("Unknown database provider: %s", config.DBProvider)
	}

	err = db.ConnectAndPrepare()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	if config.SMTPAddr != "" {
		if config.SMTPUser != "" {
			err = mail.Init(config.SMTPAddr, config.SMTPPort, config.SMTPUser, config.SMTPPass, config.EmailSender)
		} else {
			err = mail.InitNoAuth(config.SMTPAddr, config.SMTPPort, config.EmailSender)
		}

		if err != nil {
			log.Printf("Mail failed to initialize: %v", err)
		} else {
			log.Printf("Mail initialized")
		}
	}

	// Start maintenance goroutine
	go maintain()

	// Start connector checker goroutine
	go checkConnectors()

	log.Printf("Starting to listen on port %s", config.ServerPort)

	port := fmt.Sprintf(":%s", config.ServerPort)
	log.Println(http.ListenAndServe(port, Router()))

	if len(config.AdminEmail) != 0 {
		for _, addr := range config.AdminEmail {
			mail.Send(addr, "[Cloud DB] Server went down", fmt.Sprintf(`<p>Cloud DB down for some reason.</p>`))
		}
	}
}
