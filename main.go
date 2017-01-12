package main

import (
	"log"

	"net/http"

	"github.com/BurntSushi/toml"
)

var (
	properties string
	conf       Config
	db         database

	port = ":7000"
)

func main() {

	properties, err := checkProps()
	if err != nil {
		log.Println("Couldn't find properties file, generating one")
		file, err := generateProps()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("Generated %s with dummy values next to executable. Please update it with real values and restart the connector", file)
	}

	if _, err := toml.DecodeFile(properties, &conf); err != nil {
		panic(err)
	}

	log.Println("Starting with properties:")
	log.Println("Vendor:\t\t", conf.Vendor)
	log.Println("Version:\t\t", conf.Version)
	log.Println("Database port:\t", conf.DBPort)
	log.Println("Executable:\t\t", conf.Exec)
	log.Println("Username:\t\t", conf.User)
	log.Println("Password:\t\t ******")
	log.Println("Master address:\t", conf.MasterAddress)

	err = db.Connect(conf.Vendor, conf.User, conf.Password, conf.DBPort)
	if err != nil {
		log.Fatal("Could not establish database connection:\n\t\t", err.Error())
	}
	defer db.Close()

	log.Println("Database connection established")
	log.Println("Starting to listen on port", port)

	log.Fatal(http.ListenAndServe(port, Router()))

}
