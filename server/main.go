package main

import (
	"log"
	"net/http"
)

func main() {
	initRegistry()

	port := ":7010"

	log.Println("Starting...")
	log.Fatal(http.ListenAndServe(port, Router()))
}
