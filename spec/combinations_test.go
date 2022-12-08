package spec

import (
	"fmt"
	"testing"
	// qt "github.com/frankban/quicktest"
)

type combinationError struct {
	// one or none of:
	SpecError     error
	MatchingError []error
}

func TestCombinations(t *testing.T) {
	SpecError := func(e error) combinationError {
		return combinationError{SpecError: e}
	}
	MatchingError := func(e ...error) combinationError {
		return combinationError{MatchingError: e}
	}
	Acceptable := combinationError{}
	_ = SpecError // not yet used -- but there's a good number of constructions that will pass the lexer but we should reject at another level.

	cases := []struct {
		spec     string
		args     string
		expected combinationError
	}{
		// All errors in the cases below are currently phrased as strings, but we may structuralize this to improve the tests later.  TBD.  (Seeing prose in fixtures can be good too.)
		// When there are multiple errors, the order is significant.  (A user probably wouldn't care if they were permuted, but we care about stability.)
		//   In general, this order is the product of "march across the arguments in the order provided; report their constraint satisfaction failures in the order that the constraints are found in the spec".
		// TODO: There's a few cases below where there might be additional errors that are the mirror of the ones currently listed (particularly around exclusive multiple choices).  I'm hoping we can construct "leftwards only" conditions there during spec parse time, but not sure how complex that will be; if it's complex, it may not be worth bothering with.
		{"[-a]", "-a", Acceptable},
		{"[-a]", "", Acceptable},
		{"[-ab]", "-a", Acceptable},
		{"[-ab]", "-b", Acceptable},
		{"[-ab]", "-ab", Acceptable},
		{"(-a | -b)", "-a", Acceptable},
		{"(-a | -b)", "-b", Acceptable},
		{"(-a | -b)", "-ab", MatchingError(fmt.Errorf("When '-a' is present, '-b' is conflicting and not allowed"))},
		{"(-a | -b | -c)", "-ac", MatchingError(fmt.Errorf("When '-a' is present, '-c' is conflicting and not allowed"))},
		{"(-a | -b | -c)", "-abc", MatchingError(
			fmt.Errorf("When '-a' is present, '-b' is conflicting and not allowed"),
			fmt.Errorf("When '-a' is present, '-c' is conflicting and not allowed"),
			fmt.Errorf("When '-b' is present, '-c' is conflicting and not allowed"),
		)},
		{"-a [-b]", "-b", MatchingError(fmt.Errorf("'-a' is required"))},
		{"-a [-b]", "-ab", Acceptable},
		{"[-a -b]", "-a", MatchingError(fmt.Errorf("When '-a' is present, '-b' is required"))},
		{"[-a -b]", "-b", MatchingError(fmt.Errorf("When '-b' is present, '-a' is required"))},
		{"[-a -b -c]", "-b", MatchingError( // Observe how we decompose this to a series of pairwise mismatch reports, rather than trying to build more complex english grammar in the reports.
			fmt.Errorf("When '-b' is present, '-a' is required"),
			fmt.Errorf("When '-b' is present, '-c' is required"),
		)},
		{"[-a -b -c]", "-bc", MatchingError(
			fmt.Errorf("When '-b' is present, '-a' is required"),
			fmt.Errorf("When '-c' is present, '-a' is required"),
		)},
		{"[-a -b -c -d]", "-bc", MatchingError( // Reports in this form of error can get a little long.  Still, this is the price to pay for clarity -- and real-world CLIs don't often combine many requirements in a row like this.
			fmt.Errorf("When '-b' is present, '-a' is required"),
			fmt.Errorf("When '-b' is present, '-d' is required"),
			fmt.Errorf("When '-c' is present, '-a' is required"),
			fmt.Errorf("When '-c' is present, '-d' is required"),
		)},
		{"[-a [-b]]", "-b", MatchingError(fmt.Errorf("'-b' is only an option when '-a' is present"))},
		{"[(-a|-b) [-c]]", "-c", MatchingError(fmt.Errorf("'-c' is only an option when '(-a|-b)' is satisfied"))}, // REVIEW: this is the one composition I both find defensible and yet really can't find a good way to simplify to a purely pairwise report.
		// "[(-a|-b) [-c [-d]]]" adds some more details but the hard part is similar to above.
		// "[(-a|-b) [-c -d]]" adds some more details but the hard part is similar to above.
	}

	for _, c := range cases {
		_ = c
	}
}
