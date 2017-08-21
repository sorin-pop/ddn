package main

import (
	"fmt"
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
	dumps := http.StripPrefix("/dumps/", http.FileServer(http.Dir(fmt.Sprintf("%s/web/dumps/", workdir))))
	router.PathPrefix("/dumps/").Handler(dumps)

	// Add static serving of images / css / js from res directory.
	res := http.StripPrefix("/res/", http.FileServer(http.Dir(fmt.Sprintf("%s/web/res", workdir))))
	router.PathPrefix("/res/").Handler(res)

	// Serve node_modules folder as well
	nodeModules := http.StripPrefix("/node_modules/", http.FileServer(http.Dir(fmt.Sprintf("%s/web/node_modules", workdir))))
	router.PathPrefix("/node_modules/").Handler(nodeModules)

	return router
}
