package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"strings"

	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/server/brwsr"
	"github.com/djavorszky/ddn/server/database"
	"github.com/djavorszky/ddn/server/registry"
	"github.com/djavorszky/liferay"
	"github.com/gorilla/mux"
)

// Page is a struct holding the data to be displayed on the welcome page.
type Page struct {
	UseCDN           bool
	Connectors       []model.Connector
	AnyOnline        bool
	Title            string
	Pages            map[string]string
	ActivePage       string
	Message          string
	MessageType      string
	User             string
	HasUser          bool
	HasEntry         bool
	PrivateDatabases []database.Entry
	PublicDatabases  []database.Entry
	HasPrivateDBs    bool
	HasPublicDBs     bool
	Ext62            liferay.JDBC
	ExtDXP           liferay.JDBC
	FileList         brwsr.FileList
	HasMountedFolder bool
	DumpLoc          string
}

func loadPage(w http.ResponseWriter, r *http.Request, pages ...string) {

	page := Page{
		UseCDN:           config.UseCDN,
		Connectors:       registry.List(),
		Title:            getTitle(r.URL.Path),
		Pages:            getPages(),
		ActivePage:       r.URL.Path,
		HasMountedFolder: config.MountLoc != "",
	}

	for _, conn := range registry.List() {
		if conn.Up {
			page.AnyOnline = true
			break
		}
	}

	userCookie, err := r.Cookie("user")
	if err != nil || userCookie.Value == "" {
		// if there's an err, it can only happen if there is no cookie.
		toLoad := []string{"base", "nav", "login"}
		tmpl, err := buildTemplate(toLoad...)
		if err != nil {
			panic(err)
		}

		err = tmpl.ExecuteTemplate(w, "base", page)
		if err != nil {
			panic(err)
		}
		return
	}

	page.User = userCookie.Value
	page.HasUser = true

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if flashes := session.Flashes("success"); len(flashes) > 0 {
		page.Message = flashes[0].(string)
		page.MessageType = "success"

		id := session.Values["id"].(int)

		if id != 0 {
			page.HasEntry = true
			entry, err := database.FetchByID(id)
			if err != nil {
				log.Printf("Failed querying for database: %s", err.Error())
				session.AddFlash("Failed querying database", "fail")
			} else {
				switch entry.DBVendor {
				case "mysql":
					page.Ext62 = liferay.MysqlJDBC(entry.DBAddress, entry.DBPort, entry.DBName, entry.DBUser, entry.DBPass)
					page.ExtDXP = liferay.MysqlJDBCDXP(entry.DBAddress, entry.DBPort, entry.DBName, entry.DBUser, entry.DBPass)
				case "postgres":
					page.Ext62 = liferay.PostgreJDBC(entry.DBAddress, entry.DBPort, entry.DBName, entry.DBUser, entry.DBPass)
					page.ExtDXP = page.Ext62
				case "oracle":
					page.Ext62 = liferay.OracleJDBC(entry.DBAddress, entry.DBPort, entry.DBSID, entry.DBUser, entry.DBPass)
					page.ExtDXP = page.Ext62
				case "mssql":
					page.Ext62 = liferay.MSSQLJDBC(entry.DBAddress, entry.DBPort, entry.DBName, entry.DBUser, entry.DBPass)
					page.ExtDXP = page.Ext62
				}
			}
		}
	} else if flashes := session.Flashes("fail"); len(flashes) > 0 {
		page.Message = flashes[0].(string)
		page.MessageType = "danger"
	} else if flashes := session.Flashes("msg"); len(flashes) > 0 {
		page.Message = flashes[0].(string)
		page.MessageType = "success"
	} else {
		page.Message = ""
	}

	/*
		// DEBUG:
		if !page.HasEntry {
			page.HasEntry = true
			entry := database.FetchByID(1)

			page.ExtDXP = portalExt(entry, true)
			page.Ext62 = portalExt(entry, false)
		}
	*/
	session.Save(r, w)

	if pages[0] == "browse" && page.HasMountedFolder {
		loc := mux.Vars(r)["loc"]

		files, err := brwsr.List(loc)
		if err != nil {
			session.AddFlash("Failed listing folder", "fail")
		}

		page.FileList = files
	}

	if pages[0] == "srvimport" {
		dumploc := r.URL.Query().Get("dump")

		page.DumpLoc = dumploc
	}

	if pages[0] == "home" {
		pages = append(pages, "databases")

		privateDBs, err := database.FetchByCreator(page.User)
		if err != nil {
			log.Printf("couldn't list databases: %s", err.Error())
		}

		if len(privateDBs) != 0 {
			page.PrivateDatabases = privateDBs
			page.HasPrivateDBs = true
		}

		publicDBs, err := database.FetchPublic()
		if err != nil {
			log.Printf("couldn't list databases: %s", err.Error())
		}

		if len(publicDBs) != 0 {
			page.PublicDatabases = publicDBs
			page.HasPublicDBs = true
		}
	}

	toLoad := []string{"base", "nav", "connectors", "properties"}
	toLoad = append(toLoad, pages...)

	tmpl, err := buildTemplate(toLoad...)
	if err != nil {
		panic(err)
	}

	err = tmpl.ExecuteTemplate(w, "base", page)
	if err != nil {
		panic(err)
	}
}

func buildTemplate(pages ...string) (*template.Template, error) {
	var templates []string
	for _, page := range pages {
		templates = append(templates, fmt.Sprintf("web/html/%s.html", page))
	}

	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates failed: %s", err.Error())
	}

	return tmpl, nil
}

func getTitle(page string) string {
	title, ok := getPages()[page]
	if ok {
		return title
	}

	if strings.HasPrefix(page, "/browse") {
		return "Server Browser"
	}

	switch page {
	case "/fileimport":
		return "Import Database"
	default:
		return "Unknown"
	}

}

func getPages() map[string]string {
	pages := make(map[string]string)

	pages["/"] = "Home"
	pages["/createdb"] = "Create database"
	pages["/importdb"] = "Import database"
	return pages
}
