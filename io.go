package main

import (
	"log"
	"os"
)

func generateProps(filename string) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	prop := `vendor="mysql"
version="5.5.53"
executable="/usr/bin/mysql"
dbport="3306"
connectorPort="7000"
username="root"
password="root"
masterAddress="127.0.0.1"`

	_, err = file.WriteString(prop)
	if err != nil {
		return "", err
	}

	file.Sync()

	return filename, nil
}
