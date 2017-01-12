package main

import (
	"log"

	"net/http"

	"github.com/BurntSushi/toml"
)

var (
	properties string
	conf       Config
	port       = ":7000"
)

func main() {

	properties, err := checkProps()
	if err != nil {
		path, err := generateProps()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("Generated %s with dummy values inside. Please update it with real values and restart the client", path)
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

	validateConnection()

	log.Println("Starting to listen on port", port)

	log.Fatal(http.ListenAndServe(port, Router()))

}
