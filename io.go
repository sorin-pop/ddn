package main

import (
	"fmt"
	"os"
	"text/template"
)

func generateProps(filename string) error {
	filename, conf := generateInteractive(filename)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("couldn't create file: %s", err.Error())
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
		return fmt.Errorf("couldn't parse template: %s", err.Error())
	}

	err = tmpl.Execute(file, conf)
	if err != nil {
		return fmt.Errorf("couldn't execute template: %s", err.Error())
	}

	file.Sync()

	return nil
}
