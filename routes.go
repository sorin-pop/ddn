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
		"Index",
		"GET",
		"/",
		Index,
	},
	route{
		"ListDatabase",
		"GET",
		"/list-databases",
		ListDatabase,
	},
}
