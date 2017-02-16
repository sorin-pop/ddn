package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
)

func listConnectors(w http.ResponseWriter, r *http.Request) {
	list := make(map[string]string, 10)
	for _, con := range registry {
		list[con.ShortName] = con.LongName
	}

	msg := inet.MapMessage{Status: status.Success, Message: list}

	inet.SendResponse(w, http.StatusOK, msg)
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
	var req model.ClientRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	if req.DatabaseName == "" {
		req.DatabaseName = sutils.RandDBName()
	}

	if req.Username == "" {
		req.Username = sutils.RandUserName()
	}

	if req.Password == "" {
		req.Password = sutils.RandPassword()
	}

	con, ok := registry[req.ConnectorIdentifier]
	if !ok {
		e := fmt.Errorf("requested identifier %q not in registry", req.ConnectorIdentifier)
		log.Println(e.Error())

		inet.SendResponse(w, http.StatusNotFound, inet.ErrorResponse())
		return
	}

	dest := fmt.Sprintf("http://%s/create-database", con.Address)

	resp, err := notif.SndLoc(req, dest)
	if err != nil {
		log.Printf("couldn't create database on connector: %s", err.Error())

		inet.SendResponse(w, http.StatusInternalServerError, inet.ErrorResponse())
		return
	}

	var msg inet.Message
	respBytes := bytes.NewBufferString(resp)

	err = json.NewDecoder(respBytes).Decode(&msg)
	if err != nil {
		e := fmt.Errorf("malformed response from connector: %s", err.Error())
		log.Println(e.Error())

		inet.SendResponse(w, http.StatusInternalServerError, inet.ErrorJSONResponse(e))
		return
	}

	if msg.Status == status.Success {
		db.persist(req)
	}

	inet.SendResponse(w, http.StatusOK, msg)
}

func createDatabaseGET(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	connector, version := values.Get("db"), values.Get("lrversion")

	if connector == "" || version == "" {
		fmt.Fprintf(w, "Not enough parameters specified. Required: 'db', 'lrversion'")
		return
	}

	var req model.ClientRequest

	req.ConnectorIdentifier = connector
	req.DatabaseName, req.Username, req.Password = values.Get("dbname"), values.Get("dbuser"), values.Get("dbpass")

	if req.DatabaseName == "" {
		req.DatabaseName = sutils.RandDBName()
	}

	if req.Username == "" {
		req.Username = sutils.RandUserName()
	}

	if req.Password == "" {
		req.Password = sutils.RandPassword()
	}

	con, ok := registry[connector]
	if !ok {
		log.Printf("requested identifier %q not in registry", connector)

		fmt.Fprintf(w, "Connector %q not online", connector)
		return
	}

	log.Printf("%#v", req)

	dest := fmt.Sprintf("http://%s:%s/create-database", con.Address, con.ConnectorPort)

	resp, err := notif.SndLoc(req, dest)
	if err != nil {
		log.Printf("couldn't create database on connector: %s", err.Error())

		fmt.Fprintf(w, "Could not create database: %s", err.Error())
		return
	}

	var msg inet.Message
	respBytes := bytes.NewBufferString(resp)

	err = json.NewDecoder(respBytes).Decode(&msg)
	if err != nil {
		e := fmt.Errorf("malformed response from connector: %s", err.Error())
		log.Println(e.Error())

		fmt.Fprintf(w, e.Error())
		return
	}

	if msg.Status != status.Success {
		inet.WriteHeader(w, http.StatusOK)
		fmt.Fprintf(w, msg.Message)
		return
	}

	db.persist(req)

	fmt.Fprintf(w, "jdbc.default.driverClassName=%s\n", jdbcClassName(con.DBVendor, version))

	var url string
	switch con.DBVendor {
	case "mysql":
		url = mjdbcURL(req.DatabaseName, version, con.Address, con.DBPort)
	case "oracle":
		url = ojdbcURL(con.DBSID, con.Address, con.DBPort)
	case "postgres":
		url = pjdbcURL(req.DatabaseName, con.Address, con.DBPort)
	}

	fmt.Fprintf(w, "jdbc.default.url=%s\n", url)
	fmt.Fprintf(w, "jdbc.default.username=%s\n", req.Username)
	fmt.Fprintf(w, "jdbc.default.password=%s\n", req.Password)
}

func register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	index := strings.LastIndex(r.RemoteAddr, ":")
	addr := r.RemoteAddr[:index]

	ddnc := model.Connector{
		ID:            getID(),
		DBVendor:      req.DBVendor,
		DBPort:        req.DBPort,
		DBSID:         req.DBSID,
		ShortName:     req.ShortName,
		LongName:      req.LongName,
		Identifier:    req.ConnectorName,
		Version:       req.Version,
		Address:       addr,
		ConnectorPort: req.Port,
		Up:            true,
	}

	registry[req.ShortName] = ddnc

	log.Printf("Registered: %s", req.ConnectorName)
	log.Printf("%+v", ddnc)

	conAddr := fmt.Sprintf("%s:%s", ddnc.Address, ddnc.ConnectorPort)

	resp, _ := inet.JSONify(model.RegisterResponse{ID: ddnc.ID, Address: conAddr})

	inet.WriteHeader(w, http.StatusOK)
	w.Write(resp)
}

func unregister(w http.ResponseWriter, r *http.Request) {
	var con model.Connector

	err := json.NewDecoder(r.Body).Decode(&con)
	if err != nil {
		log.Printf("Could not jsonify message: %s", err.Error())
		return
	}

	delete(registry, con.Identifier)

	log.Printf("Unregistered: %s", con.Identifier)
}

func alive(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	buf.WriteString("yup")

	inet.WriteHeader(w, http.StatusOK)

	w.Write(buf.Bytes())
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
