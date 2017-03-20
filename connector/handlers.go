package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
)

// index should display whenever someone visits the main page.
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the index!")
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
	var (
		dbreq model.DBRequest
		msg   inet.Message
	)

	err := json.NewDecoder(r.Body).Decode(&dbreq)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	if ok := sutils.Present(db.RequiredFields(dbreq, createDB)...); !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.InvalidResponse())
		return
	}

	httpStatus := http.StatusOK
	err = db.CreateDatabase(dbreq)
	if err != nil {
		httpStatus = http.StatusInternalServerError
		msg.Status = status.CreateDatabaseFailed
		msg.Message = fmt.Sprintf("creating database failed: %s", err.Error())
	} else {
		msg.Status = status.Success
		msg.Message = "Successfully created the database and user!"
	}

	inet.SendResponse(w, httpStatus, msg)
}

// listDatabase lists the supervised databases in a JSON format
func listDatabases(w http.ResponseWriter, r *http.Request) {
	var (
		msg inet.ListMessage
		err error
	)

	msg.Status = status.Success
	msg.Message, err = db.ListDatabase()
	if err != nil {
		errStr := fmt.Sprintf("listing databases failed: %s", err.Error())
		log.Printf(errStr)

		var errMsg inet.Message

		errMsg.Status = status.ListDatabaseFailed
		errMsg.Message = errStr

		inet.SendResponse(w, http.StatusInternalServerError, errMsg)
		return
	}

	inet.SendResponse(w, http.StatusOK, msg)
}

// echo echoes whatever it receives (as JSON) to the log.
func echo(w http.ResponseWriter, r *http.Request) {
	var msg notif.Msg

	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	log.Printf("%+v", msg)
}

// dropDatabase will drop the named database with its tablespace and user
func dropDatabase(w http.ResponseWriter, r *http.Request) {
	var (
		dbreq model.DBRequest
		msg   inet.Message
	)

	err := json.NewDecoder(r.Body).Decode(&dbreq)
	if err != nil {
		log.Printf("couldn't drop database: %s", err.Error())

		inet.SendResponse(w, http.StatusInternalServerError, inet.ErrorJSONResponse(err))
		return
	}

	if ok := sutils.Present(db.RequiredFields(dbreq, dropDB)...); !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.InvalidResponse())
		return
	}

	httpStatus := http.StatusOK

	err = db.DropDatabase(dbreq)
	if err != nil {
		httpStatus = http.StatusInternalServerError
		msg.Status = status.DropDatabaseFailed
		msg.Message = fmt.Sprintf("dropping database failed: %s", err.Error())
	} else {
		msg.Status = status.Success
		msg.Message = "Successfully dropped the database and user!"
	}

	inet.SendResponse(w, httpStatus, msg)
}

// importDatabase will import the specified dumpfile to the database
// creating the database, tablespace and user
func importDatabase(w http.ResponseWriter, r *http.Request) {
	var (
		dbreq model.DBRequest
		msg   inet.Message
	)

	err := json.NewDecoder(r.Body).Decode(&dbreq)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	if ok := sutils.Present(db.RequiredFields(dbreq, importDB)...); !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.InvalidResponse())
		return
	}

	if exists := inet.AddrExists(dbreq.DumpLocation); !exists {
		msg.Status = status.NotFound
		msg.Message = "Specified file doesn't exist or is not reachable."

		inet.SendResponse(w, http.StatusNotFound, msg)
		return
	}

	err = db.CreateDatabase(dbreq)
	if err != nil {
		msg.Status = status.CreateDatabaseFailed
		msg.Message = fmt.Sprintf("creating database failed: %s", err.Error())

		inet.SendResponse(w, http.StatusInternalServerError, msg)
		return
	}

	msg.Status = status.Accepted
	msg.Message = "Understood request, starting import process."

	inet.SendResponse(w, http.StatusOK, msg)

	go startImport(dbreq)
}

func whoami(w http.ResponseWriter, r *http.Request) {
	info := make(map[string]string)

	info["database-vendor"] = conf.Vendor
	info["database-version"] = conf.Version
	info["connector-version"] = version

	duration := time.Since(startup)

	// Round to milliseconds.
	info["connector-uptime"] = fmt.Sprintf("%s", duration-(duration%time.Millisecond))

	var msg inet.MapMessage

	msg.Status = status.Success
	msg.Message = info

	inet.SendResponse(w, http.StatusOK, msg)
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	var msg inet.Message

	msg.Status = status.Success
	msg.Message = "Still alive"

	err := db.Alive()
	if err != nil {
		log.Printf("database dead: %s", err.Error())
		msg = inet.ErrorResponse()
	}

	inet.SendResponse(w, http.StatusOK, msg)
}
