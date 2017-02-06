package main

import (
	"sync"

	"github.com/djavorszky/ddn/common/model"
)

var registry map[string]model.Connector

var (
	id     = 0
	idChan = make(chan int, 1)
	mutex  = &sync.Mutex{}
)

// For now, only initialize the map. Later on there will be other
// tasks as well, such as reading up all the Connectors from the
// database, checking each one that has "UP" set to true and keeping
// the ones that are still alive in the memory, or initializing the
// 'id' value to the next value
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
