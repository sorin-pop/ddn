package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/disco"
)

const version = "2.0.2"

var (
	conf       Config
	db         Database
	port       string
	usr        *user.User
	hostname   string
	startup    time.Time
	registered bool

	connector model.Connector
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			logger.Error("Panic... Unregistering")
			unregisterConnector()
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		unregisterConnector()
		os.Exit(1)
	}()

	var err error
	filename := flag.String("p", "ddnc.conf", "Specify the configuration file's name")
	logname := flag.String("l", "std", "Specify the log's filename. If set to std, logs to the terminal.")

	flag.Parse()

	loadProperties(*filename)

	if *logname != "std" {
		if _, err := os.Stat(*logname); err == nil {
			rotated := fmt.Sprintf("%s.%d", *logname, time.Now().Unix())

			os.Rename(*logname, rotated)
		}

		logOut, err := os.OpenFile(*logname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("error opening file %s, will continue logging to stderr: %s", *logname, err.Error())
			logOut = os.Stderr
		}
		defer logOut.Close()

		log.SetOutput(logOut)
	}
	if err = grabRemoteLogger(); err != nil {
		logger.Warn("Couldn't get remote logger: %v", err)
	}
	defer logger.Close()

	usr, err = user.Current()
	if err != nil {
		logger.Fatal("couldn't get default user: ", err.Error())
	}

	hostname, err = os.Hostname()
	if err != nil {
		logger.Fatal("couldn't get hostname: ", err.Error())
	}

	db, err = GetDB(conf.Vendor)
	if err != nil {
		logger.Fatal("couldn't get database instance:", err)
	}

	logger.Info("Starting with properties:")
	conf.Print()

	err = db.Connect(conf)
	if err != nil {
		logger.Fatal("couldn't establish database connection:", err.Error())
	}
	defer db.Close()
	logger.Info("Database connection established")

	ver, err := db.Version()
	if err != nil {
		logger.Fatal("database: %v", err)
	}

	if ver != conf.Version {
		logger.Warn("Version mismatch, please update configuration file:")
		logger.Warn("> Configuration:\t%s", conf.Version)
		logger.Warn("> Read from DB:\t%s", ver)

		conf.Version = ver
	}

	// Check and create the 'dumps' folder
	if _, err = os.Stat(filepath.Join(".", "dumps")); os.IsNotExist(err) {
		err = os.Mkdir("dumps", os.ModePerm)
		if err != nil {
			logger.Fatal("Couldn't create dumps folder, please create it manually: %v", err)
		}

		logger.Info("Created 'dumps' folder")
	}

	// For Oracle, create or replace the stored procedure that executes the import, by running the sql/oracle/import_procedure.sql file
	if odb, ok := db.(*oracle); ok {
		logger.Info("Creating or replacing the import_dump stored procedure.")
		err := odb.RefreshImportStoredProcedure()
		if err != nil {
			logger.Fatal("oracle: %v", err)
		}
	}

	err = registerConnector()
	if err != nil {
		logger.Error("could not register connector: %s", err.Error())
		logger.Error(">> will try to connect to it if it comes online")
	}

	go keepAlive()

	// Announce presence
	if err := announce(); err != nil {
		logger.Fatal("Failed announcing presence: %v", err)
	}

	logger.Info("Starting to listen on port %s", conf.ConnectorPort)

	port = fmt.Sprintf(":%s", conf.ConnectorPort)

	startup = time.Now()

	logger.Fatal("server: %v", http.ListenAndServe(port, Router()))
}

func loadProperties(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Warn("Couldn't find properties file, trying to download one.")

		tmpConfig, err := inet.DownloadFile(".", "https://raw.githubusercontent.com/djavorszky/ddn/master/connector/con.conf")
		if err != nil {
			logger.Fatal("Could not fetch configuration file, please download it manually from https://github.com/djavorszky/ddn")
		}

		os.Rename(tmpConfig, filename)

		logger.Info("Continuing with default configuration...")
	}

	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		logger.Fatal("couldn't read configuration file: ", err.Error())
	}

	if _, err := os.Stat(conf.Exec); os.IsNotExist(err) {
		logger.Fatal("database executable doesn't exist:", conf.Exec)
	}
}

func grabRemoteLogger() error {
	logger.Info("Trying to get remote logger")

	host := fmt.Sprintf("%s:%s", conf.ConnectorAddr, conf.ConnectorPort)
	service, err := disco.Query("224.0.0.1:9999", host, "rlog", 4*time.Second)
	if err != nil {
		return fmt.Errorf("query rlog: %v", err)
	}

	return logger.UseRemoteLogger(service.Addr, "clouddb", conf.ShortName)
}

func announce() error {
	host := fmt.Sprintf("%s:%s", conf.ConnectorAddr, conf.ConnectorPort)
	err := disco.Announce("224.0.0.1:9999", host, conf.ShortName)
	if err != nil {
		return fmt.Errorf("announce: %v", err)
	}

	return nil
}
