package main

// Message is a struct to hold a simple status-message type response
type Message struct {
	Status  int
	Message string
}

// ListMessage is a struct to hold a status-list of strings type response
type ListMessage struct {
	Status  int
	Message []string
}

// MapMessage is a struct to hold a status and a key+value type response
type MapMessage struct {
	Status  int
	Message map[string]string
}
