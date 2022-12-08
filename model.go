package cliharder

import (
	"context"
	"io"
)

// Code in this file is the core structures for describing commands and their options.

// TODO: Everything below is a sketch, so far.

// TODO: Value-binding structures aren't here.  There should be some for each `*ParamSpec` type.

type Application struct {
	Context context.Context // Just carried through, for your use if necessary.
	Name    string          // May be the empty string, in which case `args[0]` will be used instead wherever this would have been used.
	Stdout  io.Writer
	Stderr  io.Writer
	Command *Command // Root of the subcommand tree.  The `.Name` value of the root node is ignored.
}

type Command struct {
	Name    string
	AliasOf []string // If present, no other fields are valid -- this is a redirection.  The AliasOf list starts at the root of the application (it can alias even to non-sibling subcommands).
	// REVIEW: turn that ^ into just a string, and split on space?  Easier to use.

	// Modes do the bulk of the lifting.
	// You'll find the args specs and the action callbacks in here.
	//
	// This may be a list of one element (which will have no conditions nor names, etc) for simple commands,
	// in which case always that mode and its args specs and action are what occurs;
	// or it may contain a list of modes,
	// in which case processing the modes conditions is the first step in figuring out
	// which further args specs should be validated and which action may follow.
	//
	// As a special case: if there's a single mode, with no conditions,
	// and it also has no action callback,
	// and subcommands are present,
	// then by default ths library will generate a fallback action which emits a summary of subcommand usages.
	Modes []CommandMode

	// Every command may optionally have subcommands.
	// Subcommands may coexist with the command itself having actions
	// (or, not -- in which case, by default ths library will generate a fallback action which emits a summary of subcommand usages).
	//
	// Subcommands are recognized by further positional arguments which match the subcommand names.
	// Subcommand matches for positionals always have precidence over any positions in the command arguments.
	// (Using positional arguments in a command that has subcommands is inadvisable.  It's technically possible, but likely to be unclear and have what appear to be odd edge cases to users.)
	Subcommands    []*Command
	WildSubcommand func(name string) *Command // Lets you do things like `appname subcommand1 wildcardvalue subcommand2`, where "wildcardvalue" isn't known in advance.
}

// CommandMode is sort of like a sub-subcommand:
// it's used to describe different ways of using the thing,
// but instead of being switched between by a single obvious keyword,
// it can be engaged by some combination of the other flags
// (or anything, really -- we support a callback for the recognition function).
//
// This is what we use to build really complex commands --
// think "things like `git branch`".
//
// Technically, often it would be possible to reach similar behaviors
// with a very complex combination of the `(a|b)` operations (e.g. exclusive)
// such as in a construction like `(a | b (c d))`...
// but as these get longer and more complex, they get worse to work with:
//
//   - the legibility as a usage string goes down drastically!
//   - the pain of authoring the spec goes up rapidly,
//   - and the complexity of generating a good error message becomes challenging.
//
// So, instead of leaving complex combations of exclusivity with other nested
// constructions as the only option, we offer command modes.
//
// Command modes let us produce good documentation and clear behaviors,
// because they are a check placed at the beginning of processing
// (so it's clear when it takes effect... and also, the english grammar
// in any mismatch reports doesn't have to nest with other conditions!)
// and also because they break up the problem nicely...
// and also, correspondingly, let us break up the usage synopsys text we generate.
// Win-win-win.
//
// Use of a CommandMode can produce mismatch report messages for mismatches
// within the rest of their spec that read like:
//
//   > "When using %command in %modename style (indicated by %opt1 existing), %opt2 is required"
//
// and
//
//   > "When using %command in %modename style (indicated by %opt1 existing), and %opt3 is present, %opt4 is conflicting and not allowed"
//
// In other words, the CommandMode's rules simply get described first,
// as easily-human-readable context to the rest of the pattern.
//
// Any condition can be used for deciding if a CommandMode matches or not,
// freely specifiable as a golang callback.
// The "(indicated by %opt1 existing)" clause above is also freely configurable.
// In practice, however, we advise exercising restraint with this: typically,
// checking for one or handful of opts is sufficient to get desirable behaviors,
// and going much beyond this may produce a more confusing user experience.
//
// When deciding which CommandMode applies in a Command that has several of them,
// the MatchCondition is checked for each mode is checked.
// Exactly one of them should return true.
// An error message about ambiguous CLI construction will be returned if more than one condition matches;
// as an application developer, you should aim for this to not be a reachable situation.
// As a special case, a single mode for a command may have no condition at all,
//
type CommandMode struct {
	Name           string               // Only appears in mismatch report messages, and in some parts of generated documentation.  Does not appear in the CLI nor in the synopsys example bodies.
	Spec           string               // The rest of the requirements!  Note this must still include everything that you'll read in the decision callback.
	MatchCondition func(ArgsSpool) bool // Typically, should just peek for if a certain option is existing.
	MatchCondDesc  string               // Human-readable description of MatchCondition.  Expect it to be read in a sentence prefixed by "indicated by".

	// TODO have configuration options on this which set some well-known match conditions and messages.  Some for "optionExists(a)" and "oneOf(a,b)" is probably entirely sufficient for 99% of usecases.
	// TODO consider bundling those two fields in another type.  `CommandModeRecognizer`?

	Action func(*Application, *Command, *CommandMode, ArgsSpool) error
	// TODO ^ this param list is getting silly; seems like we should name a struct for those first three or so.
	// TODO ^ this should probably get more than just the ArgsSpool, at this depth.  param specs, recognition, and possibly even assignments into any bindings should occur before we throw things over the wall.
}

type PositionalParamSpec struct {
	Name             string
	ValueTransformer func(string) (interface{}, error)

	// Note that whether something is optional isn't described here:
	//  instead, that's handled in the spec string,
	//   because it may also involve some nuanced relationship to other params.
}

type KeywordParamSpec struct {
	PrimaryLongname  string
	PrimaryShortname string
	AliasLongnames   []string
	AliasShortnames  []string
	ValueTransformer func(string) (interface{}, error)

	// Note that whether something is optional isn't described here:
	//  instead, that's handled in the spec string,
	//   because it may also involve some nuanced relationship to other params.
}
