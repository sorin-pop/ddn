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
	"github.com/djavorszky/ddn/common/model"
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
			log.Println("Panic... Unregistering")
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
	usr, err = user.Current()
	if err != nil {
		log.Fatal("couldn't get default user: ", err.Error())
	}

	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal("couldn't get hostname: ", err.Error())
	}

	loadProperties(*filename)

	db, err = GetDB(conf.Vendor)
	if err != nil {
		log.Fatal("couldn't get database instance:", err)
	}

	log.Println("Starting with properties:")
	conf.Print()

	err = db.Connect(conf)
	if err != nil {
		log.Fatal("couldn't establish database connection:", err.Error())
	}
	defer db.Close()
	log.Println("Database connection established")

	ver, err := db.Version()
	if err != nil {
		log.Fatal(err)
	}

	if ver != conf.Version {
		log.Println("Version mismatch, please update configuration file:")
		log.Println("> Configuration:\t", conf.Version)
		log.Println("> Read from DB:\t", ver)

		conf.Version = ver
	}

	// Check and create the 'dumps' folder
	if _, err = os.Stat(filepath.Join(".", "dumps")); os.IsNotExist(err) {
		err = os.Mkdir("dumps", os.ModePerm)
		if err != nil {
			log.Fatalf("Couldn't create dumps folder, please create it manually: %s", err.Error())
		}

		log.Println("Created 'dumps' folder")
	} else {
		log.Println("'dumps' folder already exists.")
	}

	// For Oracle, create or replace the stored procedure that executes the import, by running the sql/oracle/import_procedure.sql file
	if odb, ok := db.(*oracle); ok {
		log.Println("Creating or replacing the import_dump stored procedure.")
		err := odb.RefreshImportStoredProcedure()
		if err != nil {
			log.Fatal(err)
		}
	}

	err = registerConnector()
	if err != nil {
		log.Printf("could not register connector: %s", err.Error())
		log.Println(">> will try to connect to it if it comes online")
	}

	go keepAlive()

	log.Println("Starting to listen on port", conf.ConnectorPort)

	port = fmt.Sprintf(":%s", conf.ConnectorPort)

	startup = time.Now()

	log.Fatal(http.ListenAndServe(port, Router()))
}

func setupLog(logname string) (*File, error) {

}

func loadProperties(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Println("Couldn't find properties file, trying to download one.")

		tmpConfig, err := inet.DownloadFile(".", "https://raw.githubusercontent.com/djavorszky/ddn/master/connector/con.conf")
		if err != nil {
			log.Fatal("Could not fetch configuration file, please download it manually from https://github.com/djavorszky/ddn")
		}

		os.Rename(tmpConfig, filename)

		log.Println("Continuing with default configuration...")
	}

	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		log.Fatal("couldn't read configuration file: ", err.Error())
	}

	if _, err := os.Stat(conf.Exec); os.IsNotExist(err) {
		log.Fatal("database executable doesn't exist:", conf.Exec)
	}
}
