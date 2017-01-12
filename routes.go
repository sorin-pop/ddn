package main

import "net/http"

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes contains all available routes.
type Routes []route

var routes = Routes{
	route{
		"index",
		"GET",
		"/",
		index,
	},
	route{
		"createDatabase",
		"GET",
		"/create-database",
		createDatabase,
	},
	route{
		"listDatabase",
		"GET",
		"/list-databases",
		listDatabase,
	},
	route{
		"getDatabase",
		"GET",
		"/get-database",
		getDatabase,
	},
	route{
		"dropDatabase",
		"GET",
		"/drop-database",
		dropDatabase,
	},
	route{
		"importDatabase",
		"GET",
		"/import-database",
		importDatabase,
	},
	route{
		"whoami",
		"GET",
		"/whoami",
		whoami,
	},
	route{
		"heartbeat",
		"GET",
		"/heartbeat",
		heartbeat,
	},
}
