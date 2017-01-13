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

	valid, msg := validDBReq(dbreq.DatabaseName, dbreq.Username, dbreq.Password)
	if !valid {
		sendResponse(w, msg)
		return
	}

	err = db.createDatabase(dbreq)
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
	msg.Message, err = db.listDatabase()
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

	valid, msg := validDBReq(dbreq.DatabaseName, dbreq.Username)
	if !valid {
		sendResponse(w, msg)
		return
	}

	err = db.dropDatabase(dbreq)
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

func errorResponse() Message {
	var errMsg Message

	errMsg.Status = http.StatusServiceUnavailable
	errMsg.Message = "The server is unable to process requests as the underlying database is down."

	return errMsg
}

func errorJSONResponse(err error) Message {
	var msg Message

	log.Println("Could not decode JSON message:", err.Error())

	msg.Status = http.StatusBadRequest
	msg.Message = fmt.Sprintf("Invalid JSON request, received error: %s", err.Error())

	return msg
}

func validDBReq(reqFields ...string) (bool, Message) {
	var msg Message

	for _, field := range reqFields {
		if field == "" {
			msg.Status = http.StatusBadRequest
			msg.Message = "One or more required fields are missing from the call"

			return false, msg
		}
	}

	return true, msg
}
