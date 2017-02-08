package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"
)

func listConnectors(w http.ResponseWriter, r *http.Request) {
	list := make(map[string]string, 10)
	for _, con := range registry {
		list[con.ShortName] = con.LongName
	}

	msg := inet.MapMessage{Status: http.StatusOK, Message: list}

	b, st := msg.Compose()

	inet.WriteHeader(w, st)
	w.Write(b)
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
	var req model.ClientRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, inet.ErrorJSONResponse(err))
		return
	}

	con, ok := registry[req.ConnectorIdentifier]
	if !ok {
		log.Printf("requested identifier %q not in registry", req.ConnectorIdentifier)
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

	dest := fmt.Sprintf("http://%s/create-database", con.Address)

	resp, err := notif.SndLoc(req, dest)
	if err != nil {
		log.Printf("couldn't create database on connector: %s", err.Error())

		inet.SendResponse(w, inet.ErrorResponse())
		return
	}

	respBytes := bytes.NewBufferString(resp)

	inet.WriteHeader(w, http.StatusOK)
	w.Write(respBytes.Bytes())
}

func register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, inet.ErrorJSONResponse(err))
		return
	}

	index := strings.LastIndex(r.RemoteAddr, ":")

	addr := fmt.Sprintf("%s:%s", r.RemoteAddr[:index], req.Port)

	ddnc := model.Connector{
		ID:         getID(),
		ShortName:  req.ShortName,
		LongName:   req.LongName,
		Identifier: req.ConnectorName,
		Version:    req.Version,
		Address:    addr,
		Up:         true,
	}

	registry[req.ShortName] = ddnc

	log.Printf("Registered: %s", req.ConnectorName)

	resp, _ := inet.JSONify(model.RegisterResponse{ID: ddnc.ID, Address: ddnc.Address})

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

		inet.SendResponse(w, inet.ErrorJSONResponse(err))
		return
	}

	log.Printf("%+v", msg)
}
