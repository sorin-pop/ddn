package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/djavorszky/ddn/common/errs"
	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/gorilla/mux"
)

/*
	GET /api/agents 						-> lists agents - done
	GET /api/agents/${agent-name:string} 	-> gets all info of specific agent - done

	GET /api/database 									-> lists databases - done
	GET /api/database/${id:int} 						-> gets all info of a specific database
	GET /api/database/${agent:string}/${dbname:string} 	-> gets all info of a specific database

	POST /api/database	-> Creates or imports a new database (json body)

	DELETE /api/database/${id:int} 							-> drops a database
	DELETE /api/database/${agent:string}/${dbname:string} 	-> drops a database
*/

func getAPIAgents(w http.ResponseWriter, r *http.Request) {
	_, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	list := make(map[string]model.Agent, 10)
	for _, agent := range registry.List() {
		list[agent.ShortName] = agent
	}

	inet.SendSuccess(w, http.StatusOK, list)
}

// apiAgentByName returns an agent by its shortname
func getAPIAgentByName(w http.ResponseWriter, r *http.Request) {
	_, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)

	shortname := vars["shortname"]

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
	dbs, err := db.FetchByCreator(user)
	if err != nil {
		if err != nil {
			inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed)

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
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed)

		logger.Error("Fetching public dbs failed: %v", err)
		return
	}

	for _, db := range dbs {
		databases[db.ID] = db
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
		inet.SendFailure(w, http.StatusBadRequest, errs.InvalidURL)

		logger.Error("Failed converting id to integer from URL: %s, %v", r.URL, err)
		return
	}

	// Get private ones
	db, err := db.FetchByID(id)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed)

		logger.Error("Fetching database failed: %v", err)
		return

	}

	if db.Creator == "" {
		inet.SendFailure(w, http.StatusNotFound, errs.QueryNoResults)
		return
	}

	if db.Creator != user {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	inet.SendSuccess(w, http.StatusOK, db)
}

func getAPIDatabaseByAgentDBName(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	agent, dbname := vars["agent"], vars["dbname"]

	// Get private ones
	db, err := db.FetchByDBNameAgent(agent, dbname)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed)

		logger.Error("Fetching database failed: %v", err)
		return

	}

	if db.Creator == "" {
		inet.SendFailure(w, http.StatusNotFound, errs.QueryNoResults)
		return
	}

	if db.Creator != user {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	inet.SendSuccess(w, http.StatusOK, db)
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
