package registry

import (
	"testing"

	"sort"

	"github.com/djavorszky/ddn/common/model"
)

const (
	name1 = "firstConn"
	long1 = "this is a longer string for one"

	name2 = "secondConn"
	long2 = "this is a longer string for two"

	name3 = "thirdConn"
	long3 = "this is a longer string for three"

	missing = "nonexistent"
)

var (
	c1 = model.Connector{ShortName: name1, LongName: long1}
	c2 = model.Connector{ShortName: name2, LongName: long2}
	c3 = model.Connector{ShortName: name3, LongName: long3}
)

func setup() {
	registry[name1] = c1
	registry[name2] = c2
	registry[name3] = c3
}

func teardown() {
	delete(registry, name1)
	delete(registry, name2)
	delete(registry, name3)
}

func TestGet(t *testing.T) {
	setup()
	defer teardown()

	_, ok := Get(missing)
	if ok {
		t.Errorf("Get(%q) returned no error", missing)
	}

	c, ok := Get(name1)
	if !ok {
		t.Errorf("Get(%q) returned false", name1)
	}

	if c.ShortName != c1.ShortName || c.LongName != c1.LongName {
		t.Errorf("Get(%q) returned wrong connector. Expected: %v, Got: %v", name1, c, c1)
	}

	c, ok = Get(name2)
	if !ok {
		t.Errorf("Get(%q) returned false", name2)
	}

	if c.ShortName != c2.ShortName || c.LongName != c2.LongName {
		t.Errorf("Get(%q) returned wrong connector. Expected: %v, Got: %v", name1, c, c2)
	}
}

func TestRemove(t *testing.T) {
	setup()
	defer teardown()

	if _, ok := registry[name1]; !ok {
		t.Errorf("Test error: %q not in registry", name1)
	}

	Remove(name1)

	if _, ok := registry[name1]; ok {
		t.Errorf("Remove(%q) did not remove connector", name1)
	}
}

func TestStore(t *testing.T) {
	setup()
	defer teardown()

	delete(registry, name1)

	if _, ok := registry[name1]; ok {
		t.Errorf("Test error: %q still in registry", name1)
	}

	Store(c1)

	if _, ok := registry[name1]; !ok {
		t.Errorf("Store(%q) did not store connector", name1)
	}
}

func TestList(t *testing.T) {
	setup()
	defer teardown()

	conns := List()

	if len(conns) != 3 {
		t.Errorf("Connector count not correct")
	}
}

func TestExists(t *testing.T) {
	setup()
	defer teardown()

	if !Exists(name1) {
		t.Errorf("Exists(%q) = false", name1)
	}

	if !Exists(name2) {
		t.Errorf("Exists(%q) = false", name2)
	}

	if !Exists(name3) {
		t.Errorf("Exists(%q) = false", name3)
	}

	if Exists(missing) {
		t.Errorf("Exists(%q) = true", missing)
	}
}

func TestID(t *testing.T) {
	var id int

	for i := 1; i < 12; i++ {
		id = ID()

		if id != i {
			t.Errorf("ID() = '%d', should be '%d'", id, i)
		}
	}
}

func TestSort(t *testing.T) {
	var list = []model.Connector{c3, c2, c1}

	if list[0].ShortName < list[1].ShortName && list[1].ShortName < list[2].ShortName {
		t.Errorf("List is sorted to begin with")
	}

	sort.Sort(ByName(list))

	if list[0].ShortName > list[1].ShortName || list[1].ShortName > list[2].ShortName {
		t.Errorf("List have not been sorted")
	}

}
