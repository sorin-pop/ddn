package main

import (
	"fmt"
	"html/template"
	"os"
)

func createProps(filename string, conf Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("couldn't create file: %s", err.Error())
	}
	defer file.Close()

	prop := `dbaddress="{{.DBAddress}}"
dbport="{{.DBPort}}"
dbuser="{{.DBUser}}"
dbpass="{{.DBPass}}"
dbname="{{.DBName}}"
serverport="{{.ServerPort}}"
`
	tmpl, err := template.New("props").Parse(prop)
	if err != nil {
		return fmt.Errorf("couldn't parse template: %s", err.Error())
	}

	err = tmpl.Execute(file, conf)
	if err != nil {
		return fmt.Errorf("couldn't execute template: %s", err.Error())
	}

	file.Sync()

	return nil
}
