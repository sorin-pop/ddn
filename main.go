package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/BurntSushi/toml"
)

const version = "0.6.8"

var (
	conf    Config
	db      Database
	port    string
	usr     *user.User
	startup time.Time
)

func main() {
	var err error
	filename := flag.String("p", "ddnc.conf", "Specify the configuration file's name")

	flag.Parse()

	usr, err = user.Current()
	if err != nil {
		log.Fatal("couldn't get default user:", err.Error())
	}

	if _, err = os.Stat(*filename); os.IsNotExist(err) {
		log.Println("Couldn't find properties file, generating one.")

		err := generateProps(*filename)
		if err != nil {
			log.Fatal("properties generation failed:", err.Error())
		}
	}

	if _, err := toml.DecodeFile(*filename, &conf); err != nil {
		log.Fatal("couldn't read configuration file: ", err.Error())
	}

	if _, err = os.Stat(conf.Exec); os.IsNotExist(err) {
		log.Fatal("database executable doesn't exist:", conf.Exec)
	}

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

	log.Println("Starting to listen on port", conf.ConnectorPort)

	port = fmt.Sprintf(":%s", conf.ConnectorPort)

	startup = time.Now()

	log.Fatal(http.ListenAndServe(port, Router()))
}
