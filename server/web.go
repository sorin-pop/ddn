package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/djavorszky/ddn/common/model"
)

type Page struct {
	Connectors *map[string]model.Connector
}

func displayWelcomePage(w http.ResponseWriter) {
	page := buildPage()
	tmpl, err := buildTemplate()
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, page)
	if err != nil {
		panic(err)
	}
}

func buildPage() Page {
	return Page{Connectors: &registry}
}

func buildTemplate() (*template.Template, error) {
	files, err := ioutil.ReadDir("tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading tmpl directory failed: %s", err.Error())
	}

	var templates []string
	for _, file := range files {
		templates = append(templates, fmt.Sprintf("tmpl/%s", file.Name()))
	}

	fmt.Println(templates)

	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates failed: %s", err.Error())
	}

	return tmpl, nil
}
