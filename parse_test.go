package cliharder

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParse(t *testing.T) {
	qt.Assert(t, ParseArgsSpool([]string{"oof", "dah"}), qt.CmpEquals(),
		ArgsSpool{
			Positional: []string{"oof", "dah"},
		})

	qt.Assert(t, ParseArgsSpool([]string{"oof", "--dah"}), qt.CmpEquals(),
		ArgsSpool{
			Positional: []string{"oof"},
			LongOpts:   map[string][]string{"dah": []string{""}},
		})

	qt.Assert(t, ParseArgsSpool([]string{"oof", "--dah=may"}), qt.CmpEquals(),
		ArgsSpool{
			Positional: []string{"oof"},
			LongOpts:   map[string][]string{"dah": []string{"may"}},
		})

	qt.Assert(t, ParseArgsSpool([]string{"oof", "--dah=may", "--dah=my"}), qt.CmpEquals(),
		ArgsSpool{
			Positional: []string{"oof"},
			LongOpts:   map[string][]string{"dah": []string{"may", "my"}},
		})

	qt.Assert(t, ParseArgsSpool([]string{"one", "--two=three", "-avv", "four"}), qt.CmpEquals(),
		ArgsSpool{
			Positional: []string{"one", "four"},
			LongOpts:   map[string][]string{"two": []string{"three"}},
			ShortOpts:  map[string]int{"a": 1, "v": 2},
		})

	qt.Assert(t, ParseArgsSpool([]string{"one", "--two=three", "--", "-avv", "four"}), qt.CmpEquals(),
		ArgsSpool{
			Positional:  []string{"one"},
			LongOpts:    map[string][]string{"two": []string{"three"}},
			Positional2: []string{"-avv", "four"},
		})
}
