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
	"github.com/djavorszky/ddn/server/database"
	"github.com/djavorszky/sutils"
)

// apiList will list all available connectors in a JSON format.
func apiList(w http.ResponseWriter, r *http.Request) {
	list := make(map[string]string, 10)
	for _, con := range registry {
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
	if conn, ok = registry[req.ConnectorIdentifier]; !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: fmt.Sprintf("Connector '%s' not found in registry", req.ConnectorIdentifier)})
		return
	}

	if req.ID == 0 {
		req.ID = getID()
	}

	ensureValues(&req.DatabaseName, &req.Username, &req.Password)

	_, err = conn.CreateDatabase(req.ID, req.DatabaseName, req.Username, req.Password)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.CreateDatabaseFailed,
			Message: fmt.Sprintf("creating database failed: %s", err.Error())})
		return
	}

	dbe := database.Entry{
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

	err = database.Insert(&dbe)
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
