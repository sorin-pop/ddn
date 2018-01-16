package registry

import (
	"sort"
	"sync"

	"github.com/djavorszky/ddn/common/model"
)

var (
	curID    = 0
	ids      = make(chan int)
	registry = make(map[string]model.Agent)

	rw sync.RWMutex
)

func init() {
	go inc()
}

// Store registers the agent in the registry, or overwrites
// if agent already in.
func Store(agent model.Agent) {
	rw.Lock()
	registry[agent.ShortName] = agent
	rw.Unlock()
}

// Get returns the agent associated with the shortName, or
// an error if no agent are registered with that name
func Get(shortName string) (model.Agent, bool) {
	rw.RLock()
	agent, ok := registry[shortName]
	rw.RUnlock()

	return agent, ok
}

// Remove removes the agent added with shortName. Does not error
// if agent not in registry.
func Remove(shortName string) {
	rw.Lock()
	delete(registry, shortName)
	rw.Unlock()
}

// List returns the list of agents as a slice
func List() []model.Agent {
	var agents []model.Agent

	rw.RLock()
	for _, c := range registry {
		agents = append(agents, c)
	}
	rw.RUnlock()

	sort.Sort(ByName(agents))

	return agents
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

// ByName implements sort.Interface for []model.Agent based on
// the ShortName field
type ByName []model.Agent

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].ShortName < a[j].ShortName }
