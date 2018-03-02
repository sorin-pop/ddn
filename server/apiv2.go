package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/djavorszky/ddn/common/errs"
	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	vis "github.com/djavorszky/ddn/common/visibility"
	"github.com/djavorszky/ddn/server/brwsr"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/djavorszky/liferay"
	"github.com/gorilla/mux"
)

func apiSetLogLevel(w http.ResponseWriter, r *http.Request) {
	_, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	var lvl logger.LogLevel

	level := mux.Vars(r)["level"]
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
		inet.SendFailure(w, http.StatusBadRequest, errs.UnknownParameter, level)
		return
	}

	if logger.Level == lvl {
		inet.SendSuccess(w, http.StatusOK, "Loglevel already at "+level)
		return
	}

	logger.Info("Changing loglevel: %s->%s", logger.Level, lvl)

	msg := fmt.Sprintf("Loglevel changed from %s to %s", logger.Level, lvl)

	logger.Level = lvl

	inet.SendSuccess(w, http.StatusOK, msg)
	return
}

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

func getAPIActiveAgents(w http.ResponseWriter, r *http.Request) {
	_, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	result := make([]model.Agent, 0)

	agents := registry.List()
	for _, agent := range agents {
		if !agent.Up {
			continue
		}

		result = append(result, agent)
	}

	if len(result) == 0 {
		inet.SendFailure(w, http.StatusNotFound, errs.NoAgentsAvailable)
		return
	}

	inet.SendSuccess(w, http.StatusOK, result)
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
	meta, errr := getDatabaseByIDFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
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
	meta, errr := getDatabaseByAgentDBNameFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
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
	meta, errr := getDatabaseByIDFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	agent, ok := registry.Get(meta.AgentName)
	if !ok {
		inet.SendFailure(w, http.StatusForbidden, errs.AgentNotFound)
		return
	}

	meta.Status = status.DropInProgress

	db.Update(&meta)

	go dropAsync(agent, meta.ID, meta.DBName, meta.DBUser)

	inet.SendSuccess(w, http.StatusOK, "Started dropping database")
}

func dropAPIDatabaseByAgentDBName(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	agentName, dbname := vars["agent"], vars["dbname"]

	// Get private ones
	meta, err := db.FetchByDBNameAgent(dbname, agentName)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.QueryFailed, err.Error())

		logger.Error("Fetching database failed: %v", err)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	agent, ok := registry.Get(meta.AgentName)
	if !ok {
		inet.SendFailure(w, http.StatusForbidden, errs.AgentNotFound)
		return
	}

	meta.Status = status.DropInProgress

	db.Update(&meta)

	go dropAsync(agent, meta.ID, meta.DBName, meta.DBUser)

	inet.SendSuccess(w, http.StatusOK, "Started dropping database")
}

func importAPIDB(w http.ResponseWriter, r *http.Request) {
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

	if req.DumpLocation == "" {
		inet.SendFailure(w, http.StatusBadRequest, errs.MissingParameters, "dumpfile_location")
		return
	}

	agent, ok := registry.Get(req.AgentIdentifier)
	if !ok {
		inet.SendFailure(w, http.StatusBadRequest, errs.AgentNotFound, req.AgentIdentifier)

		return
	}

	ensureValues(&req.DatabaseName, &req.Username, &req.Password, agent.DBVendor)

	dbe := data.Row{
		DBName:     req.DatabaseName,
		DBUser:     req.Username,
		DBPass:     req.Password,
		DBSID:      agent.DBSID,
		AgentName:  req.AgentIdentifier,
		Dumpfile:   req.DumpLocation,
		Creator:    user,
		CreateDate: time.Now(),
		ExpiryDate: time.Now().AddDate(0, 1, 0),
		DBAddress:  agent.DBAddr,
		DBPort:     agent.DBPort,
		DBVendor:   agent.DBVendor,
		Status:     status.Accepted,
	}

	err = db.Insert(&dbe)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.PersistFailed, err.Error())

		logger.Error("failed inserting database: %v", err)
		db.Delete(dbe)
		return
	}

	if strings.HasPrefix(dbe.Dumpfile, "/") && config.MountLoc == "" {
		inet.SendFailure(w, http.StatusBadRequest, errs.NoFoldersMounted)
		db.Delete(dbe)
		return
	}

	go startImport(agent, dbe)

	inet.SendSuccess(w, http.StatusAccepted, dbe)
}

func startImport(agent model.Agent, dbe data.Row) {
	defer db.Update(&dbe)

	url := dbe.Dumpfile
	if strings.HasPrefix(dbe.Dumpfile, "/") {
		_, filename := filepath.Split(dbe.Dumpfile)
		logger.Debug("Starting to copy %s to %s/web/dumps/", filename, workdir)
		dst, err := os.OpenFile(fmt.Sprintf("%s/web/dumps/%s", workdir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			errMsg := fmt.Sprintf("Failed creating downloadable file at web/dumps: %v", err)

			logger.Error(errMsg)
			dbe.Status = status.ImportFailed
			dbe.Message = errMsg
			return
		}
		defer dst.Close()

		src, err := os.Open(filepath.Join(config.MountLoc, dbe.Dumpfile))
		if err != nil {
			errMsg := fmt.Sprintf("Failed opening dumpfile at %v: %v", dbe.Dumpfile, err)

			logger.Error(errMsg)
			dbe.Status = status.ImportFailed
			dbe.Message = errMsg
			return
		}
		defer src.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			errMsg := fmt.Sprintf("Failed copying dumpfile %s -> %s: %v", src.Name(), dst.Name(), err)

			logger.Error(errMsg)
			dbe.Status = status.ImportFailed
			dbe.Message = errMsg
			return
		}

		logger.Debug("Copy successful, starting import")

		url = fmt.Sprintf("http://%s:%s/dumps/%s", config.ServerHost, config.ServerPort, filename)
	}

	_, err := agent.ImportDatabase(dbe.ID, dbe.DBName, dbe.DBUser, dbe.DBPass, url)
	if err != nil {
		errMsg := fmt.Sprintf("Import failed: %v", err)

		logger.Error(errMsg)

		dbe.Status = status.ImportFailed
		dbe.Message = errMsg
		return
	}

	dbe.Status = status.ImportInProgress
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
		inet.SendFailure(w, http.StatusInternalServerError, errs.PersistFailed, err.Error())

		logger.Error("failed inserting database: %v", err)
		return
	}

	_, err = agent.CreateDatabase(req.ID, req.DatabaseName, req.Username, req.Password)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.CreateFailed, err.Error())

		db.Delete(dbe)
		return
	}

	inet.SendSuccess(w, http.StatusOK, dbe)
}

func exportAPIDB(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	meta, errr := getDatabaseByIDFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	agent, ok := registry.Get(meta.AgentName)
	if !ok {
		inet.SendFailure(w, http.StatusInternalServerError, errs.AgentNotFound, meta.AgentName)
		return
	}

	resp, err := agent.ExportDatabase(meta.ID, meta.DBName, meta.DBUser, meta.DBPass)
	if err != nil {
		meta.Status = status.ExportFailed
		db.Update(&meta)

		inet.SendFailure(w, http.StatusInternalServerError, errs.ExportFailed)
		return
	}

	logger.Info("Um ok done I guess")

	inet.SendSuccess(w, http.StatusOK, resp)
}

func recreateAPIDB(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	meta, errr := getDatabaseByIDFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	agent, ok := registry.Get(meta.AgentName)
	if !ok {
		inet.SendFailure(w, http.StatusInternalServerError, errs.AgentNotFound, meta.AgentName)
		return
	}

	_, err = agent.DropDatabase(meta.ID, meta.DBName, meta.DBUser)
	if err != nil {
		meta.Status = status.DropDatabaseFailed
		db.Update(&meta)

		inet.SendFailure(w, http.StatusInternalServerError, errs.DropFailed)
		return
	}

	_, err = agent.CreateDatabase(meta.ID, meta.DBName, meta.DBUser, meta.DBPass)
	if err != nil {
		meta.Status = status.CreateDatabaseFailed
		db.Update(&meta)

		inet.SendFailure(w, http.StatusInternalServerError, errs.CreateFailed)
		return
	}

	inet.SendSuccess(w, http.StatusOK, meta)
}

func browseAPI(w http.ResponseWriter, r *http.Request) {
	_, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	if config.MountLoc == "" {
		inet.SendFailure(w, http.StatusFailedDependency, errs.NoFoldersMounted)
		return
	}

	vars := mux.Vars(r)
	loc, ok := vars["loc"]
	if !ok {
		loc = "/"
	}

	files, err := brwsr.List(loc)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.FailedListingDirectory, err.Error())

		logger.Error("failed listing folder: %v", err)
		return
	}

	inet.SendSuccess(w, http.StatusOK, files)
}

func apiSetVisibility(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	meta, errr := getDatabaseByIDFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	visibility := vars["visibility"]

	var visibilityNum int
	switch visibility {
	case "public":
		visibilityNum = vis.Public
	case "private":
		visibilityNum = vis.Private
	default:
		inet.SendFailure(w, http.StatusBadRequest, errs.MissingParameters, visibility)
		return
	}

	// If no change needed
	if visibilityNum == meta.Public {
		inet.SendSuccess(w, http.StatusOK, "Visibility already set to "+visibility)
		return
	}

	meta.Public = visibilityNum

	err = db.Update(&meta)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.UpdateFailed, err.Error())

		logger.Error("failed listing folder: %v", err)
		return
	}

	inet.SendSuccess(w, http.StatusOK, "Visibility updated successfully")
}

func apiExtendExpiry(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	meta, errr := getDatabaseByIDFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	amount, err := strconv.Atoi(vars["amount"])
	if err != nil {
		inet.SendFailure(w, http.StatusBadRequest, errs.InvalidURL, err.Error())

		logger.Error("Failed converting 'amount' to integer from URL: %s, %v", r.URL, err)
		return
	}

	var newExpiry time.Time
	switch vars["unit"] {
	case "days":
		newExpiry = meta.ExpiryDate.AddDate(0, 0, amount)
	case "months":
		newExpiry = meta.ExpiryDate.AddDate(0, amount, 0)
	case "year":
		newExpiry = meta.ExpiryDate.AddDate(amount, 0, 0)
	default:
		inet.SendFailure(w, http.StatusBadRequest, errs.UnknownParameter, vars["unit"])
		return
	}

	meta.ExpiryDate = newExpiry

	err = db.Update(&meta)
	if err != nil {
		inet.SendFailure(w, http.StatusInternalServerError, errs.UpdateFailed, err.Error())

		logger.Error("failed listing folder: %v", err)
		return
	}

	inet.SendSuccess(w, http.StatusOK, meta.ExpiryDate)
}

func apiAccessInfoByAgentDB(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	meta, errr := getDatabaseByAgentDBNameFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	inet.SendSuccess(w, http.StatusOK, getDBAccess(meta))
}

func apiAccessInfoByID(w http.ResponseWriter, r *http.Request) {
	user, err := getAPIUser(r)
	if err != nil {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	vars := mux.Vars(r)
	meta, errr := getDatabaseByIDFrom(vars)
	if errr.httpStatus != 0 {
		inet.SendFailure(w, errr.httpStatus, errr.errors...)
		return
	}

	if !hasAccess(meta, user) {
		inet.SendFailure(w, http.StatusForbidden, errs.AccessDenied)
		return
	}

	inet.SendSuccess(w, http.StatusOK, getDBAccess(meta))
}

type dbAccess struct {
	JDBCDriver  string `json:"jdbc_driver"`
	JDBCUrl     string `json:"jdbc_url"`
	JDBCUrl6210 string `json:"jdbc_url_6210,omitempty"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Database    string `json:"database, omitempty"`
	URL         string `json:"url"`
}

func getDBAccess(meta data.Row) dbAccess {
	var (
		jdbc     liferay.JDBC
		jdbc6210 liferay.JDBC
	)

	switch meta.DBVendor {
	case "mysql":
		jdbc = liferay.MysqlJDBCDXP(meta.DBAddress, meta.DBPort, meta.DBName, meta.DBUser, meta.DBPass)
		jdbc6210 = liferay.MysqlJDBC(meta.DBAddress, meta.DBPort, meta.DBName, meta.DBUser, meta.DBPass)
	case "mariadb":
		jdbc = liferay.MariaDBJDBC(meta.DBAddress, meta.DBPort, meta.DBName, meta.DBUser, meta.DBPass)
	case "postgres":
		jdbc = liferay.PostgreJDBC(meta.DBAddress, meta.DBPort, meta.DBName, meta.DBUser, meta.DBPass)
	case "oracle":
		jdbc = liferay.OracleJDBC(meta.DBAddress, meta.DBPort, meta.DBSID, meta.DBUser, meta.DBPass)
	case "mssql":
		jdbc = liferay.MSSQLJDBC(meta.DBAddress, meta.DBPort, meta.DBName, meta.DBUser, meta.DBPass)
	}

	dba := dbAccess{
		JDBCDriver: jdbc.Driver,
		JDBCUrl:    jdbc.URL,
		User:       meta.DBUser,
		Password:   meta.DBPass,
		URL:        meta.DBAddress + ":" + meta.DBPort,
	}

	if jdbc6210.URL != "" {
		dba.JDBCUrl6210 = jdbc6210.URL
	}

	if meta.DBVendor != "oracle" {
		dba.Database = meta.DBName
	}

	return dba
}

func getAPIUser(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("unauthorized request")
	}

	return auth, nil
}

func hasResult(meta data.Row) bool {
	if meta.Creator == "" {
		return false
	}
	return true
}

func hasAccess(meta data.Row, user string) bool {
	if meta.Public == vis.Private && meta.Creator != user {
		return false
	}
	return true
}

type errResult struct {
	httpStatus int
	errors     []string
}

func getDatabaseByIDFrom(vars map[string]string) (data.Row, errResult) {
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return data.Row{}, errResult{
			httpStatus: http.StatusBadRequest,
			errors:     []string{errs.InvalidURL},
		}
	}

	meta, err := db.FetchByID(id)
	if err != nil {
		logger.Error("Fetching database failed: %v", err)

		return data.Row{}, errResult{
			httpStatus: http.StatusInternalServerError,
			errors:     []string{errs.QueryFailed, err.Error()},
		}
	}

	if !hasResult(meta) {
		return data.Row{}, errResult{
			httpStatus: http.StatusNotFound,
			errors:     []string{errs.QueryNoResults},
		}
	}

	return meta, errResult{}
}

func getDatabaseByAgentDBNameFrom(vars map[string]string) (data.Row, errResult) {
	agent, dbname := vars["agent"], vars["dbname"]
	meta, err := db.FetchByDBNameAgent(dbname, agent)
	if err != nil {
		logger.Error("Fetching database failed: %v", err)

		return data.Row{}, errResult{
			httpStatus: http.StatusInternalServerError,
			errors:     []string{errs.QueryFailed, err.Error()},
		}
	}

	if !hasResult(meta) {
		return data.Row{}, errResult{
			httpStatus: http.StatusNotFound,
			errors:     []string{errs.QueryNoResults},
		}
	}

	return meta, errResult{}
}

/*
	func method(w http.ResponseWriter, r *http.Request) {}
*/
