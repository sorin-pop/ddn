package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/djavorszky/ddnc/common/inet"
	"github.com/djavorszky/ddnc/common/model"
)

func listConnectors(w http.ResponseWriter, r *http.Request) {
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
}

func register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest

	log.Println("Registry request!")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, inet.ErrorJSONResponse(err))
		return
	}

	ddnc := model.Connector{
		ID:        getID(),
		ShortName: req.ShortName,
		LongName:  req.LongName,
		Version:   req.Version,
		Address:   r.RemoteAddr,
		Up:        true,
	}

	registry[ddnc.Address] = ddnc

	fmt.Printf("Registered: %+v", ddnc)

	resp, _ := inet.JSONify(model.RegisterResponse{ID: ddnc.ID, Address: ddnc.Address})

	inet.WriteHeader(w, http.StatusOK)
	w.Write(resp)
}

func unregister(w http.ResponseWriter, r *http.Request) {

	var con model.Connector

	err := json.NewDecoder(r.Body).Decode(&con)
	if err != nil {
		log.Printf("Could not jsonify message: %s", err.Error())
	}

	log.Printf("Unregistered: %+v", con)
}

func alive(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	buf.WriteString("yup")

	inet.WriteHeader(w, http.StatusOK)

	w.Write(buf.Bytes())
}
