package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/djavorszky/sutils"
	"github.com/gorilla/mux"
)

// apiList will list all available connectors in a JSON format.
func apiList(w http.ResponseWriter, r *http.Request) {
	list := make(map[string]string, 10)
	for _, con := range registry.List() {
		list[con.ShortName] = con.LongName
	}

	msg := inet.MapMessage{Status: status.Success, Message: list}

	inet.SendResponse(w, http.StatusOK, msg)
}

// apiCreate will create a database with the provided details
func apiCreate(w http.ResponseWriter, r *http.Request) {
	var (
		req model.ClientRequest
		err error
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	if ok := sutils.Present(req.ConnectorIdentifier, req.RequesterEmail); !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: fmt.Sprintf("Need values for 'connector_identifier' and 'requester_email', but got: %q and %q", req.ConnectorIdentifier, req.RequesterEmail)})
		return
	}

	var (
		conn model.Connector
		ok   bool
	)

	conn, ok = registry.Get(req.ConnectorIdentifier)
	if !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: fmt.Sprintf("Connector '%s' not found in registry", req.ConnectorIdentifier)})
		return
	}

	if req.ID == 0 {
		req.ID = registry.ID()
	}

	if req.DatabaseName == "" && req.Username != "" {
		req.DatabaseName = req.Username
	}

	ensureValues(&req.DatabaseName, &req.Username, &req.Password)

	_, err = conn.CreateDatabase(req.ID, req.DatabaseName, req.Username, req.Password)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.CreateDatabaseFailed,
			Message: fmt.Sprintf("creating database failed: %s", err.Error())})
		return
	}

	dbe := data.Row{
		DBName:        req.DatabaseName,
		DBUser:        req.Username,
		DBPass:        req.Password,
		DBSID:         conn.DBSID,
		ConnectorName: req.ConnectorIdentifier,
		Creator:       req.RequesterEmail,
		CreateDate:    time.Now(),
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
		DBAddress:     conn.DBAddr,
		DBPort:        conn.DBPort,
		DBVendor:      conn.DBVendor,
		Status:        status.Success,
	}

	err = db.Insert(&dbe)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.CreateDatabaseFailed,
			Message: fmt.Sprintf("persisting database locally failed: %s", err.Error())})
		return
	}

	resp, err := json.Marshal(dbe)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.CreateDatabaseFailed,
			Message: fmt.Sprintf("failed to marshal response: %s", err.Error())})
		return
	}

	inet.WriteHeader(w, http.StatusOK)

	w.Write(resp)
}

// apiConnectorByName returns a connector by its shortname
func apiConnectorByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	shortname := vars["shortname"]

	conn, ok := registry.Get(shortname)
	if !ok {
		msg := inet.Message{Status: http.StatusServiceUnavailable, Message: "ERR_CONNECTOR_NOT_FOUND"}

		inet.SendResponse(w, http.StatusServiceUnavailable, msg)
		return
	}

	msg := inet.StructMessage{Status: http.StatusOK, Message: conn}

	inet.SendResponse(w, http.StatusOK, msg)
}

func apiSafe2Restart(w http.ResponseWriter, r *http.Request) {
	imports := make(map[string]int)

	// Check if server and connectors are restartable
	entries, err := db.FetchAll()
	if err != nil {
		msg := inet.Message{Status: http.StatusInternalServerError, Message: "ERR_QUERY_FAILED"}

		inet.SendResponse(w, http.StatusInternalServerError, msg)
		return
	}

	for _, entry := range entries {
		if entry.InProgress() {
			imports[entry.ConnectorName]++
		}
	}

	result := inet.MapMessage{Status: http.StatusOK, Message: make(map[string]string)}

	conns := registry.List()

	if len(imports) == 0 {
		result.Message["server"] = "yes"

		for _, c := range conns {
			result.Message[c.ShortName] = "yes"
		}

		inet.SendResponse(w, http.StatusOK, result)
		return
	}

	result.Message["server"] = "no"

	for _, c := range conns {
		if imports[c.ShortName] == 0 {
			result.Message[c.ShortName] = "yes"
			continue
		}

		result.Message[c.ShortName] = fmt.Sprintf("No, %d imports running", imports[c.ShortName])
	}

	inet.SendResponse(w, http.StatusOK, result)
}
