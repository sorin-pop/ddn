package main

import "testing"

type testcase struct {
	Value    string
	Search   string
	Expected bool
}

func TestIContains(t *testing.T) {
	pairs := []testcase{
		testcase{"hello", "hello", true},
		testcase{"hello", "HELLO", true},
		testcase{"HELLO", "hello", true},
		testcase{"heLLo", "hello", true},
		testcase{"hello", "HeLLo", true},
		testcase{"hello", "", false},
		testcase{"", "hello", false},
		testcase{"", "", false},
	}

	for _, c := range pairs {
		if res := iContains(c.Value, c.Search); res != c.Expected {
			t.Errorf("Mismatch. Expected: %v, got: %v for testcase (%q, %q)", c.Expected, res, c.Value, c.Search)
		}
	}
}
