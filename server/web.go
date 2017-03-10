package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/djavorszky/ddn/common/model"
)

// Page is a struct holding the data to be displayed on the welcome page.
type Page struct {
	Connectors  *map[string]model.Connector
	AnyOnline   bool
	Title       string
	Pages       map[string]string
	ActivePage  string
	Message     string
	MessageType string
	HasEntry    bool
	Databases   []model.DBEntry
	Ext62       model.PortalExt
	ExtDXP      model.PortalExt
}

func loadPage(w http.ResponseWriter, r *http.Request, pages ...string) {
	page := Page{
		Connectors: &registry,
		AnyOnline:  len(registry) > 0,
		Title:      getTitle(r.URL.Path),
		Pages:      getPages(),
		ActivePage: r.URL.Path,
	}

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if flashes := session.Flashes("success"); len(flashes) > 0 {
		page.Message = flashes[0].(string)
		page.MessageType = "success"

		id := session.Values["id"].(int64)

		page.HasEntry = true
		entry := db.entryByID(id)

		page.ExtDXP = portalExt(entry, true)
		page.Ext62 = portalExt(entry, false)

	} else if flashes := session.Flashes("fail"); len(flashes) > 0 {
		page.Message = flashes[0].(string)
		page.MessageType = "danger"
	} else if flashes := session.Flashes("debug"); len(flashes) > 0 {
		page.Message = flashes[0].(string)
		page.MessageType = "success"
	} else {
		page.Message = ""
	}

	/*
		// DEBUG:
		if !page.HasEntry {
			page.HasEntry = true
			entry := db.entryByID(1)

			page.ExtDXP = portalExt(entry, true)
			page.Ext62 = portalExt(entry, false)
		}
	*/

	session.Save(r, w)

	toLoad := []string{"base", "head", "nav", "connectors", "properties"}
	toLoad = append(toLoad, pages...)

	tmpl, err := buildTemplate(toLoad...)
	if err != nil {
		panic(err)
	}

	page.Databases, _ = db.list()

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
	return getPages()[page]
}

func getPages() map[string]string {
	pages := make(map[string]string)

	pages["/"] = "Home"
	pages["/createdb"] = "Create database"
	pages["/importdb"] = "Import database"

	return pages
}
