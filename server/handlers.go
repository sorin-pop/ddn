package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("veryverysecretkey"))

func index(w http.ResponseWriter, r *http.Request) {
	loadPage(w, r, "home")
}

func createdb(w http.ResponseWriter, r *http.Request) {
	loadPage(w, r, "createdb")
}

func importdb(w http.ResponseWriter, r *http.Request) {
	loadPage(w, r, "importdb")
}

func importAction(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	r.ParseMultipartForm(32 << 20)

	var (
		connector = r.PostFormValue("connector")
		dbname    = r.PostFormValue("dbname")
		dbuser    = r.PostFormValue("user")
		dbpass    = r.PostFormValue("password")
	)

	file, handler, err := r.FormFile("dbdump")
	if err != nil {
		log.Printf("File upload failed: %s", err.Error())
		http.Error(w, "File upload failed: "+err.Error(), http.StatusInternalServerError)

		return
	}
	defer file.Close()

	f, err := os.OpenFile("./web/dumps/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Saving file to dumps directory failed: %s", err.Error())
		http.Error(w, "Saving file to dumps directory failed: "+err.Error(), http.StatusInternalServerError)

		return
	}
	defer f.Close()

	io.Copy(f, file)

	url := fmt.Sprintf("http://%s:%s/dumps/%s", config.ServerHost, config.ServerPort, handler.Filename)

	log.Println(url)

	// This is for debugging reasons only:
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	conn, ok := registry[connector]
	if !ok {
		session.AddFlash(fmt.Sprintf("Failed creating database, connector %s went offline", connector), "fail")
		return
	}

	//	ID := getID()
	if dbname == "" && dbuser != "" {
		dbname = dbuser
	}

	ensureHasValues(&dbname, &dbuser, &dbpass)

	resp, err := conn.ImportDatabase(id, dbname, dbuser, dbpass, url)
	if err != nil {
		session.AddFlash(fmt.Sprintf("failed to create database: %s", err.Error()), "fail")
		return
	}

	session.AddFlash(resp, "success")
}

func createAction(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	r.ParseForm()

	var (
		connector = r.PostFormValue("connector")
		dbname    = r.PostFormValue("dbname")
		dbuser    = r.PostFormValue("user")
		dbpass    = r.PostFormValue("password")
	)

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	conn, ok := registry[connector]
	if !ok {
		session.AddFlash(fmt.Sprintf("Failed creating database, connector %s went offline", connector), "fail")
		return
	}

	ID := getID()
	if dbname == "" && dbuser != "" {
		dbname = dbuser
	}

	ensureHasValues(&dbname, &dbuser, &dbpass)

	resp, err := conn.CreateDatabase(ID, dbname, dbuser, dbpass)
	if err != nil {
		session.AddFlash(fmt.Sprintf("failed to create database: %s", err.Error()), "fail")
		return
	}

	entry := model.DBEntry{
		DBName:        dbname,
		DBUser:        dbuser,
		DBPass:        dbpass,
		ConnectorName: connector,
		DBAddress:     conn.Address,
		DBPort:        conn.DBPort,
		DBVendor:      conn.DBVendor,
	}

	dbID, err := db.persist(entry)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	session.Values["id"] = dbID
	session.AddFlash(resp, "success")
}

func listConnectors(w http.ResponseWriter, r *http.Request) {
	list := make(map[string]string, 10)
	for _, con := range registry {
		list[con.ShortName] = con.LongName
	}

	msg := inet.MapMessage{Status: status.Success, Message: list}

	inet.SendResponse(w, http.StatusOK, msg)
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

	if addr == "[::1]" {
		addr = "127.0.0.1"
	}

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

	delete(registry, con.ShortName)

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

func ensureHasValues(vals ...*string) {
	for _, v := range vals {
		if *v == "" {
			*v = sutils.RandName()
		}
	}
}

func doCreateDatabase(req model.ClientRequest) (model.Connector, error) {
	con, ok := registry[req.ConnectorIdentifier]
	if !ok {
		return model.Connector{}, fmt.Errorf("requested identifier %q not in registry", req.ConnectorIdentifier)
	}

	dest := fmt.Sprintf("http://%s:%s/create-database", con.Address, con.ConnectorPort)

	resp, err := notif.SndLoc(req, dest)
	if err != nil {
		return model.Connector{}, fmt.Errorf("couldn't create database on connector: %s", err.Error())
	}

	var msg inet.Message
	respBytes := bytes.NewBufferString(resp)

	err = json.NewDecoder(respBytes).Decode(&msg)
	if err != nil {
		return model.Connector{}, fmt.Errorf("malformed response from connector: %s", err.Error())
	}

	if msg.Status != status.Success {
		return model.Connector{}, fmt.Errorf("creating database failed: %s", msg.Message)
	}

	dbentry := model.DBEntry{
		DBName:        req.DatabaseName,
		DBUser:        req.Username,
		DBPass:        req.Password,
		Creator:       req.Requester,
		Dumpfile:      req.DumpLocation,
		ConnectorName: req.ConnectorIdentifier,
	}

	db.persist(dbentry)

	return con, nil
}
