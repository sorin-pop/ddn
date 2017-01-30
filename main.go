package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"

	"github.com/BurntSushi/toml"
)

var (
	conf Config
	db   Database
	port string
	usr  *user.User
)

func main() {
	var err error

	filename := flag.String("p", "ddnc.properties", "Specify the configuration file's name")
	vendor := flag.String("v", "mysql", "Specify the vendor's name.")
	flag.Parse()

	if err = VendorSupported(*vendor); err != nil {
		log.Fatal(err)
	}

	usr, err = user.Current()
	if err != nil {
		log.Fatal("Couldn't get default user.")
	}

	if _, err = os.Stat(*filename); os.IsNotExist(err) {
		log.Println("Couldn't find properties file, generating one")
		file, err := generateProps(*vendor, *filename)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("Generated '%s' with dummy values next to executable. Please update it with real values and restart the connector", file)
	}

	if _, err := toml.DecodeFile(*filename, &conf); err != nil {
		log.Fatal(err)
	}

	if _, err = os.Stat(conf.Exec); os.IsNotExist(err) {
		log.Fatalf("Database executable '%s' doesn't exist.", conf.Exec)
	}

	db, err = GetDB(conf.Vendor)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting with properties:")
	conf.Print()

	err = db.Connect(conf)
	if err != nil {
		log.Fatal("Could not establish database connection:\n\t\t", err.Error())
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

	log.Fatal(http.ListenAndServe(port, Router()))

}
