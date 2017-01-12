package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// index should display whenever someone visits the main page.
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the index!")
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var dbinfo CreateRequest

	err := decoder.Decode(&dbinfo)
	if err != nil {
		log.Println(err)
	}
	var msg Message

	if dbinfo.DatabaseName == "" || dbinfo.Username == "" || dbinfo.Password == "" {
		log.Println("Missing required field from JSON message:")
		log.Println("> DatabaseName:\t", dbinfo.DatabaseName)
		log.Println("> Username:\t\t", dbinfo.Username)
		log.Println("> Password:\t\t", dbinfo.Password)

		msg.Status = http.StatusBadRequest
		msg.Message = "One or more required fields are missing from the call"
	} else {
		err = db.createDatabase(dbinfo)
		if err != nil {
			msg.Status = http.StatusInternalServerError
			msg.Message = err.Error()
		} else {
			msg.Status = http.StatusOK
			msg.Message = "Successfully created the Database and user!"
		}
	}

	writeHeader(w, msg.Status)

	b := compose(msg)

	w.Write(b)
}

// listDatabase lists the supervised databases in a JSON format
func listDatabases(w http.ResponseWriter, r *http.Request) {
	var msg ListMessage

	msg.Status = http.StatusOK
	msg.Message = db.listDatabase()

	b := compose(msg)

	writeHeader(w, msg.Status)

	w.Write(b)
}

// getDatabase will get a specific database with a specific name
func getDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to getDatabase")
}

// dropDatabase will drop the named database with its tablespace and user
func dropDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to dropDatabase")
}

// importDatabase will import the specified dumpfile to the database
// creating the database, tablespace and user as needed
func importDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to importDatabase")
}

func whoami(w http.ResponseWriter, r *http.Request) {

	info := make(map[string]string)

	info["vendor"] = conf.Vendor
	info["version"] = conf.Version

	// TODO add other information if needed
	var msg MapMessage

	msg.Status = http.StatusOK
	msg.Message = info

	b := compose(msg)

	writeHeader(w, msg.Status)

	w.Write(b)
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	var msg Message

	// TODO: db.Ping() always returns true for some reason. Need to check why

	err := db.Ping()
	if err != nil {
		msg.Status = http.StatusServiceUnavailable
		msg.Message = err.Error()
	} else {
		msg.Status = http.StatusOK
		msg.Message = "Still alive"
	}

	b := compose(msg)

	writeHeader(w, msg.Status)

	w.Write(b)
}
