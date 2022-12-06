package spec

// Considerably derived from the lexer package in github.com/jawher/mow.cli,
// which is MIT licensed and by Jawher Moussa.
// This version started there and has diverged;
// original credit remains where it's due,
// and further changes are Apache2-OR-MIT licensed.

import (
	"strings"

	"fmt"
)

// TokenType is a type representing the different kinds of tokens
type TokenType string

const (
	// TTArg is an arg token, e.g. <arg> <src> <dst>
	TTArg TokenType = "Arg"
	// TTOpenGroup (
	TTOpenGroup TokenType = "OpenPar"
	// TTCloseGroup )
	TTCloseGroup TokenType = "ClosePar"
	// TTOpenOptional [
	TTOpenOptional TokenType = "OpenSq"
	// TTCloseOptional ]
	TTCloseOptional TokenType = "CloseSq"
	// TTChoice |
	TTChoice TokenType = "Choice"
	// TTElipsis ...
	TTElipsis TokenType = "Elipsis"
	// TTShortOpt -a, -f, ...
	TTShortOpt TokenType = "ShortOpt"
	// TTLongOpt --force, --retry, ...
	TTLongOpt TokenType = "LongOpt"
	// TTOptSeq a folded option sequence, -rm
	TTOptSeq TokenType = "OptSeq"
	// TTOptValue is the special =<example> syntax token
	TTOptValue TokenType = "OptValue"
	// TTDoubleDash --
	TTDoubleDash TokenType = "DblDash"
)

// Token has a type, a value and a position in the input
type Token struct {
	// Type is the token type
	Typ TokenType
	// Val the textual content
	Val string
	// Pos is the token position in the input
	Pos int
}

func (t *Token) String() string {
	return fmt.Sprintf("%s('%s')@%d", t.Typ, t.Val, t.Pos)
}

// ParseError represents a parsing error
type ParseError struct {
	// Input is the text to parse
	Input string
	// Post is where in the input the error occurred
	Pos int
	// Msg s the error message
	Msg string
}

func (t *ParseError) ident() string {
	return strings.Map(func(c rune) rune {
		switch c {
		case '\t':
			return c
		default:
			return ' '
		}
	}, t.Input[:t.Pos])
}

func (t *ParseError) Error() string {
	return fmt.Sprintf("Parse error at position %d:\n%s\n%s^ %s",
		t.Pos, t.Input, t.ident(), t.Msg)
}

// Tokenize transforms the provided input into a slice of tokens or returns a ParseError
// Note this doesn't check that parens or open and close of seq match at all.  It's purely linear: no recursion and no state statck.
func Tokenize(usage string) ([]Token, error) {
	pos := 0
	var res []Token
	var (
		tk = func(t TokenType, v string) {
			res = append(res, Token{t, v, pos})
		}

		tkp = func(t TokenType, v string, p int) {
			res = append(res, Token{t, v, p})
		}

		err = func(msg string) *ParseError {
			return &ParseError{usage, pos, msg}
		}
	)
	eof := len(usage)
	for pos < eof {
		switch c := usage[pos]; c {
		case ' ':
			pos++
		case '\t':
			pos++
		case '[':
			tk(TTOpenOptional, "[")
			pos++
		case ']':
			tk(TTCloseOptional, "]")
			pos++
		case '(':
			tk(TTOpenGroup, "(")
			pos++
		case ')':
			tk(TTCloseGroup, ")")
			pos++
		case '|':
			tk(TTChoice, "|")
			pos++
		case '.':
			start := pos
			pos++
			if pos >= eof {
				return nil, err("Unexpected end of usage, was expecting rest of elipsis")
			}
			if usage[pos] != '.' {
				return nil, err("Invalid syntax: elipsis must be composed of three subsequent periods")
			}
			pos++
			if pos >= eof {
				return nil, err("Unexpected end of usage, was expecting rest of elipsis")
			}
			if usage[pos] != '.' {
				return nil, err("Invalid syntax: elipsis must be composed of three subsequent periods")
			}
			tkp(TTElipsis, "...", start)
			pos++
		case '-':
			start := pos
			pos++
			if pos >= eof {
				return nil, err("Unexpected end of usage, was expecting an option name")
			}
			switch o := usage[pos]; {
			case isLetter(o):
				pos++
				for ; pos < eof; pos++ {
					ok := isLetter(usage[pos])
					if !ok {
						break
					}
				}
				typ := TTShortOpt
				opt := usage[start:pos]
				if pos-start > 2 {
					typ = TTOptSeq
					opt = opt[1:]
				}
				tkp(typ, opt, start)
				if pos < eof && usage[pos] == '-' {
					return nil, err("Invalid syntax: cannot have dashes in middle of a short opt group")
				}
			case o == '-':
				pos++
				if pos == eof || usage[pos] == ' ' {
					tkp(TTDoubleDash, "--", start)
					continue
				}
				for pos0 := pos; pos < eof; pos++ {
					ok := isOkLongOpt(usage[pos], pos == pos0)
					if !ok {
						break
					}
				}
				opt := usage[start:pos]
				if len(opt) == 2 {
					return nil, err("Was expecting a long option name")
				}
				tkp(TTLongOpt, opt, start)
			}
		case '=':
			start := pos
			pos++
			if pos >= eof {
				return nil, err("Unexpected end of usage, was expecting '=<'")
			}
			if usage[pos] != '<' {
				return nil, err("Invalid syntax: only valid occurance of '=' is in '=<value>'")
			}
			closed := false
			for ; pos < eof; pos++ {
				closed = usage[pos] == '>'
				if closed {
					break
				}
			}
			if !closed {
				return nil, err("Unclosed option value")
			}
			if pos-start == 2 {
				return nil, err("Was expecting an option value")
			}
			pos++
			tkp(TTOptValue, usage[start:pos], start)
		case '<':
			start := pos
			pos++
			closed := false
			for ; pos < eof; pos++ {
				closed = usage[pos] == '>'
				if closed {
					break
				}
				if !isOkLongOpt(usage[pos], pos == start) {
					return nil, err("Invalid syntax: only [-_A-Za-z0-9] are permissible in arg names")
				}
			}
			if !closed {
				return nil, err("Unclosed arg declaration")
			}
			pos++
			tkp(TTArg, usage[start:pos], start)
		default:
			return nil, err("Unexpected input")
		}
	}

	return res, nil
}

func isLowercase(c uint8) bool { return c >= 'a' && c <= 'z' }
func isUppercase(c uint8) bool { return c >= 'A' && c <= 'Z' }
func isLetter(c uint8) bool    { return isLowercase(c) || isUppercase(c) }
func isDigit(c uint8) bool     { return c >= '0' && c <= '9' }
func isOkLongOpt(c uint8, first bool) bool {
	return isLetter(c) || isDigit(c) || c == '_' || (!first && c == '-')
}
