package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/djavorszky/ddn/common/model"
)

// Page is a struct holding the data to be displayed on the welcome page.
type Page struct {
	Connectors *map[string]model.Connector
	AnyOnline  bool
	Title      string
	Pages      map[string]string
	ActivePage string
}

func loadPage(w http.ResponseWriter, r *http.Request, pages ...string) {
	page := buildPage(r.URL.Path)
	page.ActivePage = r.URL.Path

	toLoad := []string{"base", "head", "nav", "connectors"}
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

func buildPage(activePage string) Page {
	p := Page{
		Connectors: &registry,
		AnyOnline:  len(registry) > 0,
		Title:      getTitle(activePage),
		Pages:      getPages(),
		ActivePage: activePage,
	}

	return p

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
