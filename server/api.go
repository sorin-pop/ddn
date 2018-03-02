package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/djavorszky/ddn/common/errs"
	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/ddn/common/visibility"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/djavorszky/liferay"
	"github.com/djavorszky/sutils"
	"github.com/gorilla/mux"
)

func apiPage(w http.ResponseWriter, r *http.Request) {
	api := fmt.Sprintf("%s/web/html/api.html", workdir)

	t, err := template.ParseFiles(api)
	if err != nil {
		logger.Error("Failed parsing files: %v", err)
		return
	}

	t.Execute(w, nil)
}

// apiList will list all available agents in a shortname:dbvendor mapping format.
func apiList(w http.ResponseWriter, r *http.Request) {
	list := make(map[string]string, 10)
	for _, agent := range registry.List() {
		list[agent.ShortName] = agent.DBVendor
	}

	msg := inet.MapMessage{Status: status.Success, Message: list}

	inet.SendResponse(w, http.StatusOK, msg)
}

// apiListAgents will list all available agents in a JSON format.
func apiListAgents(w http.ResponseWriter, r *http.Request) {
	list := make(map[string]model.Agent, 10)
	for _, agent := range registry.List() {
		list[agent.ShortName] = agent
	}

	msg := inet.StructMessage{Status: status.Success, Message: list}

	inet.SendResponse(w, http.StatusOK, msg)
}

// apiListDatabases will list all databases
func apiListDatabases(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendResponse(w, http.StatusForbidden, inet.Message{
			Status:  http.StatusForbidden,
			Message: errs.AccessDenied,
		})
		return
	}

	// Get private ones
	dbs, err := db.FetchByCreator(user)
	if err != nil {
		if err != nil {
			inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
				Status:  http.StatusInternalServerError,
				Message: errs.QueryFailed,
			})

			logger.Error("Fetching private dbs failed: %v", err)
			return
		}
	}

	databases := make(map[int]data.Row)

	for _, db := range dbs {
		databases[db.ID] = db
	}

	// Get public ones
	dbs, err = db.FetchPublic()
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.QueryFailed,
		})

		logger.Error("Fetching public dbs failed: %v", err)
		return
	}

	for _, db := range dbs {
		databases[db.ID] = db
	}

	msg := inet.StructMessage{Status: http.StatusOK, Message: databases}

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

		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.JSONDecodeFailed,
		})
		return
	}

	if ok := sutils.Present(req.AgentIdentifier, req.RequesterEmail); !ok {
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingParameters,
		})
		return
	}

	agent, ok := registry.Get(req.AgentIdentifier)
	if !ok {
		logger.Error("Agent %q not found", req.AgentIdentifier)
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  status.MissingParameters,
			Message: errs.AgentNotFound,
		})
		return
	}

	if req.DatabaseName == "" && req.Username != "" {
		req.DatabaseName = req.Username
	}

	ensureValues(&req.DatabaseName, &req.Username, &req.Password, agent.DBVendor)

	req.ID = registry.ID()

	dbe := data.Row{
		DBName:     req.DatabaseName,
		DBUser:     req.Username,
		DBPass:     req.Password,
		DBSID:      agent.DBSID,
		AgentName:  req.AgentIdentifier,
		Creator:    req.RequesterEmail,
		CreateDate: time.Now(),
		ExpiryDate: time.Now().AddDate(0, 1, 0),
		DBAddress:  agent.DBAddr,
		DBPort:     agent.DBPort,
		DBVendor:   agent.DBVendor,
		Status:     status.Success,
	}

	err = db.Insert(&dbe)
	if err != nil {
		logger.Error("failed inserting database: %v", err)

		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.PersistFailed,
		})
		return
	}

	_, err = agent.CreateDatabase(req.ID, req.DatabaseName, req.Username, req.Password)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("creating database failed: %v", err),
		})

		db.Delete(dbe)
		return
	}

	resp, err := json.Marshal(dbe)
	if err != nil {
		logger.Error("json marshal failed: %v", err)
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.JSONEncodeFailed,
		})
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
		msg := inet.Message{
			Status:  http.StatusServiceUnavailable,
			Message: errs.AgentNotFound,
		}

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
		logger.Error("failed FetchAll: %v", err)
		msg := inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.QueryFailed,
		}

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
			Status:  http.StatusBadRequest,
			Message: errs.JSONDecodeFailed,
		})
		return
	}

	if ok := sutils.Present(subscription.Endpoint, subscription.Keys.P256dh, subscription.Keys.Auth); !ok {
		logger.Error("Missing or empty subscription parameters were received from the /api/save-subscription API call!")
		//TODO
		// log the received request body
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingParameters,
		})
		return
	}

	userCookie, err := r.Cookie("user")
	if err != nil {
		logger.Error("getting user cookie failed: %v", err)
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingUserCookie,
		})
		return
	}

	err = db.InsertPushSubscription(&subscription, userCookie.Value)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.PersistFailed,
		})
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
			Status:  http.StatusBadRequest,
			Message: errs.JSONDecodeFailed,
		})
		return
	}

	if ok := sutils.Present(subscription.Endpoint, subscription.Keys.P256dh, subscription.Keys.Auth); !ok {
		logger.Error("Missing or empty subscription parameters were received from the /api/remove-subscription API call!")
		//TODO
		// log the received request body
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingParameters,
		})
		return
	}

	userCookie, err := r.Cookie("user")
	if err != nil {
		logger.Error("getting user cookie failed: " + err.Error())
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingUserCookie,
		})
		return
	}

	err = db.DeletePushSubscription(&subscription, userCookie.Value)
	if err != nil {
		logger.Error("failed deleting push subscription: %v", err)

		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.DropFailed,
		})
		return
	}

	msg := inet.Message{Status: status.Success, Message: fmt.Sprintf("Subscription has been removed from back end.")}

	inet.SendResponse(w, http.StatusOK, msg)
}

// apiDBAccess will list useful information for connecting to the specified database (JDBC driver, url, etc.)
func apiDBAccess(w http.ResponseWriter, r *http.Request) {
	var jdbcParams liferay.JDBC

	list := make(map[string]string, 5)
	vars := mux.Vars(r)
	requester, dbname, agent := vars["requester"], vars["dbname"], vars["agent"]

	if ok := sutils.Present(requester, dbname, agent); !ok {
		logger.Error("Missing parameters: requester: %q, dbname: %q, agent: %q", requester, dbname, agent)
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingParameters,
		})
		return
	}

	dbe, err := db.FetchByDBNameAgent(dbname, agent)
	if err != nil {
		logger.Error("FetchByAgentDBName: %v", err)
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.QueryFailed,
		})
		return
	}

	if dbe.Public == vis.Private && dbe.Creator != requester {
		logger.Error("User %q tried to get portalext of db created by %q.", requester, dbe.Creator)
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusForbidden,
			Message: errs.AccessDenied})
		return
	}

	switch dbe.DBVendor {
	case "mysql":
		jdbcParams = liferay.MysqlJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
	case "mariadb":
		jdbcParams = liferay.MariaDBJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
	case "postgres":
		jdbcParams = liferay.PostgreJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
	case "oracle":
		jdbcParams = liferay.OracleJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBSID, dbe.DBUser, dbe.DBPass)
	case "mssql":
		jdbcParams = liferay.MSSQLJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
	}

	list["jdbc-driver"] = jdbcParams.Driver
	list["jdbc-url"] = jdbcParams.URL
	list["user"] = dbe.DBUser
	list["password"] = dbe.DBPass
	list["url"] = dbe.DBAddress + ":" + dbe.DBPort

	msg := inet.MapMessage{Status: status.Success, Message: list}

	inet.SendResponse(w, http.StatusOK, msg)
}
