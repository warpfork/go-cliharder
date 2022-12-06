go-cliharder
============

It's a CLI library.



Another CLI lib?!
-----------------

Yeah.  I know.

I'm honestly not overjoyed about writing this either.
Can't say I want Another Thing to maintain.

Nonetheless, here's some features that I couldn't find anywhere else, or, couldn't find all together in one place:

- The basics: `--longopts`, short `-asdf -a -s -d -f` opts, `--value=opts`, etc.
- Controllable exit codes and output routing.
- GOOD ERROR MESSAGES.
- Good auto-generated usage docs.
- Position-independent flags.  (E.g., most of the time, `foo bar -v` and `foo -v bar` mean the same thing.)
- Subcommands.
- Ergonomic extraction of the parsed values.  And more than one option for this.

And considered in some other libraries:

- Controllable exit codes and output routing:
	- I introduced this myself in `jahwer/mow.cli`!  See github.com/jawher/mow.cli/pulls/128 .
	- This _kinda_ exists in `urfave/cli`... but controlling the exit code is callback-based and a bit weird and honestly has been a source of bugs in about half the applications I've seen developers try to do it in.
	- Of all the complaints I have with other libraries, this is the one that's at least _sometimes_ satiated... but it's pretty hit-or-miss, and usually overcomplicated (e.g. needing callbacks?  why??).
- GOOD ERROR MESSAGES.
	- `mow.cli` really struggles with this.  See github.com/jawher/mow.cli/issues/84 (among others).  The flexibility of the library's parser means producing a good error message isn't just a missing feature, it's nearly impossible by design.  (It's getting its extremely flexibility because it's trying to pattern-match by more or less brute force, including liberal backtracking... which means if it has to explain what it tried, it would have a _lot_ of explaining to do.  Too much for a human to read, and find any terse point in!)
- Good auto-generated usage.
	- `urfave/cli` really tried hard with this.  But I don't think the outcome satisfies me.
		- The API doesn't guide you to write good usage.
		- You still end up having to write most of the most-important parts (the tl;dr usage and the essential flag examples at the top of the usage) yourself!
		- The spacing is just honestly weird.
		- I think most applications that care about good help text are going to write custom templates anyway.  And at that point.. what's the point?
	- `jahwer/mow.cli` is outstandingly good at this.  I hope we can match it!  Parsing the example usage as a spec is a brilliant idea that makes sure it's load-bearing; really good.
- Position-independent flags.  Most of the time, `foo bar -v` and `foo -v bar` mean the same thing.
	- `urfave/cli` doesn't do this at all.  Bugs the heck out of me.
		- Workaround: write a construction assistance function that adds the same flag to every leafmost subcommand.  Works; but doesn't feel great.  And definitely doesn't produce autodocumentation that looks right.
	- `jahwer/mow.cli` apparently doesn't support this either; see github.com/jawher/mow.cli/issues/64 .
- Subcommands have to be good.
	- Lots of libraries have done this, actually.  But not all of them.  So it does evict some libraries from consideration.
- I need more than one way to handle value extraction.
	- Getters like `parsedArgs.GetLongOpt("foo")` are a good option, but shouldn't be the only one -- this approach is flexible, but relies on magic constant alignment, so it's a bit fragile to maintain.
	- Systems where you bind a value at the time of constructing the CLI model are quite good.  We'll do that in go-cliharder, for sure.
	- Systems where you have to write a struct and use tags to line things up... can look nice, and are a neat option to have.  But absolutely shouldn't be the _only_ option, because it means programmatically generating a model is impossible, and that's silly.



Status
------

Not even early alpha.  More of a dream.  A shamble of code that I've pushed.

If you want to run with it, please do.  It's "free" as in "free puppy".



How does go-cliharder get good error messages?
----------------------------------------------

tl;dr: We focused it from the start as an overridingly high priority, and the cornerstone of the design.

Any time there's a _choice_ in how the parser will proceed, and that choice will affect several subsequent options, _you have to express it that way_ in your setup of the CLI model, too.

So: for example, if you want to parse something like:

```
USAGE:

	frobnicator froz [SRC...] [DST]
	frobnicator froz --zot [--frong] [--dangle]
```

... you need to set up a chooser, like this:

```go
	var frozSubcmd cliharder.Subcommand = /*...*/
	frozSubcmd.MultipleChoice(
		func(args cliharder.ArgSpool) bool { return args.Contains("--zot") },
		cliharder.Subcommand(/*...*/),
	)
	frozSubcmd.MultipleChoice(
		nil, // nil for the matcher predicate means this is the default / last resort.
		cliharder.Subcommand(/*...*/),
	)
```

The first arg is a matcher callback.  You can do whatever you want in these.
They'll be evaluated in order.  The first one to return `true` wins.
The argument unpacking process will only follow that path further down...
and that means if we have any errors there, we can report _exactly that_ error
(rather than needing to backtrack and try all other possible matches,
which would then stick us with a _much_ harder problem to tersely explain what we tried).

(I have to give thanks for `mow.cli` for this idea; it came to me while looking at
that library's FSM and its error flow path, and thinking "wow, how can I _not_ end up in this situation?".
Adding some printf debugs to that thing and seeing it try hundreds of combinations
for even fairly simple CLIs was eye-opening, and steered me hard towards
designing a model that's explicit enough that the model's evaluator can explain itself later.)

As a bonus, we then use this info during help message generation, too.
You see this sort of thing in complex commands in the wild all the time.  For example, `git branch`:

```
SYNOPSIS
       git branch [--color[=<when>] | --no-color] [-r | -a]
               [--list] [-v [--abbrev=<length> | --no-abbrev]]
               [--column[=<options>] | --no-column] [--sort=<key>]
               [(--merged | --no-merged) [<commit>]]
               [--contains [<commit]] [--no-contains [<commit>]]
               [--points-at <object>] [--format=<format>] [<pattern>...]
       git branch [--track | --no-track] [-l] [-f] <branchname> [<start-point>]
       git branch (--set-upstream-to=<upstream> | -u <upstream>) [<branchname>]
       git branch --unset-upstream [<branchname>]
       git branch (-m | -M) [<oldbranch>] <newbranch>
       git branch (-c | -C) [<oldbranch>] <newbranch>
       git branch (-d | -D) [-r] <branchname>...
       git branch --edit-description [<branchname>]
```

go-cliharder will produce a synopsis sorta like that, with multiple distinct entries,
when you use the MultipleChoice feature.



Other Insights into The Guts
----------------------------

I like knowing how libraries work from the bottom up, so here's that info for ya.

Parsing goes roughly in this order:

- we split up all the argument strings and decide if each token looks like a subcommand, a longopt, a shortopt, or a dangling argument.
	- this can be a little vague sometimes -- subcommands don't necessarily _stop_ being acceptable at the first opt.  But we'll finish handling this before making it a user-bother.
- we put those in a disordered heap of data, but prepare it to be queryable.  We call this the "ArgsSpool".
	- e.g. you can ask this heap if it's got a shortopt called "-a" (and that'll return true even if the actual args were "-xab").
	- stuff like `--x=y` will get parsed apart as a key-value assignment.
	- stuff like `-x y` (assuming those are actually separate strings in `os.Args`, that is) will get parsed as key-value assignment too.
	- no other checking is going on at this stage, yet.
- we'll look at your CLI model and start matching subcommands.
- when we find the command of note: we'll look at its argument expectations and start checking that we've got them.
	- you can have some control of this with callbacks if you need extremely custom logic!  See the `Command.MultipleChoice` API, for example.

We support all sorts of weird (and not-so-weird) stuff with this:

- You want piled-up short opts, like "-aux", to be the same as "-a -u -x"?  Yep.  So it is.
- You want long opts like `--foo=bar`?  Fine.
- You want multiple short opts, like "-vvv", to be distinct from "-v"?  You got it.  (We'll count up the occurrences and let you know.)
- You want multiple long opts to let you accumulate a list, so `--foo=bar --foo=baz` produces `["bar", "baz"]`?  Sure thing.  Supported.
- You want `foo -- --bar` to parse as two positional arguments, not a long opt?  Yep.  We did that.
- If you want to have an argument actually in the _middle_ of subcommand tree...
	- Okay, "wildcard subcommands" are not supported _yet_, but we probably will.

There's a very short list of things we don't support:

- We don't support key-value with short args.  e.g. `-ax=foo` is going to get parsed as the short opts ["a", "x", "=", "f", "o", "o"].
	- It's true some applications have flags like this, but it's uncommon, and I don't think it's a good idea to write new applications with this behavior.
- You can't differentiate `foo --bar --baz` from `foo --baz --bar`.  We don't preserve order.
	- I don't think you should ever want to do this.  It would surprise almost any user.
- You can't differentiate `foo -v bar` from `foo bar -v`.  Again, we *really* don't preserve order.
	- I'd still argue that you simply shouldn't want to do this.  It will surprise most users.
- If you're really enthusiastic about any of these things, and really want to do them, but you also want what go-cliharder does most of the time, then the best thing I recommend you is to just look at the raw args array again a second time after we've done the first round of processing for you.  That works fine; and then anything is possible.



License
-------

This is open-source stuff.  Your choice of Apache-2 or MIT.  Literally:

SPDX-License-Identifier: Apache-2.0 OR MIT

There are also some snippets of code which derive from pure MIT sources (not Apache2),
so that limitation is probably overriding.
