package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/ddn/server/brwsr"
	"github.com/djavorszky/ddn/server/database"
	"github.com/djavorszky/ddn/server/mail"
)

var (
	config Config
)

const version = "1.5.1"

func main() {
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

			fmt.Printf("Logfile %s already exists, rotating it to %s", *logname, rotated)

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
		log.Println("Couldn't find properties file, generating one.")

		filename, config = setup(*filename)

		if err := createProps(*filename, config); err != nil {
			log.Fatal("couldn't create properties file:", err.Error())
		}
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

	err = database.ConnectAndPrepare(config.DBAddress, config.DBPort, config.DBUser, config.DBPass, config.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

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

	port := fmt.Sprintf(":%s", config.ServerPort)

	log.Println(http.ListenAndServe(port, Router()))

	if len(config.AdminEmail) != 0 {
		for _, addr := range config.AdminEmail {
			mail.Send(addr, "[Cloud DB] Server went down", fmt.Sprintf(`<p>Cloud DB down for some reason.</p>`))
		}
	}
}
