package registry

import (
	"sort"
	"sync"

	"github.com/djavorszky/ddn/common/model"
)

var (
	curID    = 0
	ids      = make(chan int)
	registry = make(map[string]model.Connector)

	rw sync.RWMutex
)

func init() {
	go inc()
}

// Store registers the connector in the registry, or overwrites
// if connector already in.
func Store(conn model.Connector) {
	rw.Lock()
	registry[conn.ShortName] = conn
	rw.Unlock()
}

// Get returns the connector associated with the shortName, or
// an error if no connectors are registered with that name
func Get(shortName string) (model.Connector, bool) {
	rw.RLock()
	conn, ok := registry[shortName]
	rw.RUnlock()

	return conn, ok
}

// Remove removes the connector added with shortName. Does not error
// if connector not in registry.
func Remove(shortName string) {
	rw.Lock()
	delete(registry, shortName)
	rw.Unlock()
}

// List returns the list of connectors as a slice
func List() []model.Connector {
	var conns []model.Connector

	rw.RLock()
	for _, c := range registry {
		conns = append(conns, c)
	}
	rw.RUnlock()

	sort.Sort(ByName(conns))

	return conns
}

// Exists checks the registry for the existence of
// an entry registered with the supplied shortName
func Exists(shortName string) bool {
	rw.RLock()
	_, ok := registry[shortName]
	rw.RUnlock()

	return ok
}

// ID returns a new ID that is unique
func ID() int {
	return <-ids
}

func inc() int {
	for {
		curID++
		ids <- curID
	}
}

// ByName implements sort.Interface for []model.Connector based on
// the ShortName field
type ByName []model.Connector

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].ShortName < a[j].ShortName }
