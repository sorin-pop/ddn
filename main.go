package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

var (
	properties string
	conf       Config
)

func main() {

	properties, err := checkProperties()
	if err != nil {
		path, err := generateProps()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("Generated %s with dummy values inside. Please edit it and restart the client", path)
	}

	if _, err := toml.DecodeFile(properties, &conf); err != nil {
		fmt.Println(err)
		panic(err)
	}

	log.Println("Starting with properties:")
	log.Println("Vendor:\t\t", conf.Vendor)
	log.Println("Version:\t\t", conf.Version)
	log.Println("Executable:\t\t", conf.Exec)
	log.Println("Username:\t\t", conf.User)
	log.Println("Password:\t\t ******")
	log.Println("Master address:\t", conf.MasterAddress)

}
