package main

import (
	"log"
	"os"
	"os/user"
)

func checkProps() (string, error) {
	prop, err := checkLocalProps()
	if err == nil {
		return prop, err
	}

	prop, err = checkHomeProps()
	if err == nil {
		return prop, err
	}

	return "", err
}

func checkLocalProps() (string, error) {
	if _, err := os.Stat("ddnc.properties"); err != nil {
		if os.IsNotExist(err) {
			log.Println("ddnc.properties not found next to executable")
			return "", err
		}
	}
	log.Println("ddnc.properties found next to executable")
	return "ddnc.properties", nil
}

func checkHomeProps() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	path := usr.HomeDir + "/.ddnc/ddnc.properties"
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Println("ddnc.properties not found at", path)
			return "", err
		}
	}
	log.Println("ddnc.properties found at", path)

	return path, nil
}

func generateProps() (string, error) {
	log.Println("Generating properties file for MySQL")

	prop := `
vendor="mysql"
version="5.5.53"
executable="/usr/bin/mysql"
dbport="3306"
username="root"
password="root"
masterAddress="127.0.0.1"`

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	ddncFolder := usr.HomeDir + "/.ddnc"

	_, err = os.Stat(ddncFolder)
	if err != nil {
		err = os.Mkdir(ddncFolder, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	path := ddncFolder + "/ddnc.properties"

	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString(prop)
	if err != nil {
		return "", err
	}

	file.Sync()

	return path, nil
}
