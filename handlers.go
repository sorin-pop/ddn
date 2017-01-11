package main

import (
	"fmt"
	"net/http"
)

// ListDatabase lists the supervised databases in a JSON format
func ListDatabase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

// Index should display whenever someone visits the main page.
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the index!")
}
