package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"

	"github.com/djavorszky/liferay"
	"github.com/gorilla/mux"
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

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	file, handler, err := r.FormFile("dbdump")
	if err != nil {
		log.Printf("File upload failed: %s", err.Error())
		session.AddFlash("File upload failed: "+err.Error(), "fail")

		return
	}
	defer file.Close()

	f, err := os.OpenFile("./web/dumps/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Saving file to dumps directory failed: %s", err.Error())
		session.AddFlash("Saving file to dumps directory failed: "+err.Error(), "fail")

		return
	}
	defer f.Close()

	io.Copy(f, file)

	err = r.MultipartForm.RemoveAll()
	if err != nil {
		log.Printf("Could not removeall multipartform: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%s/dumps/%s", config.ServerHost, config.ServerPort, handler.Filename)

	conn, ok := registry[connector]
	if !ok {
		session.AddFlash(fmt.Sprintf("Failed creating database, connector %s went offline", connector), "fail")
		return
	}

	if dbname == "" && dbuser != "" {
		dbname = dbuser
	}

	ensureHasValues(&dbname, &dbuser, &dbpass)

	entry := model.DBEntry{
		DBName:        dbname,
		DBUser:        dbuser,
		DBPass:        dbpass,
		DBSID:         conn.DBSID,
		ConnectorName: connector,
		Creator:       getUser(r),
		Dumpfile:      url,
		DBAddress:     conn.DBAddr,
		DBPort:        conn.DBPort,
		DBVendor:      conn.DBVendor,
		Status:        status.Started,
	}

	dbID, err := db.persist(entry)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		session.AddFlash(fmt.Sprintf("failed persisting database locally: %s", err.Error()))
		return
	}

	resp, err := conn.ImportDatabase(int(dbID), dbname, dbuser, dbpass, url)
	if err != nil {
		session.AddFlash(err.Error(), "fail")

		db.delete(dbID)
		return
	}

	session.AddFlash(resp, "msg")
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
		session.AddFlash(err.Error(), "fail")
		return
	}

	entry := model.DBEntry{
		DBName:        dbname,
		DBUser:        dbuser,
		DBPass:        dbpass,
		DBSID:         conn.DBSID,
		ConnectorName: connector,
		Creator:       getUser(r),
		DBAddress:     conn.DBAddr,
		DBPort:        conn.DBPort,
		DBVendor:      conn.DBVendor,
		Status:        status.Success,
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

	ddnc := model.Connector{
		ID:            getID(),
		DBVendor:      req.DBVendor,
		DBPort:        req.DBPort,
		DBAddr:        req.DBAddr,
		DBSID:         req.DBSID,
		ShortName:     req.ShortName,
		LongName:      req.LongName,
		Identifier:    req.ConnectorName,
		Version:       req.Version,
		Address:       req.Addr,
		ConnectorPort: req.Port,
		Up:            true,
	}

	registry[req.ShortName] = ddnc

	log.Printf("Registered: %v", req.ConnectorName)

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

	delete(registry, con.ShortName)

	log.Printf("Unregistered: %s", con.Identifier)
}

func alive(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	buf.WriteString("yup")

	inet.WriteHeader(w, http.StatusOK)

	w.Write(buf.Bytes())
}

func login(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	r.ParseForm()

	email := r.PostFormValue("email")

	cookie := http.Cookie{
		Name:    "user",
		Value:   email,
		Expires: time.Now().AddDate(1, 0, 0),
	}

	http.SetCookie(w, &cookie)
}

func logout(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	userCookie, err := r.Cookie("user")
	if err != nil {
		return
	}

	userCookie.Value = ""

	http.SetCookie(w, userCookie)
}

func extend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "couldn't convert id to int.", http.StatusInternalServerError)
		return
	}

	db.updateColumn(ID, "expiryDate", "NOW() + INTERVAL 30 DAY")
	db.updateColumn(ID, "status", status.Success)

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	session.AddFlash("Successfully extended the expiry date", "msg")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func drop(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	user := getUser(r)

	if user == "" {
		log.Println("Drop database tried without a logged in user.")
		return
	}

	vars := mux.Vars(r)

	ID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "couldn't convert id to int.", http.StatusInternalServerError)
		return
	}

	dbe, err := db.entryByID(int64(ID))
	if err != nil {
		log.Printf("Failed querying for database: %s", err.Error())
		session.AddFlash("Failed querying database", "fail")
		return
	}

	if dbe.Creator != user {
		log.Printf("User %q tried to drop database of user %q.", user, dbe.Creator)
		session.AddFlash("Failed dropping database: You can only drop databases you created.", "fail")
		return
	}

	conn, ok := registry[dbe.ConnectorName]
	if !ok {
		log.Printf("Connector %q is offline, can't drop database with id '%d'", dbe.ConnectorName, ID)
		session.AddFlash("Unable to drop database: Connector is down.", "fail")
		return
	}

	db.updateColumn(ID, "status", status.PendingImmediateRemoval)

	go dropAsync(conn, ID, dbe.DBName, dbe.DBUser)

	session.AddFlash("Started to drop the database.", "msg")
}

func dropAsync(conn model.Connector, ID int, dbname, dbuser string) {
	_, err := conn.DropDatabase(ID, dbname, dbuser)
	if err != nil {
		log.Printf("Couldn't drop database %q on connector %q: %s", dbname, conn.ShortName, err.Error())
		return
	}

	db.delete(int64(ID))
}

func portalext(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	user := getUser(r)

	if user == "" {
		log.Println("Portal-ext request without logged in user.")
		return
	}

	vars := mux.Vars(r)

	ID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "couldn't convert id to int.", http.StatusInternalServerError)
		return
	}

	dbe, err := db.entryByID(int64(ID))
	if err != nil {
		log.Printf("Failed querying for database: %s", err.Error())
		session.AddFlash("Failed querying database", "fail")
		return
	}

	if dbe.Creator != user {
		log.Printf("User %q tried to get portalext of db created by %q.", user, dbe.Creator)
		session.AddFlash("Failed dropping database: You can only drop databases you created.", "fail")
		return
	}

	session.Values["id"] = int64(ID)
	session.AddFlash("Portal-exts are as follows", "success")
}

// upd8 updates the status of the databases.
func upd8(w http.ResponseWriter, r *http.Request) {
	var msg notif.Msg

	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Printf("couldn't decode json request: %s", err.Error())

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	db.updateColumn(msg.ID, "status", msg.StatusID)

	dbe, err := db.entryByID(int64(msg.ID))
	if err != nil {
		log.Printf("Failed querying for database: %s", err.Error())
		return
	}

	// Delete the dumpfile once import is started or if an error has occurred.
	if dbe.Status == status.ImportInProgress || dbe.IsErr() {
		loc := strings.LastIndex(dbe.Dumpfile, "/")

		file := fmt.Sprintf("./web/dumps/%s", dbe.Dumpfile[loc+1:])

		err = os.Remove(file)
		if err != nil {
			log.Printf("Failed to remove dumpfile %s: %s", file, err.Error())
		}
	}

	if dbe.IsErr() {
		sendMail(dbe.Creator, fmt.Sprintf("[Cloud DB] Importing %q failed", dbe.DBName), fmt.Sprintf(`<h3>Import database failed</h3>
		
<p>Your request to import a(n) %q database named %q has failed with the following message:</p>
<p>%q</p>

<p>We're sorry for the inconvenience caused.</p>
<p>Visit <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>`, dbe.DBVendor, dbe.DBName, msg.Message))
	}

	if dbe.Status == status.Success {
		var (
			jdbc62x liferay.JDBC
			jdbcDXP liferay.JDBC
		)

		switch dbe.DBVendor {
		case "mysql":
			jdbc62x = liferay.MysqlJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
			jdbcDXP = liferay.MysqlJDBCDXP(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
		case "postgres":
			jdbc62x = liferay.PostgreJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
			jdbcDXP = jdbc62x
		case "oracle":
			jdbc62x = liferay.OracleJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBSID, dbe.DBUser, dbe.DBPass)
			jdbcDXP = jdbc62x
		}

		sendMail(dbe.Creator, fmt.Sprintf("[Cloud DB] Importing %q succeeded", dbe.DBName), fmt.Sprintf(`<h3>Import database successful</h3>
		
<p>The %s import that you started completed successfully.</p>
<p>Below you can find the portal-exts, should you need them:</p>

<h2><= 6.2 EE properties</h2>
<pre>
%s
%s
%s
%s
</pre>

<h2>DXP properties</h2>
<pre>
%s
%s
%s
%s
</pre>

<p>Visit <a href="http://cloud-db.liferay.int">Cloud DB</a> for more awesomeness.</p>
<p>Cheers</p>`, dbe.DBVendor, jdbc62x.Driver, jdbc62x.URL, jdbc62x.User, jdbc62x.Password, jdbcDXP.Driver, jdbcDXP.URL, jdbcDXP.User, jdbcDXP.Password))
	}
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
		Status:        status.Success,
	}

	db.persist(dbentry)

	return con, nil
}

func getUser(r *http.Request) string {
	usr, _ := r.Cookie("user")

	return usr.Value
}
