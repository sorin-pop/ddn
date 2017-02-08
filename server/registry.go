package main

import "github.com/djavorszky/ddn/common/model"

var (
	id     = 0
	idChan = make(chan int, 1)

	registry map[string]model.Connector
)

// For now, only initialize the map.
func initRegistry() {
	registry = make(map[string]model.Connector)

	go increment()
}

func getID() int {
	return <-idChan
}

// Increment should ALWAYS be used in a seperate goroutine! Otherwise,
// the server will hang.
func increment() int {
	for {
		id++
		idChan <- id
	}
}
