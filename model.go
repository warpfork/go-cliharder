package cliharder

import (
	"context"
	"io"
)

// Code in this file is the core structures for describing commands and their options.

// TODO: Everything below is a sketch, so far.

type Application struct {
	Context context.Context // Just carried through, for your use if necessary.
	Stdout  io.Writer
	Stderr  io.Writer
	Command *Command
}

type Command struct {
	Name    string
	AliasOf []string // If present, no other fields are valid -- this is a redirection.  The AliasOf list starts at the root of the application (it can alias even to non-sibling subcommands).

	RecognitionPhaseAllowsUnrecognizedOpts bool
	AdditionalValidation                   func(ArgsSpool) []string // Runs after the recognition phase but before the action.  Any strings return become part of an error report.  There's no reason you couldn't do this work in the action, too, but sometimes the separation increases legibility.

	Action         func(Application, ArgsSpool) error // If unassigned, and subcommands are present, a fallback which emits a summary of subcommands will be used.
	Subcommands    []*Command
	WildSubcommand func(name string) *Command // Lets you do things like `appname subcommand1 wildcardvalue subcommand2`, where "wildcardvalue" isn't known in advance.
}

type Opt struct {
	PrimaryLongname  string
	PrimaryShortname string
	AliasLongnames   []string
	AliasShortnames  []string
	ValueTransformer func(string) (interface{}, error)
}
