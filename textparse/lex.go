package textparse

import (
	"unicode"
	"unicode/utf8"
)

const maxInputBytes = 4096

type token struct {
	text       string
	start, end int
}

type cursor struct {
	input  string
	tokens []token
	pos    int
	endPos int
}

func scan(input string, stateComments bool) (*cursor, error) {
	if len(input) > maxInputBytes {
		return nil, diagnostic(input, token{start: 0, end: len(input)}, "invalid_input", "input", "input exceeds 4096 bytes", "Use one bounded, single-line command.", nil)
	}
	for i := 0; i < len(input); {
		r, n := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && n == 1 {
			return nil, diagnostic(input, token{start: i, end: i + 1}, "invalid_input", "input", "input must be valid UTF-8", "Provide valid UTF-8 text.", nil)
		}
		if r == '\r' || r == '\n' || r == '\v' || r == '\f' || r == '\u0085' || r == '\u2028' || r == '\u2029' {
			return nil, diagnostic(input, token{start: i, end: i + n}, "invalid_input", "input", "input must be a single line", "Parse each command line separately.", nil)
		}
		i += n
	}
	c := &cursor{input: input, endPos: len(input)}
	// Each token contains at least one byte, so the byte limit also bounds the
	// number of tokens and all scanning work. No recursive grammar is used.
	for i := 0; i < len(input); {
		if stateComments && i+1 < len(input) && input[i:i+2] == "//" {
			previous, _ := utf8.DecodeLastRuneInString(input[:i])
			if i == 0 || unicode.IsSpace(previous) {
				c.endPos = i
				break
			}
		}
		r, n := utf8.DecodeRuneInString(input[i:])
		if unicode.IsSpace(r) {
			i += n
			continue
		}
		start := i
		if punctuation(r) {
			i += n
		} else {
			for i < len(input) {
				r, n = utf8.DecodeRuneInString(input[i:])
				if unicode.IsSpace(r) || punctuation(r) {
					break
				}
				i += n
			}
		}
		c.tokens = append(c.tokens, token{text: input[start:i], start: start, end: i})
	}
	return c, nil
}

func punctuation(r rune) bool { return r == '(' || r == ')' || r == ',' || r == '=' }

func lowerASCII(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			if b == nil {
				b = []byte(s)
			}
			b[i] += 'a' - 'A'
		}
	}
	if b == nil {
		return s
	}
	return string(b)
}

func (c *cursor) end() bool { return c.pos == len(c.tokens) }

func (c *cursor) peek() token {
	if c.end() {
		return token{start: c.endPos, end: c.endPos}
	}
	return c.tokens[c.pos]
}

func (c *cursor) next() token {
	t := c.peek()
	if !c.end() {
		c.pos++
	}
	return t
}

func (c *cursor) take(text string) bool {
	if lowerASCII(c.peek().text) == text {
		c.pos++
		return true
	}
	return false
}

func (c *cursor) err(t token, code, field, message, hint string, cause error) error {
	return diagnostic(c.input, t, code, field, message, hint, cause)
}

func (c *cursor) expect(text, field string) error {
	if c.take(text) {
		return nil
	}
	code := "unexpected_input"
	if c.end() {
		code = "missing_field"
	}
	return c.err(c.peek(), code, field, "expected "+text, "Use the documented command syntax.", nil)
}
