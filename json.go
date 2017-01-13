package main

import (
	"encoding/json"
	"log"
)

// JSONMessage is an interface that can hold many types of messages that
// can be json'ified. The reason we need to return an int as well is because
// I can't figure out how we could easily get the status of the JSONMessage
// without too much boilerplate code. This way, we can return the status
// in the same step and, if not needed, discard it.
type JSONMessage interface {
	Compose() ([]byte, int)
}

// Message is a struct to hold a simple status-message type response
type Message struct {
	Status  int
	Message string
}

// Compose creates a JSON formatted byte slice from the Message
func (msg Message) Compose() ([]byte, int) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	return b, msg.Status
}

// ListMessage is a struct to hold a status-list of strings type response
type ListMessage struct {
	Status  int
	Message []string
}

// Compose creates a JSON formatted byte slice from the ListMessage
func (msg ListMessage) Compose() ([]byte, int) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	return b, msg.Status
}

// MapMessage is a struct to hold a status and a key+value type response
type MapMessage struct {
	Status  int
	Message map[string]string
}

// Compose creates a JSON formatted byte slice from the Message
func (msg MapMessage) Compose() ([]byte, int) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	return b, msg.Status
}
