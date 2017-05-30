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
	vis "github.com/djavorszky/ddn/common/visibility"
	"github.com/djavorszky/ddn/server/database"
	"github.com/djavorszky/ddn/server/mail"
	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"

	"path/filepath"

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
	if config.MountLoc != "" {
		loadPage(w, r, "importchooser")
	} else {
		loadPage(w, r, "fileimport")
	}
}

func fileimport(w http.ResponseWriter, r *http.Request) {
	loadPage(w, r, "fileimport")
}

func srvimport(w http.ResponseWriter, r *http.Request) {
	loadPage(w, r, "srvimport")
}

func browseroot(w http.ResponseWriter, r *http.Request) {
	loadPage(w, r, "browse")
}

func browse(w http.ResponseWriter, r *http.Request) {
	loadPage(w, r, "browse")
}

func prepImportAction(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	var (
		connector = r.PostFormValue("connector")
		dbname    = r.PostFormValue("dbname")
		dbuser    = r.PostFormValue("user")
		dbpass    = r.PostFormValue("password")
		dumpfile  = r.PostFormValue("dbdump")
		public    = r.PostFormValue("public")
	)

	dbID, err := doPrepImport(getUser(r), connector, dumpfile, dbname, dbuser, dbpass, public)
	if err != nil {
		session.AddFlash(fmt.Sprintf("Failed preparing import: %s", err.Error()), "fail")
		return
	}

	go doImport(int(dbID), dumpfile)

	session.AddFlash("Started the import process...")
}

func doImport(dbID int, dumpfile string) {
	dbe, err := database.FetchByID(dbID)
	if err != nil {
		log.Printf("Failed getting entry by ID: %s", err.Error())
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: " + err.Error()

		database.Update(&dbe)

		return
	}

	dbe.Status = status.InProgress
	database.Update(&dbe)

	url, err := copyFile(dumpfile)
	if err != nil {
		log.Printf("Failed copying file: %s", err.Error())
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: " + err.Error()

		database.Update(&dbe)
		return
	}

	dbe.Dumpfile = url
	database.Update(&dbe)

	conn, ok := registry[dbe.ConnectorName]
	if !ok {
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: connector went offline."

		database.Update(&dbe)
		return
	}

	_, err = conn.ImportDatabase(int(dbID), dbe.DBName, dbe.DBUser, dbe.DBPass, url)
	if err != nil {
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: " + err.Error()

		database.Update(&dbe)
		os.Remove("./web/dumps/" + dumpfile)
		return
	}

}

func doPrepImport(creator, connector, dumpfile, dbname, dbuser, dbpass, public string) (int, error) {
	conn, ok := registry[connector]
	if !ok {
		return 0, fmt.Errorf("connector went offline")
	}

	if conn.DBVendor == "mssql" {
		dbuser = "clouddb"
		dbpass = "password"
	}

	if dbname == "" && dbuser != "" {
		dbname = dbuser
	}

	ensureValues(&dbname, &dbuser, &dbpass)

	entry := database.Entry{
		DBName:        dbname,
		DBUser:        dbuser,
		DBPass:        dbpass,
		DBSID:         conn.DBSID,
		ConnectorName: connector,
		Creator:       creator,
		DBAddress:     conn.DBAddr,
		DBPort:        conn.DBPort,
		DBVendor:      conn.DBVendor,
		Status:        status.Started,
	}

	if public == "on" {
		entry.Public = vis.Public
	}

	err := database.Insert(&entry)
	if err != nil {
		return 0, fmt.Errorf("failed persisting database locally: %s", err.Error())
	}

	return entry.ID, nil
}

func copyFile(dump string) (string, error) {
	filename := filepath.Base(dump)

	src, err := os.OpenFile(filepath.Join(config.MountLoc, dump), os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed opening source file: %s", err.Error())

	}
	defer src.Close()

	dst, err := os.OpenFile("./web/dumps/"+filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed creating file: %s", err.Error())

	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("failed copying file: %s", err.Error())

	}

	url := fmt.Sprintf("http://%s:%s/dumps/%s", config.ServerHost, config.ServerPort, filename)

	return url, nil
}

func importAction(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)
	defer r.Body.Close()

	r.ParseMultipartForm(32 << 24)

	var (
		connector = r.PostFormValue("connector")
		dbname    = r.PostFormValue("dbname")
		dbuser    = r.PostFormValue("user")
		dbpass    = r.PostFormValue("password")
		public    = r.PostFormValue("public")
	)

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	var filename string
	for _, uploadFile := range r.MultipartForm.File {
		filename = uploadFile[0].Filename

		dst, err := os.OpenFile("./web/dumps/"+filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("Failed creating file: %s", err.Error())
			return
		}
		defer dst.Close()

		upf, err := uploadFile[0].Open()
		if err != nil {
			log.Printf("Failed opening uploaded file: %s", err.Error())
			return
		}

		_, err = io.Copy(dst, upf)
		if err != nil {
			log.Printf("Failed saving file: %s", err.Error())

			os.Remove("./web/dumps/" + filename)
			return
		}
	}

	err = r.MultipartForm.RemoveAll()
	if err != nil {
		log.Printf("Could not removeall multipartform: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%s/dumps/%s", config.ServerHost, config.ServerPort, filename)

	conn, ok := registry[connector]
	if !ok {
		session.AddFlash(fmt.Sprintf("Failed importing database, connector %s went offline", connector), "fail")
		os.Remove("./web/dumps/" + filename)
		return
	}

	if conn.DBVendor == "mssql" {
		dbuser = "clouddb"
		dbpass = "password"
	}

	if dbname == "" && dbuser != "" {
		dbname = dbuser
	}

	ensureValues(&dbname, &dbuser, &dbpass)

	entry := database.Entry{
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

	if public == "on" {
		entry.Public = vis.Public
	}

	err = database.Insert(&entry)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		session.AddFlash(fmt.Sprintf("failed persisting database locally: %s", err.Error()))
		os.Remove("./web/dumps/" + filename)
		return
	}

	resp, err := conn.ImportDatabase(entry.ID, dbname, dbuser, dbpass, url)
	if err != nil {
		session.AddFlash(err.Error(), "fail")

		database.Delete(entry)
		os.Remove("./web/dumps/" + filename)
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
		public    = r.PostFormValue("public")
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

	if conn.DBVendor == "mssql" {
		dbuser = "clouddb"
		dbpass = "password"
	}

	ID := getID()
	if dbname == "" && dbuser != "" {
		dbname = dbuser
	}

	ensureValues(&dbname, &dbuser, &dbpass)

	resp, err := conn.CreateDatabase(ID, dbname, dbuser, dbpass)
	if err != nil {
		session.AddFlash(err.Error(), "fail")
		return
	}

	entry := database.Entry{
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

	if public == "on" {
		entry.Public = vis.Public
	}

	err = database.Insert(&entry)
	if err != nil {
		log.Printf("Failed adding entry: %s", err.Error())
	}

	session.Values["id"] = entry.ID
	session.AddFlash(resp, "success")
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

	dbe, err := database.FetchByID(ID)
	if err != nil {
		http.Error(w, "Failed fetching entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dbe.ExpiryDate = time.Now().AddDate(0, 0, 30)

	err = database.Update(&dbe)
	if err != nil {
		http.Error(w, "Failed updating entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

	dbe, err := database.FetchByID(ID)
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

	dbe.Status = status.DropInProgress

	database.Update(&dbe)

	go dropAsync(conn, ID, dbe.DBName, dbe.DBUser)

	session.AddFlash("Started to drop the database.", "msg")
}

func dropAsync(conn model.Connector, ID int, dbname, dbuser string) {
	dbe, err := database.FetchByID(ID)
	if err != nil {
		log.Printf("Couldn't drop database %q on connector %q: %s", dbname, conn.ShortName, err.Error())
		return
	}

	_, err = conn.DropDatabase(ID, dbname, dbuser)
	if err != nil {
		dbe.Status = status.DropDatabaseFailed
		dbe.Message = err.Error()

		database.Update(&dbe)

		log.Printf("Couldn't drop database %q on connector %q: %s", dbname, conn.ShortName, err.Error())
		return
	}

	database.Delete(dbe)
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

	dbe, err := database.FetchByID(ID)
	if err != nil {
		log.Printf("Failed querying for database: %s", err.Error())
		session.AddFlash("Failed querying database", "fail")
		return
	}

	if dbe.Public == vis.Private && dbe.Creator != user {
		log.Printf("User %q tried to get portalext of db created by %q.", user, dbe.Creator)
		session.AddFlash("Failed fetching portal-ext: You can only fetch the portal-ext of public databases or ones that you created.", "fail")
		return
	}

	session.Values["id"] = ID
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

	dbe, err := database.FetchByID(msg.ID)
	if err != nil {
		log.Printf("Failed querying for database: %s", err.Error())
		return
	}

	dbe.Status = msg.StatusID

	database.Update(&dbe)

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
		mail.Send(dbe.Creator, fmt.Sprintf("[Cloud DB] Importing %q failed", dbe.DBName), fmt.Sprintf(`<h3>Import database failed</h3>
		
<p>Your request to import a(n) %q database named %q has failed with the following message:</p>
<p>%q</p>

<p>We're sorry for the inconvenience caused.</p>
<p>Visit <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>`, dbe.DBVendor, dbe.DBName, msg.Message))

		// Update dbentry as well
		dbe.Message = msg.Message

		err = database.Update(&dbe)
		if err != nil {
			log.Printf("Couldn't update local db: %s", err.Error())
		}
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
		case "mssql":
			jdbc62x = liferay.MSSQLJDBC(dbe.DBAddress, dbe.DBPort, dbe.DBName, dbe.DBUser, dbe.DBPass)
			jdbcDXP = jdbc62x
		}

		mail.Send(dbe.Creator, fmt.Sprintf("[Cloud DB] Importing %q succeeded", dbe.DBName), fmt.Sprintf(`<h3>Import database successful</h3>
		
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

func ensureValues(vals ...*string) {
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

	dbentry := database.Entry{
		DBName:        req.DatabaseName,
		DBUser:        req.Username,
		DBPass:        req.Password,
		Creator:       req.RequesterEmail,
		Dumpfile:      req.DumpLocation,
		ConnectorName: req.ConnectorIdentifier,
		Status:        status.Success,
	}

	database.Insert(&dbentry)

	return con, nil
}

func getUser(r *http.Request) string {
	usr, _ := r.Cookie("user")

	return usr.Value
}
