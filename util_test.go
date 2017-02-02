package main

import "testing"

type testcase struct {
	Value    string
	Search   string
	Expected bool
}

func TestIContains(t *testing.T) {
	pairs := []testcase{
		// check exact matches with varying cases
		testcase{"hello", "hello", true},
		testcase{"hello", "HELLO", true},
		testcase{"HELLO", "hello", true},
		testcase{"heLLo", "hello", true},
		testcase{"hello", "HeLLo", true},
		// check with space
		testcase{"HELLO there", "hello", true},
		testcase{"heLLo there", "lo th", true},
		testcase{"hello there", "LO TH", true},
		// check submatches with varying cases
		testcase{"hello", "el", true},
		testcase{"hello", "EL", true},
		testcase{"hello", "eL", true},
		// check false hits
		testcase{"hello", "oh", false},
		testcase{"hello", "OH", false},
		testcase{"hello", "oH", false},
		testcase{"this is", "ss", false},
		testcase{"this is", "sS", false},
		testcase{"this is", "SS", false},
		testcase{"", "", false},
		testcase{"hello", "", false},
		testcase{"", "hello", false},
	}

	for _, c := range pairs {
		if res := iContains(c.Value, c.Search); res != c.Expected {
			t.Errorf("Mismatch. Expected: %v, got: %v for testcase (%q, %q)", c.Expected, res, c.Value, c.Search)
		}
	}
}

func BenchmarkIContainsFound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		iContains("This is a rather long line and I'm curious whether that thing is in there or not.", "hat")
	}
}

func BenchmarkIContainsNotFound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		iContains("This is a rather long line and I'm curious whether that thing is in there or not.", "moo")
	}
}

func BenchmarkIContainsShorted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		iContains("This is a rather long line and I'm curious whether that thing is in there or not.", "")
	}
}
