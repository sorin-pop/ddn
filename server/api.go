package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/djavorszky/sutils"
	"github.com/gorilla/mux"
)

// apiList will list all available agents in a JSON format.
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
		logger.Error("couldn't decode json request: %v", err)

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	if ok := sutils.Present(req.AgentIdentifier, req.RequesterEmail); !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: fmt.Sprintf("Need values for 'agent_identifier' and 'requester_email', but got: %q and %q", req.AgentIdentifier, req.RequesterEmail)})
		return
	}

	var (
		conn model.Agent
		ok   bool
	)

	conn, ok = registry.Get(req.AgentIdentifier)
	if !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: fmt.Sprintf("Agent '%s' not found in registry", req.AgentIdentifier)})
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
		DBName:     req.DatabaseName,
		DBUser:     req.Username,
		DBPass:     req.Password,
		DBSID:      conn.DBSID,
		AgentName:  req.AgentIdentifier,
		Creator:    req.RequesterEmail,
		CreateDate: time.Now(),
		ExpiryDate: time.Now().AddDate(0, 1, 0),
		DBAddress:  conn.DBAddr,
		DBPort:     conn.DBPort,
		DBVendor:   conn.DBVendor,
		Status:     status.Success,
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

// apiAgentByName returns an agent by its shortname
func apiAgentByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	shortname := vars["shortname"]

	conn, ok := registry.Get(shortname)
	if !ok {
		msg := inet.Message{Status: http.StatusServiceUnavailable, Message: "ERR_AGENT_NOT_FOUND"}

		inet.SendResponse(w, http.StatusServiceUnavailable, msg)
		return
	}

	msg := inet.StructMessage{Status: http.StatusOK, Message: conn}

	inet.SendResponse(w, http.StatusOK, msg)
}

func apiSafe2Restart(w http.ResponseWriter, r *http.Request) {
	imports := make(map[string]int)

	// Check if server and agents are restartable
	entries, err := db.FetchAll()
	if err != nil {
		msg := inet.Message{Status: http.StatusInternalServerError, Message: "ERR_QUERY_FAILED"}

		inet.SendResponse(w, http.StatusInternalServerError, msg)
		return
	}

	for _, entry := range entries {
		if entry.InProgress() {
			imports[entry.AgentName]++
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

func apiSetLogLevel(w http.ResponseWriter, r *http.Request) {
	var lvl logger.LogLevel

	vars := mux.Vars(r)
	level := vars["level"]

	switch strings.ToLower(level) {
	case "fatal":
		lvl = logger.FATAL
	case "error":
		lvl = logger.ERROR
	case "warn":
		lvl = logger.WARN
	case "info":
		lvl = logger.INFO
	case "debug":
		lvl = logger.DEBUG
	default:
		msg := inet.Message{Status: http.StatusBadRequest, Message: "ERR_UNRECOGNIZED_LOGLEVEL"}

		inet.SendResponse(w, http.StatusBadRequest, msg)
		return
	}

	if logger.Level == lvl {
		logger.Warn("Loglevel already at %s", lvl)

		msg := inet.Message{Status: http.StatusOK, Message: fmt.Sprintf("Loglevel already at %s", lvl)}

		inet.SendResponse(w, http.StatusOK, msg)
		return
	}

	logger.Info("Changing loglevel: %s->%s", logger.Level, lvl)

	msg := inet.Message{Status: http.StatusOK, Message: fmt.Sprintf("Loglevel changed from %s to %s", logger.Level, lvl)}

	logger.Level = lvl

	inet.SendResponse(w, http.StatusOK, msg)
	return
}

// stores a web push notification subscription to the database
func apiSaveSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		subscription model.PushSubscription
		err          error
	)

	err = json.NewDecoder(r.Body).Decode(&subscription)
	if err != nil {
		logger.Error("couldn't decode json request: %v", err)

		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.InvalidJSON,
			Message: "There was a problem with the subscription request sent to the server. Server could not decode JSON request."})
		return
	}

	if ok := sutils.Present(subscription.Endpoint, subscription.Keys.P256dh, subscription.Keys.Auth); !ok {
		logger.Error("Missing or empty subscription parameters were received from the /api/save-subscription API call!")
		//TODO
		// log the received request body
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: "There was a problem with the subscription request sent to the server. Either \"endpoint\", \"p256dh\", or \"auth\" is missing, or empty."})
		return
	}

	userCookie, err := r.Cookie("user")
	if err != nil {
		logger.Error("getting user cookie failed: " + err.Error())
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.SaveSubscriptionFailed,
			Message: fmt.Sprintf("getting user cookie failed: %s", err.Error())})
		return
	}

	err = db.InsertPushSubscription(&subscription, userCookie.Value)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.SaveSubscriptionFailed,
			Message: fmt.Sprintf("persisting push subscription failed: %s", err.Error())})
		return
	}

	msg := inet.Message{Status: status.Success, Message: fmt.Sprintf("Subscription has been saved to back end.")}

	inet.SendResponse(w, http.StatusOK, msg)
}

// removes a web push notification subscription from the database
func apiRemoveSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		subscription model.PushSubscription
		err          error
	)

	err = json.NewDecoder(r.Body).Decode(&subscription)
	if err != nil {
		logger.Error("couldn't decode json request: %v", err)

		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.InvalidJSON,
			Message: "There was a problem with the subscription removal request sent to the server. Server could not decode JSON request."})
		return
	}

	if ok := sutils.Present(subscription.Endpoint, subscription.Keys.P256dh, subscription.Keys.Auth); !ok {
		logger.Error("Missing or empty subscription parameters were received from the /api/remove-subscription API call!")
		//TODO
		// log the received request body
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: "There was a problem with the subscription removal request sent to the server. Either \"endpoint\", \"p256dh\", or \"auth\" is missing, or empty."})
		return
	}

	userCookie, err := r.Cookie("user")
	if err != nil {
		logger.Error("getting user cookie failed: " + err.Error())
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.DeleteSubscriptionFailed,
			Message: fmt.Sprintf("getting user cookie failed: %s", err.Error())})
		return
	}

	err = db.DeletePushSubscription(&subscription, userCookie.Value)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  status.DeleteSubscriptionFailed,
			Message: fmt.Sprintf("deleting push subscription from database failed: %s", err.Error())})
		return
	}

	msg := inet.Message{Status: status.Success, Message: fmt.Sprintf("Subscription has been removed from back end.")}

	inet.SendResponse(w, http.StatusOK, msg)
}
