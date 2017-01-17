package main

import (
	"log"
	"os"
	"text/template"
)

func generateProps(vendor, filename string) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	conf := NewConfig(vendor)

	prop := `vendor="{{.Vendor}}"
version="{{.Version}}"
executable="{{.Exec}}"
dbport="{{.DBPort}}"
connectorPort="{{.ConnectorPort}}"
username="{{.User}}"
password="{{.Password}}"
masterAddress="{{.MasterAddress}}"`

	tmpl, err := template.New("props").Parse(prop)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(file, conf)
	if err != nil {
		log.Fatal(err)
	}

	file.Sync()

	return filename, nil
}
