package main

import (
	"log"
	"os"
)

func checkProps() (string, error) {
	if _, err := os.Stat("ddnc.properties"); os.IsNotExist(err) {
		return "", err
	}

	return "ddnc.properties", nil
}

func generateProps() (string, error) {
	file, err := os.Create("ddnc.properties")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	prop := `vendor="mysql"
version="5.5.53"
executable="/usr/bin/mysql"
dbport="3306"
username="root"
password="root"
masterAddress="127.0.0.1"`

	_, err = file.WriteString(prop)
	if err != nil {
		return "", err
	}

	file.Sync()

	return "ddnc.properties", nil
}
