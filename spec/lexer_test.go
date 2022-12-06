package spec

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestTokenize(t *testing.T) {
	cases := []struct {
		usage    string
		expected []Token
	}{
		{"<arg>", []Token{{TTArg, "<arg>", 0}}},
		{"<arg42>", []Token{{TTArg, "<arg42>", 0}}},
		{"<arg_extra>", []Token{{TTArg, "<arg_extra>", 0}}},

		{"<arg1> <arg2>", []Token{{TTArg, "<arg1>", 0}, {TTArg, "<arg2>", 7}}},
		{"<arg1>  <arg2>", []Token{{TTArg, "<arg1>", 0}, {TTArg, "<arg2>", 8}}},

		{"(", []Token{{TTOpenGroup, "(", 0}}},
		{")", []Token{{TTCloseGroup, ")", 0}}},

		{"(<arg>)", []Token{{TTOpenGroup, "(", 0}, {TTArg, "<arg>", 1}, {TTCloseGroup, ")", 6}}},
		{"( <arg> )", []Token{{TTOpenGroup, "(", 0}, {TTArg, "<arg>", 2}, {TTCloseGroup, ")", 8}}},

		{"[<arg>]", []Token{{TTOpenOptional, "[", 0}, {TTArg, "<arg>", 1}, {TTCloseOptional, "]", 6}}},
		{"[ <arg> ]", []Token{{TTOpenOptional, "[", 0}, {TTArg, "<arg>", 2}, {TTCloseOptional, "]", 8}}},
		{"<arg> [<arg2> ]", []Token{{TTArg, "<arg>", 0}, {TTOpenOptional, "[", 6}, {TTArg, "<arg2>", 7}, {TTCloseOptional, "]", 14}}},
		{"<arg> [ <arg2>]", []Token{{TTArg, "<arg>", 0}, {TTOpenOptional, "[", 6}, {TTArg, "<arg2>", 8}, {TTCloseOptional, "]", 14}}},

		{"...", []Token{{TTElipsis, "...", 0}}},
		{"<arg>...", []Token{{TTArg, "<arg>", 0}, {TTElipsis, "...", 5}}},
		{"<arg> ...", []Token{{TTArg, "<arg>", 0}, {TTElipsis, "...", 6}}}, // TODO I don't think this space should be tolerated.
		{"[<arg>...]", []Token{{TTOpenOptional, "[", 0}, {TTArg, "<arg>", 1}, {TTElipsis, "...", 6}, {TTCloseOptional, "]", 9}}},

		{"|", []Token{{TTChoice, "|", 0}}},
		{"<arg>|<arg2>", []Token{{TTArg, "<arg>", 0}, {TTChoice, "|", 5}, {TTArg, "<arg2>", 6}}},
		{"<arg> |<arg2>", []Token{{TTArg, "<arg>", 0}, {TTChoice, "|", 6}, {TTArg, "<arg2>", 7}}},
		{"<arg>| <arg2>", []Token{{TTArg, "<arg>", 0}, {TTChoice, "|", 5}, {TTArg, "<arg2>", 7}}},

		{"-p", []Token{{TTShortOpt, "-p", 0}}},
		{"-X", []Token{{TTShortOpt, "-X", 0}}},

		{"--force", []Token{{TTLongOpt, "--force", 0}}},
		{"--sig-proxy", []Token{{TTLongOpt, "--sig-proxy", 0}}},

		{"-aBc", []Token{{TTOptSeq, "aBc", 0}}},
		{"--", []Token{{TTDoubleDash, "--", 0}}},
		{"=<bla>", []Token{{TTOptValue, "=<bla>", 0}}},
		{"=<bla-bla>", []Token{{TTOptValue, "=<bla-bla>", 0}}},
		{"=<bla--bla>", []Token{{TTOptValue, "=<bla--bla>", 0}}},
		{"-p=<file-path>", []Token{{TTShortOpt, "-p", 0}, {TTOptValue, "=<file-path>", 2}}},
		{"--path=<absolute-path>", []Token{{TTLongOpt, "--path", 0}, {TTOptValue, "=<absolute-path>", 6}}},
	}
	for _, c := range cases {
		t.Run(c.usage, func(t *testing.T) {
			tks, err := Tokenize(c.usage)
			if err != nil {
				t.Errorf("[Tokenize '%s']: Unexpected error: %v", c.usage, err)
				return
			}

			//t.Logf("actual: %v\n", tks)
			if len(tks) != len(c.expected) {
				t.Errorf("[Tokenize '%s']: token count mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, c.expected, tks)
				return
			}

			for i, actual := range tks {
				expected := c.expected[i]
				switch {
				case actual.Typ != expected.Typ:
					t.Errorf("[Tokenize '%s']: token type mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
				case actual.Val != expected.Val:
					t.Errorf("[Tokenize '%s']: token text mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
				case actual.Pos != expected.Pos:
					t.Errorf("[Tokenize '%s']: token pos mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
				}
			}
		})
	}
}

func TestTokenizeErrors(t *testing.T) {
	cases := []*ParseError{
		{".", 1, "Unexpected end of usage, was expecting rest of elipsis"},
		{"..", 2, "Unexpected end of usage, was expecting rest of elipsis"},
		{"A.", 0, "Unexpected input"},
		{"<ARG>..", 7, "Unexpected end of usage, was expecting rest of elipsis"},
		{"<ARG>..x", 7, "Invalid syntax: elipsis must be composed of three subsequent periods"},
		{"-", 1, "Unexpected end of usage, was expecting an option name"},
		{"---x", 2, "Was expecting a long option name"},
		{"-x-", 2, "Invalid syntax: cannot have dashes in middle of a short opt group"},

		// The rest of this next block arguably shouldnt be accepted unless it's following an opt,
		//  but the lexer is just finding token edges, not going so far as to parse to make sure the token sequences make any sense.
		//  In some cases where it's a linear nonrecursive/nonstack thing to recognize, AND it produces a discrete token, the lexer bothers;
		//   in this case, the second condition is false.
		{"=", 1, "Unexpected end of usage, was expecting '=<'"},
		{"=<", 2, "Unclosed option value"},
		{"=<dsdf", 6, "Unclosed option value"},
		{"=<>", 2, "Was expecting an option value"},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			_, err := Tokenize(c.Input)
			qt.Assert(t, err, qt.CmpEquals(), c)
		})
	}
}
