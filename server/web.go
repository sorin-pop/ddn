package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/djavorszky/ddn/common/model"
)

// Page is a struct holding the data to be displayed on the welcome page.
type Page struct {
	Connectors    *map[string]model.Connector
	AnyOnline     bool
	Title         string
	Pages         map[string]string
	ActivePage    string
	Message       string
	MessageType   string
	Property      DBEntry
	HasProperties bool
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
		log.Println("success flash.")
		page.Message = flashes[0].(string)
		page.MessageType = "success"

		var dbentry *DBEntry

		val := session.Values["dbentry"]
		dbentry, ok := val.(*DBEntry)
		if ok {
			page.HasProperties = true
			page.Property = *dbentry
		}
	} else if flashes := session.Flashes("fail"); len(flashes) > 0 {
		log.Println("danger flash.")

		page.Message = flashes[0].(string)
		page.MessageType = "danger"
	} else {
		page.Message = ""
	}

	session.Save(r, w)

	toLoad := []string{"base", "head", "nav", "connectors", "properties"}
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
	return getPages()[page]
}

func getPages() map[string]string {
	pages := make(map[string]string)

	pages["/"] = "Home"
	pages["/createdb"] = "Create database"

	return pages
}
