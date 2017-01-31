package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// index should display whenever someone visits the main page.
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the index!")
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var (
		dbreq DBRequest
		msg   Message
	)

	err := decoder.Decode(&dbreq)
	if err != nil {
		msg = errorJSONResponse(err)
		sendResponse(w, msg)

		return
	}

	if ok := present(db.RequiredFields(dbreq, createDB)...); !ok {
		msg = invalidResponse()
		sendResponse(w, msg)
		return
	}

	err = db.CreateDatabase(dbreq)
	if err != nil {
		msg.Status = http.StatusInternalServerError
		msg.Message = err.Error()
	} else {
		msg.Status = http.StatusOK
		msg.Message = "Successfully created the database and user!"
	}

	sendResponse(w, msg)
}

// listDatabase lists the supervised databases in a JSON format
func listDatabases(w http.ResponseWriter, r *http.Request) {
	var (
		msg ListMessage
		err error
	)

	msg.Status = http.StatusOK
	msg.Message, err = db.ListDatabase()
	if err != nil {
		sendResponse(w, errorResponse())

		return
	}

	sendResponse(w, msg)
}

// getDatabase will get a specific database with a specific name
func getDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to getDatabase")
}

// dropDatabase will drop the named database with its tablespace and user
func dropDatabase(w http.ResponseWriter, r *http.Request) {
	var (
		dbreq DBRequest
		msg   Message
	)

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&dbreq)
	if err != nil {
		msg = errorJSONResponse(err)
		sendResponse(w, msg)

		return
	}

	if ok := present(db.RequiredFields(dbreq, dropDB)...); !ok {
		msg := invalidResponse()
		sendResponse(w, msg)
		return
	}

	err = db.DropDatabase(dbreq)
	if err != nil {
		msg.Status = http.StatusInternalServerError
		msg.Message = err.Error()
	} else {
		msg.Status = http.StatusOK
		msg.Message = "Successfully dropped the database and user!"
	}

	sendResponse(w, msg)

}

// importDatabase will import the specified dumpfile to the database
// creating the database, tablespace and user
func importDatabase(w http.ResponseWriter, r *http.Request) {
	var (
		dbreq DBRequest
		msg   Message
	)

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&dbreq)
	if err != nil {
		msg = errorJSONResponse(err)
		sendResponse(w, msg)

		return
	}

	if ok := present(db.RequiredFields(dbreq, importDB)...); !ok {
		msg := invalidResponse()
		sendResponse(w, msg)
		return
	}

	if exists := fileExists(dbreq.DumpLocation); exists == false {
		msg.Status = http.StatusNotFound
		msg.Message = "Specified file doesn't exist or is not reachable."

		sendResponse(w, msg)
		return
	}

	err = db.CreateDatabase(dbreq)
	if err != nil {
		msg.Status = http.StatusInternalServerError
		msg.Message = err.Error()

		sendResponse(w, msg)
		return
	}

	msg.Status = http.StatusOK
	msg.Message = "Understood request, starting import process."

	sendResponse(w, msg)

	go startImport(dbreq)
}

func whoami(w http.ResponseWriter, r *http.Request) {

	info := make(map[string]string)

	info["vendor"] = conf.Vendor
	info["version"] = conf.Version

	// TODO add other information if needed
	var msg MapMessage

	msg.Status = http.StatusOK
	msg.Message = info

	sendResponse(w, msg)
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	var msg Message

	msg.Status = http.StatusOK
	msg.Message = "Still alive"

	err := db.Alive()
	if err != nil {
		msg = errorResponse()
	}

	sendResponse(w, msg)
}

func sendResponse(w http.ResponseWriter, msg JSONMessage) {
	b, status := msg.Compose()

	writeHeader(w, status)

	w.Write(b)
}
