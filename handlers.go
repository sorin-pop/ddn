package main

import (
	"fmt"
	"net/http"
)

// index should display whenever someone visits the main page.
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the index!")
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to createDatabase")
}

// listDatabase lists the supervised databases in a JSON format
func listDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to listDatabase")
}

// getDatabase will get a specific database with a specific name
func getDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to getDatabase")
}

// dropDatabase will drop the named database with its tablespace and user
func dropDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to dropDatabase")
}

// importDatabase will import the specified dumpfile to the database
// creating the database, tablespace and user as needed
func importDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to importDatabase")
}

func whoami(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to whoami!")
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to heartbeat!")
}
