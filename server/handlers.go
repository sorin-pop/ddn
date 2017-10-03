package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
	vis "github.com/djavorszky/ddn/common/visibility"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/mail"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/djavorszky/liferay"
	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"
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
		session.AddFlash(fmt.Sprintf("Failed preparing import: %v", err), "fail")
		return
	}

	go doImport(int(dbID), dumpfile)

	session.AddFlash("Started the import process...", "msg")
}

func doImport(dbID int, dumpfile string) {
	dbe, err := db.FetchByID(dbID)
	if err != nil {
		logger.Error("Failed getting entry by ID: %v", err)
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: " + err.Error()
		dbe.ExpiryDate = time.Now().AddDate(0, 0, 2)

		db.Update(&dbe)

		return
	}

	dbe.Status = status.CopyInProgress
	db.Update(&dbe)

	url, err := copyFile(dumpfile)
	if err != nil {
		logger.Error("file copy: %v", err)
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: " + err.Error()
		dbe.ExpiryDate = time.Now().AddDate(0, 0, 2)

		db.Update(&dbe)
		return
	}

	dbe.Dumpfile = url
	db.Update(&dbe)

	conn, ok := registry.Get(dbe.ConnectorName)
	if !ok {
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: connector went offline."
		dbe.ExpiryDate = time.Now().AddDate(0, 0, 2)

		db.Update(&dbe)
		return
	}

	_, err = conn.ImportDatabase(int(dbID), dbe.DBName, dbe.DBUser, dbe.DBPass, url)
	if err != nil {
		dbe.Status = status.ImportFailed
		dbe.Message = "Server error: " + err.Error()
		dbe.ExpiryDate = time.Now().AddDate(0, 0, 2)

		db.Update(&dbe)
		os.Remove(fmt.Sprintf("%s/web/dumps/%s", workdir, dumpfile))
		return
	}

}

func doPrepImport(creator, connector, dumpfile, dbname, dbuser, dbpass, public string) (int, error) {
	conn, ok := registry.Get(connector)
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

	entry := data.Row{
		DBName:        dbname,
		DBUser:        dbuser,
		DBPass:        dbpass,
		DBSID:         conn.DBSID,
		CreateDate:    time.Now(),
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
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

	err := db.Insert(&entry)
	if err != nil {
		return 0, fmt.Errorf("database persist: %v", err)
	}

	return entry.ID, nil
}

func copyFile(dump string) (string, error) {
	filename := filepath.Base(dump)

	src, err := os.OpenFile(filepath.Join(config.MountLoc, dump), os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed opening source file: %v", err)

	}
	defer src.Close()

	dst, err := os.OpenFile(fmt.Sprintf("%s/web/dumps/%s", workdir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed creating file: %v", err)

	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("failed copying file: %v", err)

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

		dst, err := os.OpenFile(fmt.Sprintf("%s/web/dumps/%s", workdir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			logger.Error("Failed creating file: %v", err)
			return
		}
		defer dst.Close()

		upf, err := uploadFile[0].Open()
		if err != nil {
			logger.Error("Failed opening uploaded file: %v", err)
			return
		}

		_, err = io.Copy(dst, upf)
		if err != nil {
			logger.Error("Failed saving file: %v", err)

			os.Remove(fmt.Sprintf("%s/web/dumps/%s", workdir, filename))
			return
		}
	}

	err = r.MultipartForm.RemoveAll()
	if err != nil {
		logger.Error("Could not removeall multipartform: %v", err)
	}

	url := fmt.Sprintf("http://%s:%s/dumps/%s", config.ServerHost, config.ServerPort, filename)

	conn, ok := registry.Get(connector)
	if !ok {
		session.AddFlash(fmt.Sprintf("Failed importing database, connector %s went offline", connector), "fail")
		os.Remove(fmt.Sprintf("%s/web/dumps/%s", workdir, filename))
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

	entry := data.Row{
		DBName:        dbname,
		DBUser:        dbuser,
		DBPass:        dbpass,
		DBSID:         conn.DBSID,
		CreateDate:    time.Now(),
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
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

	err = db.Insert(&entry)
	if err != nil {
		logger.Error("persist: %v", err)
		session.AddFlash(fmt.Sprintf("failed persisting database locally: %v", err))
		os.Remove(fmt.Sprintf("%s/web/dumps/%s", workdir, filename))
		return
	}

	resp, err := conn.ImportDatabase(entry.ID, dbname, dbuser, dbpass, url)
	if err != nil {
		session.AddFlash(err.Error(), "fail")

		db.Delete(entry)
		os.Remove(fmt.Sprintf("%s/web/dumps/%s", workdir, filename))
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

	conn, ok := registry.Get(connector)
	if !ok {
		session.AddFlash(fmt.Sprintf("Failed creating database, connector %s went offline", connector), "fail")
		return
	}

	if conn.DBVendor == "mssql" {
		dbuser = "clouddb"
		dbpass = "password"
	}

	ID := registry.ID()
	if dbname == "" && dbuser != "" {
		dbname = dbuser
	}

	ensureValues(&dbname, &dbuser, &dbpass)

	resp, err := conn.CreateDatabase(ID, dbname, dbuser, dbpass)
	if err != nil {
		session.AddFlash(err.Error(), "fail")
		return
	}

	entry := data.Row{
		DBName:        dbname,
		DBUser:        dbuser,
		DBPass:        dbpass,
		DBSID:         conn.DBSID,
		CreateDate:    time.Now(),
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
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

	err = db.Insert(&entry)
	if err != nil {
		logger.Error("persist: %v", err)
	}

	session.Values["id"] = entry.ID
	session.AddFlash(resp, "success")
}

func register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("json decode: %v", err)

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	ddnc := model.Connector{
		ID:            registry.ID(),
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

	registry.Store(ddnc)

	logger.Info("Registered: %v", req.ConnectorName)

	conAddr := fmt.Sprintf("%s:%s", ddnc.Address, ddnc.ConnectorPort)

	resp, _ := inet.JSONify(model.RegisterResponse{ID: ddnc.ID, Address: conAddr})

	inet.WriteHeader(w, http.StatusOK)
	w.Write(resp)
}

func unregister(w http.ResponseWriter, r *http.Request) {
	var conn model.Connector

	err := json.NewDecoder(r.Body).Decode(&conn)
	if err != nil {
		logger.Error("json encode: %v", err)
		return
	}

	registry.Remove(conn.ShortName)

	logger.Info("Unregistered: %s", conn.Identifier)
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	buf.WriteString("ba-bump")

	inet.WriteHeader(w, http.StatusOK)
	w.Write(buf.Bytes())
}

func alive(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	vars := mux.Vars(r)
	shortname := vars["shortname"]

	if registry.Exists(shortname) {
		buf.WriteString("yup")
		inet.WriteHeader(w, http.StatusOK)
	} else {
		buf.WriteString("nope")
		inet.WriteHeader(w, http.StatusNotFound)
	}

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

	dbe, err := db.FetchByID(ID)
	if err != nil {
		http.Error(w, "Failed fetching entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dbe.ExpiryDate = time.Now().AddDate(0, 0, 30)
	dbe.Status = status.Success

	err = db.Update(&dbe)
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
		logger.Error("Drop database tried without a logged in user.")
		return
	}

	vars := mux.Vars(r)

	ID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "couldn't convert id to int.", http.StatusInternalServerError)
		return
	}

	dbe, err := db.FetchByID(ID)
	if err != nil {
		logger.Error("FetchById: %v", err)
		session.AddFlash("Failed querying database", "fail")
		return
	}

	if dbe.Creator != user {
		logger.Error("User %q tried to drop database of user %q.", user, dbe.Creator)
		session.AddFlash("Failed dropping database: You can only drop databases you created.", "fail")
		return
	}

	conn, ok := registry.Get(dbe.ConnectorName)
	if !ok {
		logger.Error("Connector %q is offline, can't drop database with id '%d'", dbe.ConnectorName, ID)
		session.AddFlash("Unable to drop database: Connector is down.", "fail")
		return
	}

	dbe.Status = status.DropInProgress

	db.Update(&dbe)

	go dropAsync(conn, ID, dbe.DBName, dbe.DBUser)

	session.AddFlash("Started to drop the database.", "msg")
}

func dropAsync(conn model.Connector, ID int, dbname, dbuser string) {
	dbe, err := db.FetchByID(ID)
	if err != nil {
		logger.Error("Couldn't drop database %q on connector %q: %v", dbname, conn.ShortName, err)
		return
	}

	_, err = conn.DropDatabase(ID, dbname, dbuser)
	if err != nil {
		dbe.Status = status.DropDatabaseFailed
		dbe.Message = err.Error()

		db.Update(&dbe)

		logger.Error("Couldn't drop database %q on connector %q: %s", dbname, conn.ShortName, err)
		return
	}

	db.Delete(dbe)
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
		logger.Error("Portal-ext request without logged in user.")
		return
	}

	vars := mux.Vars(r)

	ID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "couldn't convert id to int.", http.StatusInternalServerError)
		return
	}

	dbe, err := db.FetchByID(ID)
	if err != nil {
		logger.Error("FetchById: %v", err)
		session.AddFlash("Failed querying database", "fail")
		return
	}

	if dbe.Public == vis.Private && dbe.Creator != user {
		logger.Error("User %q tried to get portalext of db created by %q.", user, dbe.Creator)
		session.AddFlash("Failed fetching portal-ext: You can only fetch the portal-ext of public databases or ones that you created.", "fail")
		return
	}

	session.Values["id"] = ID
	session.AddFlash("Portal-exts are as follows", "success")
}

func recreate(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Failed getting session: "+err.Error(), http.StatusInternalServerError)
	}
	defer session.Save(r, w)

	user := getUser(r)

	if user == "" {
		logger.Error("Portal-ext request without logged in user.")
		return
	}

	vars := mux.Vars(r)

	ID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "couldn't convert id to int.", http.StatusInternalServerError)
		return
	}

	dbe, err := db.FetchByID(ID)
	if err != nil {
		logger.Error("FetchById: %v", err)
		session.AddFlash("Failed querying database", "fail")
		return
	}

	if dbe.Creator != user {
		logger.Error("User %q tried to get recreate the database created by %q.", user, dbe.Creator)
		session.AddFlash("Failed recreating databasee: You can only recreate database you created.", "fail")
		return
	}

	conn, ok := registry.Get(dbe.ConnectorName)
	if !ok {
		logger.Error("Connector %q is offline, can't recreate database with id '%d'", dbe.ConnectorName, ID)
		session.AddFlash("Unable to recreate database: Connector is down.", "fail")
		return
	}

	go recreateAsync(conn, dbe)

	session.AddFlash("Started to recreate", "msg")
}

func recreateAsync(conn model.Connector, dbe data.Row) {
	_, err := conn.DropDatabase(dbe.ID, dbe.DBName, dbe.DBUser)
	if err != nil {
		dbe.Status = status.DropDatabaseFailed
		dbe.Message = err.Error()

		db.Update(&dbe)

		logger.Error("Recreate: couldn't drop database %q on connector %q: %s", dbe.DBName, conn.ShortName, err)

		return
	}

	_, err = conn.CreateDatabase(dbe.ID, dbe.DBName, dbe.DBUser, dbe.DBPass)
	if err != nil {
		dbe.Status = status.CreateDatabaseFailed
		dbe.Message = err.Error()

		db.Update(&dbe)

		logger.Error("Recreate: couldn't create database %q on connector %q: %s", dbe.DBName, conn.ShortName, err)

		return
	}

	dbe.Status = status.Success
	db.Update(&dbe)

	return
}

// upd8 updates the status of the databases.
func upd8(w http.ResponseWriter, r *http.Request) {
	var msg notif.Msg

	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		logger.Error("json decode: %v", err)

		inet.SendResponse(w, http.StatusBadRequest, inet.ErrorJSONResponse(err))
		return
	}

	dbe, err := db.FetchByID(msg.ID)
	if err != nil {
		logger.Error("FetchById: %v", err)
		return
	}

	dbe.Status = msg.StatusID

	db.Update(&dbe)

	// Delete the dumpfile once import is started or if an error has occurred.
	if dbe.Status == status.ImportInProgress || dbe.IsErr() {
		loc := strings.LastIndex(dbe.Dumpfile, "/")

		file := fmt.Sprintf("%s/web/dumps/%s", workdir, dbe.Dumpfile[loc+1:])
		os.Remove(file)
	}

	if dbe.IsErr() {
		mail.Send(dbe.Creator, fmt.Sprintf("[Cloud DB] Importing %q failed", dbe.DBName), fmt.Sprintf(`<h3>Import database failed</h3>
		
<p>Your request to import a(n) %q database named %q has failed with the following message:</p>
<p>%q</p>

<p>We're sorry for the inconvenience caused.</p>
<p>Visit <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>`, dbe.DBVendor, dbe.DBName, msg.Message))

		// Update dbentry as well
		dbe.Message = msg.Message
		dbe.ExpiryDate = time.Now().AddDate(0, 0, 2)

		err = db.Update(&dbe)
		if err != nil {
			logger.Error("Update: %v", err)
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
	con, ok := registry.Get(req.ConnectorIdentifier)
	if !ok {
		return model.Connector{}, fmt.Errorf("requested identifier %q not in registry", req.ConnectorIdentifier)
	}

	dest := fmt.Sprintf("http://%s:%s/create-database", con.Address, con.ConnectorPort)

	resp, err := notif.SndLoc(req, dest)
	if err != nil {
		return model.Connector{}, fmt.Errorf("couldn't create database on connector: %v", err)
	}

	var msg inet.Message
	respBytes := bytes.NewBufferString(resp)

	err = json.NewDecoder(respBytes).Decode(&msg)
	if err != nil {
		return model.Connector{}, fmt.Errorf("malformed response from connector: %v", err)
	}

	if msg.Status != status.Success {
		return model.Connector{}, fmt.Errorf("creating database failed: %s", msg.Message)
	}

	dbentry := data.Row{
		DBName:        req.DatabaseName,
		DBUser:        req.Username,
		DBPass:        req.Password,
		Creator:       req.RequesterEmail,
		CreateDate:    time.Now(),
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
		Dumpfile:      req.DumpLocation,
		ConnectorName: req.ConnectorIdentifier,
		Status:        status.Success,
	}

	db.Insert(&dbentry)

	return con, nil
}

func getUser(r *http.Request) string {
	usr, _ := r.Cookie("user")

	return usr.Value
}
