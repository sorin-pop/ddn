package main

import (
	"log"
	"os"
	"text/template"
)

func generateProps(filename string) error {
	filename, conf := generateInteractive(filename)

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	prop := `vendor="{{.Vendor}}"
version="{{.Version}}"
executable="{{.Exec}}"
dbport="{{.DBPort}}"
dbAddress="{{.DBAddress}}"
connectorPort="{{.ConnectorPort}}"
username="{{.User}}"
password="{{.Password}}"
masterAddress="{{.MasterAddress}}"
`

	if conf.SID != "" {
		prop += "oracle-sid=\"{{.SID}}\"\n"
	}

	if conf.DefaultTablespace != "" {
		prop += "default-tablespace=\"{{.DefaultTablespace}}\"\n"
	}

	tmpl, err := template.New("props").Parse(prop)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(file, conf)
	if err != nil {
		log.Fatal(err)
	}

	file.Sync()

	return nil
}
