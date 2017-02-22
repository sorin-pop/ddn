package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/djavorszky/ddn/common/model"
)

// Page is a struct holding the data to be displayed on the welcome page.
type Page struct {
	Connectors *map[string]model.Connector
	AnyOnline  bool
}

func displayWelcomePage(w http.ResponseWriter, r *http.Request) {
	debug := r.URL.Query().Get("debug")

	page := buildPage(debug)
	tmpl, err := buildTemplate()
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, page)
	if err != nil {
		panic(err)
	}
}

func buildPage(debug string) Page {
	if debug == "" {
		return Page{Connectors: &registry, AnyOnline: len(registry) > 0}
	}

	conns := make(map[string]model.Connector)

	conns["mysql-55"] = model.Connector{
		ID:            1,
		DBVendor:      "mysql",
		DBPort:        "3306",
		ShortName:     "mysql-55",
		LongName:      "mysql 5.5.57",
		Identifier:    "dbcloud-mysql-55",
		ConnectorPort: "6000",
		Version:       "0.7.0",
		Address:       "127.0.0.1",
		Up:            true,
	}
	conns["mysql-56"] = model.Connector{
		ID:            2,
		DBVendor:      "mysql",
		DBPort:        "3306",
		ShortName:     "mysql-56",
		LongName:      "mysql 5.6.63",
		Identifier:    "dbcloud-mysql-56",
		ConnectorPort: "6000",
		Version:       "0.7.0",
		Address:       "127.0.0.1",
		Up:            true,
	}
	conns["mysql-57"] = model.Connector{
		ID:            3,
		DBVendor:      "mysql",
		DBPort:        "3306",
		ShortName:     "mysql-57",
		LongName:      "mysql 5.7.15",
		Identifier:    "dbcloud-mysql-57",
		ConnectorPort: "6000",
		Version:       "0.7.0",
		Address:       "127.0.0.1",
		Up:            true,
	}
	conns["oracle-11g"] = model.Connector{
		ID:            4,
		DBVendor:      "oracle",
		DBPort:        "1521",
		ShortName:     "oracle-11g",
		LongName:      "oracle 11.0.2.1",
		Identifier:    "dbcloud-oracle-11g",
		ConnectorPort: "6000",
		Version:       "0.7.0",
		Address:       "127.0.0.1",
		Up:            true,
	}

	return Page{Connectors: &conns, AnyOnline: len(conns) > 0}
}

func buildTemplate() (*template.Template, error) {
	files, err := ioutil.ReadDir("web/html")
	if err != nil {
		return nil, fmt.Errorf("reading html directory failed: %s", err.Error())
	}

	var templates []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		templates = append(templates, fmt.Sprintf("web/html/%s", file.Name()))
	}

	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates failed: %s", err.Error())
	}

	return tmpl, nil
}
