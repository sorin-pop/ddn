package main

import (
	"net/http"
)

func writeHeader(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}
