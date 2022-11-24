package cliharder

import (
	"strings"
)

type ArgsSpool struct {
	Positional  []string            // List of things with no dash in front of them.  Might end up being subcommands, or might end up being positional args.  Can't tell yet.
	Positional2 []string            // List of words but that followed a "--" token, which means they're definitely not subcommands.  Can also start with dashes.
	ShortOpts   map[string]int      // Short opts can't have values assigned, but can appear repeatedly.
	LongOpts    map[string][]string // Long opts can have values assigned, and can also appear repeatedly.

	// There are a few fun things we don't do at this level.
	//  For example, if you want multiple uses of a longopt to produce a list... AND you want `--theopt=foo,bar` to also produce two items...
	//   You can do that, but it doesn't happen while building the ArgsSpool.  You have to get that out at query time.
}

func ParseArgsSpool(args []string) ArgsSpool {
	return ArgsSpoolParseCfg{}.ParseArgsSpool(args)
}

type ArgsSpoolParseCfg struct {
	// Considered: a bool for whether to attempt quote stripping on long opts,
	// so that if set, the following would all be equal:
	//   ["--foo", "bar"]
	//   ["--foo=bar"]
	//   ["--foo bar"]
	//   ["--foo='bar'"]
	//   ["--foo=\"bar\""]
	// Aborted this idea for now.
	// The last three are all fairly silly ideas, because if someone writes that, it's almost certainly because they're not understanding where shells split things.
	// People do still do it.  But the amount of work it takes to support this cleanly is quite a bit,
	// and worse of all, prompts questions of if we should attempt to unescape *contents* of quoted strings,
	// and at that point I think we're just better off not trying to get fancy, because we'd be in danger of making the world _weirder_ instead of simpler.

	// So, oddly, I've got this whole config structure reserved, but nothing to actually do with it at the moment.  Ah well.  Let's call it "future proofing".
}

func (cfg ArgsSpoolParseCfg) ParseArgsSpool(args []string) ArgsSpool {
	var spool ArgsSpool
	for len(args) > 0 {
		arg := args[0]
		args = args[1:]
		switch {
		case arg == "--":
			spool.Positional2 = args
			args = nil // All remaining args are literals, then, so we're done.
		case len(arg) >= 2 && arg[0:2] == "--":
			// Longopts are a bit complex because of the key-value support,
			// Both of the following produce the same key-value pair:
			//   ["--foo", "bar"]  // Not right now, actually -- see comment a few lines below.
			//   ["--foo=bar"]
			ss := strings.SplitN(arg, "=", 2)
			switch len(ss) {
			case 1:
				// FUTURE: Peek ahead to see if the next value looks like another opt...
				//  We don't actually support this, right now, as it's written.
				//  And oddly, there's a reason.  It's because we don't know if the longopt should glom a value, or if it's a boolean/presense-only option.
				//   And we don't have that info because this parse phase doesn't know about the application's model at all.
				//    We _could_ take some more model info here.  Idk if I want to, though.  Because then we'd have to learn about subcommands down here in the parser to decide what applies, too.  And... this seems like concern-separation failure.
				//     It was kind of a vanguard wild-idea to have this phase of the parser do such strict concern-separation, but it is an idea that I do still like.
				spool.recvLongOpt(ss[0][2:], "")
			case 2:
				spool.recvLongOpt(ss[0][2:], ss[1])
			}
		case len(arg) >= 1 && arg[0] == '-':
			ss := strings.Split(arg[1:], "")
			for _, s := range ss {
				spool.recvShortOpt(s)
			}
		default:
			spool.Positional = append(spool.Positional, arg)
		}
	}
	return spool
}

func (spool *ArgsSpool) recvLongOpt(key, value string) {
	if spool.LongOpts == nil {
		spool.LongOpts = make(map[string][]string)
	}
	spool.LongOpts[key] = append(spool.LongOpts[key], value)
}

func (spool *ArgsSpool) recvShortOpt(key string) {
	if spool.ShortOpts == nil {
		spool.ShortOpts = make(map[string]int)
	}
	spool.ShortOpts[key]++
}
