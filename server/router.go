package main

import (
	"net/http"

	"github.com/djavorszky/ddn/common/srv"
	"github.com/gorilla/mux"
)

// Router creates a new router that registers all routes.
func Router() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = srv.Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	// Add static serving of files in dumps directory.
	dumps := http.StripPrefix("/dumps/", http.FileServer(http.Dir("./web/dumps/")))
	router.PathPrefix("/dumps/").Handler(dumps)

	// Add static serving of images / css / js from res directory.
	res := http.StripPrefix("/res/", http.FileServer(http.Dir("./web/res")))
	router.PathPrefix("/res/").Handler(res)

	// Serve node_modules folder as well
	nodeModules := http.StripPrefix("/node_modules/", http.FileServer(http.Dir("./web/node_modules")))
	router.PathPrefix("/node_modules/").Handler(nodeModules)

	return router
}
