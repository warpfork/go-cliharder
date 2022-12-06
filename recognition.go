package cliharder

// Code in this file is about mapping data from ArgsSpool into value structs constrained by the model (`Command`, `PositionalParamValue`, etc).

// TODO entirely :)

/*

### Decision Flow Overview

The decision flow looks something roughly like this:

- First we walk down subcommand trees.
	- These are very literal and straightforward.  The structure is explicit in its construction in golang calls.
	- This may consume one or more pieces from `ArgsSpool.Positional`.  We'll pass on a new ArgsSpool that's got the consumed parts sliced out.
- If this command has MultipleChoice, that table is iterated over, and the decider callbacks invoked until a match is found.
	- Each of these callbacks gets a view of the ArgsSpool (and ought not mutate it).
	- REVISIT.  This can't really be totally generally callbacks without compromising our error reporting.  Interface that also demands error description is the farthest possible.  But we should probably have some well-known ones.  There aren't that many sane options.
	- The only purpose of this vs all the `(|)` syntaxes is so we can generate broken-apart usage synopsis, and because in the API it's often nice to have separate action funcs for this.
		- It also lets us pick one thing that's the decider, regardless of order, which is kind of a big deal.  Otherwise it's not clear *which* thing is the decider, and error messages get way harder.
		- Look at the synopsis of `man git branch` for a practical example of why this is important.  Could all that have been done with massive parens?  Maybe.  Would it have been sane?  Deffo not.
- Longopts and shortopts all get mapped into their relevant structures.
	- TBD: longopt processing might be in a reasonable position to glomp a positional here?  Unclear.
		- We don't keep enough info for that right now, but we _could_ keep an index from spool parse time that says what positional came after it, thus letting us perform this decision down here.
- Any present but not-understood longopts and short opts: there's a callback for whether we want to accumulate errors for these or ignore them.
- Any remaining positional args are mapped.
	- If there are optional positionals, they are consumed greedily.
	- If there's any varargs... there's necessarily only one.  (We don't support custom literals that might be used as separators.)
	- If you're reading between the lines here: yes, we're simply not going to support many of the possible combinations of optional positionals.  Supporting everything under the sun would make parsing hard, and it would be a bad UX to your users anyway, so let's just not go there.


### Supporting Option Relationships

Now.  The Decision Flow Overview glosses over a few things.

The spec syntax supports, and we also want to validate, some moderately complex things:
for example, mutually exclusive options.
In complex, nested ways.

- Consider: `[-R [-H | -L | -P]]`.
- Consider: `[-d|--rm] IMAGE [COMMAND [ARG...]]`.

There's a whole satisfiability thing going on regarding whether each option was present.
And for positionals, because the it's not order independent, things get tricky indeed
(jump down to the next section of notes for more than that).

Error messages for for this stuff must focus on being straightforward:

- "-H" for the spec `[-R [-H | -L | -P]]` should say: "-H only valid when -R is present".
- "-ab" for the spec `(-a | -b | -c | -d)` should say: "-a cannot coexist with -b".
  - It might also say the converse -- or both.  (I don't really care if it's redundant; if it's simpler to code that way, sure, let's have it.)
- "-Z" for the spec `[-X [-Y [-Z]]` should say "-Z only valid when -Y is present".
  - Is that the whole story?  No.  Is it enough?  Yes.  **Going one level is sufficient.**
    - "-YZ" for the spec `[-X [-Y [-Z]]` will then tell you "-Y is only valid when -X is present"... and I think a user discovering that iteratively, if they didn't read the docs, is a perfectly acceptable journey -- and in fact a good one, because it keeps the error text size minimal at each step of the journey.  Smaller, faster feedback is better.

I believe the best route to this is to preprocess the satisfaction conditions for each option.
Then, we can do all the above checks by just looping over each option and checking its conditions.
Each satisfaction condition can produce one clear error message.
(Notice how in each of the examples above, the error message mentions exactly two params?  I think that's good / not a coincidence.  Pairwise validation is very clear.  Clear is good.)


### Positionals can be Optional... in Limited Circumstances.

Support for optional positionals is (intentionally) limited.

The spec syntax tokenizer is pretty silent on this at the moment, but I think most complex structures are silly.
The `X [Y [Z...]]` right-leaning structure is approximately the only one that makes any dang sense.

- `X... Y` admittedly has been seen in the wild (this is `cp` for example).  Maybe it's worth also giving this first class support.
- `X [Y] Z` is technically possible but rather strange to do.
- `X [Y...] Z` is also technically possible but odd to do.  I don't think anyone's *ever* done this in the wild.
- `X [[Y...] Z]` remains technically parsible but is really getting deranged.
- `X [[[Y...] Z] W]` is as parsible as above, but the derangement level is rising and why would you do this to users?
- `X [[W [Y...]] Z]` is... still, _technically_ parsable, at great effort, but _why would you do this to users_?

tl;dr whether something is technically parsable doesn't correspond to whether
it's a good idea to support it as a library.
Some combinations are massively more work to parse correctly;
will get harder to produce reasonably explanatory errors for;
and will be highly likely to produce end-user confusion if ever actually used anyway;
and all that sums to "bad idea to use; even worse idea to spend time supporting as a library".

Restricting our support to varags at the beginning or the end,
and generally being greedy in matching any other optional positionals,
seems like it should cover 99% of usecases.
Anything beyond that can be handled by the library user saying `<therest>...` and handling it themselves.

*/
