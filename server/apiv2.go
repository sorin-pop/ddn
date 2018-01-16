package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/djavorszky/ddn/common/errs"
	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/gorilla/mux"
)

/*
	GET /api/agents 						-> lists agents - done
	GET /api/agents/${agent-name:string} 	-> gets all info of specific agent - done

	GET /api/database 									-> lists databases - done
	GET /api/database/${id:int} 						-> gets all info of a specific database - done
	GET /api/database/${agent:string}/${dbname:string} 	-> gets all info of a specific database - done

	POST /api/database	-> Creates or imports a new database (json body) - create done

	DELETE /api/database/${id:int} 							-> drops a database - done
	DELETE /api/database/${agent:string}/${dbname:string} 	-> drops a database - done
*/

func getAPIAgents(w http.ResponseWriter, r *http.Request) {
	_, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	agents := registry.List()

	if len(agents) == 0 {
		inet.SendFailure(w, http.StatusNotFound, errs.NoAgentsAvailable)
		return
	}

	inet.SendSuccess(w, http.StatusOK, agents)
}

// apiAgentByName returns an agent by its shortname
func getAPIAgentByName(w http.ResponseWriter, r *http.Request) {
	_, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)

	shortname := vars["agent"]

	agent, ok := registry.Get(shortname)
	if !ok {
		inet.SendFailure(w, http.StatusServiceUnavailable, errs.AgentNotFound)
		return
	}

	inet.SendSuccess(w, http.StatusOK, agent)
}

func getAPIDatabases(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	// Get private ones
	metas, err := db.FetchByCreator(user)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed, err.Error())

		logger.Error("Fetching private dbs failed: %v", err)
		return
	}

	databases := make([]data.Row, 0, len(metas))

	for _, meta := range metas {
		databases = append(databases, meta)
	}

	// Get public ones
	metas, err = db.FetchPublic()
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed, err.Error())

		logger.Error("Fetching public dbs failed: %v", err)
		return
	}

	for _, meta := range metas {
		databases = append(databases, meta)
	}

	inet.SendSuccess(w, http.StatusOK, databases)
}

func getAPIDatabaseByID(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		inet.SendFailure(w, http.StatusBadRequest, errs.InvalidURL, err.Error())

		logger.Error("Failed converting id to integer from URL: %s, %v", r.URL, err)
		return
	}

	meta, err := db.FetchByID(id)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed)

		logger.Error("Fetching database failed: %v", err)
		return

	}

	if meta.Creator == "" {
		inet.SendFailure(w, http.StatusNotFound, errs.QueryNoResults)
		return
	}

	if meta.Creator != user {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	inet.SendSuccess(w, http.StatusOK, meta)
}

func getAPIDatabaseByAgentDBName(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	agent, dbname := vars["agent"], vars["dbname"]

	meta, err := db.FetchByDBNameAgent(dbname, agent)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed, err.Error())

		logger.Error("Fetching database failed: %v", err)
		return

	}

	if meta.Creator == "" {
		inet.SendFailure(w, http.StatusNotFound, errs.QueryNoResults)
		return
	}

	if meta.Creator != user {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	inet.SendSuccess(w, http.StatusOK, meta)
}

func dropAPIDatabaseByID(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		inet.SendFailure(w, http.StatusBadRequest, errs.InvalidURL, err.Error())

		logger.Error("Failed converting id to integer from URL: %s, %v", r.URL, err)
		return
	}

	meta, err := db.FetchByID(id)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed, err.Error())

		logger.Error("Fetching database failed: %v", err)
		return
	}

	if meta.Creator == "" {
		inet.SendFailure(w, http.StatusNotFound, errs.QueryNoResults)
		return
	}

	if meta.Creator != user {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	err = db.Delete(meta)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.DeleteFailed)
		return
	}

	inet.SendSuccess(w, http.StatusOK, "Delete successful")
}

func dropAPIDatabaseByAgentDBName(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	agent, dbname := vars["agent"], vars["dbname"]

	// Get private ones
	meta, err := db.FetchByDBNameAgent(dbname, agent)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed, err.Error())

		logger.Error("Fetching database failed: %v", err)
		return
	}

	if meta.Creator == "" {
		inet.SendFailure(w, http.StatusNotFound, errs.QueryNoResults)
		return
	}

	if meta.Creator != user {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	err = db.Delete(meta)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.DeleteFailed)
		return
	}

	inet.SendSuccess(w, http.StatusOK, "Delete successful")
}

func createAPIDB(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	var req model.ClientRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		inet.SendFailure(w, http.StatusBadRequest, errs.InvalidURL, err.Error())

		logger.Error("couldn't decode json request: %v", err)
		return
	}

	if req.AgentIdentifier == "" {
		inet.SendFailure(w, http.StatusBadRequest, errs.MissingParameters, "agent_identifier")

		return
	}

	agent, ok := registry.Get(req.AgentIdentifier)
	if !ok {
		inet.SendFailure(w, http.StatusBadRequest, errs.AgentNotFound, req.AgentIdentifier)

		return
	}

	ensureValues(&req.DatabaseName, &req.Username, &req.Password, agent.DBVendor)

	req.ID = registry.ID()
	dbe := data.Row{
		DBName:     req.DatabaseName,
		DBUser:     req.Username,
		DBPass:     req.Password,
		DBSID:      agent.DBSID,
		AgentName:  req.AgentIdentifier,
		Creator:    user,
		CreateDate: time.Now(),
		ExpiryDate: time.Now().AddDate(0, 1, 0),
		DBAddress:  agent.DBAddr,
		DBPort:     agent.DBPort,
		DBVendor:   agent.DBVendor,
		Status:     status.Success,
	}

	err = db.Insert(&dbe)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.PersistFailed, err)

		logger.Error("failed inserting database: %v", err)
		return
	}

	_, err = agent.CreateDatabase(req.ID, req.DatabaseName, req.Username, req.Password)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.CreateFailed, err)

		db.Delete(dbe)
		return
	}

	inet.SendSuccess(w, http.StatusOK, dbe)
}

func getAPIUser(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("unauthorized request")
	}

	return auth, nil
}

/*
	func method(w http.ResponseWriter, r *http.Request) {}
*/
