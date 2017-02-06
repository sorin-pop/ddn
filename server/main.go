package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			// PANIC! Do something.
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Received SIGTERM, do something.
		os.Exit(1)
	}()

	var err error
	filename := flag.String("p", "server.conf", "Specify the configuration file's name")

	flag.Parse()

	var conf Config
	if _, err = os.Stat(*filename); os.IsNotExist(err) {
		log.Println("Couldn't find properties file, generating one.")

		filename, conf = setup(*filename)

		if err := createProps(*filename, conf); err != nil {
			log.Fatal("couldn't create properties file:", err.Error())
		}
	}

	if _, err := toml.DecodeFile(*filename, &conf); err != nil {
		log.Fatal("couldn't read configuration file: ", err.Error())
	}

	initRegistry()

	log.Println("Starting with properties:")

	conf.Print()

	port := fmt.Sprintf(":%s", conf.ServerPort)

	log.Fatal(http.ListenAndServe(port, Router()))
}
