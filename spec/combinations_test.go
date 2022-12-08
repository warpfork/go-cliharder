package spec

import (
	"fmt"
	"testing"
	// qt "github.com/frankban/quicktest"
)

func TestCombinations(t *testing.T) {
	cases := []struct {
		spec     string
		args     string
		expected []error
	}{
		{"[-a]", "-a", nil},
		{"[-a]", "", nil},
		{"[-ab]", "-a", nil},
		{"[-ab]", "-b", nil},
		{"[-ab]", "-ab", nil},
		{"(-a | -b)", "-a", nil},
		{"(-a | -b)", "-b", nil},
		{"(-a | -b)", "-ab", []error{fmt.Errorf("cannot combine -a and -b")}},
		{"(-a | -b | -c)", "-ac", []error{fmt.Errorf("cannot combine -a and -c")}},
		{"(-a | -b | -c)", "-abc", []error{
			fmt.Errorf("cannot combine -a and -c"),
			fmt.Errorf("cannot combine -a and -b"),
			fmt.Errorf("cannot combine -b and -c"),
		}},
	}

	for _, c := range cases {
		_ = c
	}
}
