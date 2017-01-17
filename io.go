package main

import (
	"log"
	"os"
	"runtime"
	"text/template"
)

func generateProps(filename string) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	conf := Config{"mysql", "5.5.53", "/usr/bin/mysql", "3306", "7000", "root", "root", "127.0.0.1"}

	switch runtime.GOOS {
	case "windows":
		conf.Exec = "C:\\path\\to\\mysql.exe"
	default:
		conf.Exec = "/usr/bin/mysql"
	}

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
