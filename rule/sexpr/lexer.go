package sexpr

import "unicode"

type Token int

const (
	EOF Token = iota
	WHITESPACE
	LPAREN // (
	RPAREN // )
	STRING
	NUMBER
	BOOL
	COMMENT
	SYMBOL
)

// isWhitespace returns true if the rune is the first rune of a
// Whitespace sequence.
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

// isLParen returns true if the rune is a left parenthesis
func isLParen(r rune) bool {
	return r == '('
}

// isRParen returns true if the rune is a right parenthesis
func isRParen(r rune) bool {
	return r == ')'
}

// isString returns true if the rune is a double quote indicating the
// beginning of string
func isString(r rune) bool {
	return r == '"'
}

// isNumber returns true if the rune is an Arabic number or the minus
// symbol, indicating the beginning of a number.
func isNumber(r rune) bool {
	// Note, although we allow a number to contain a decimal
	// point, it can't start with one so we don't include that in
	// the predicate.
	return r == '-' || (r >= '0' && r <= '9')
}

// isBool returns true if the rune is the # (hash or octothorpe)
// indicating the beginning of a boolean.
func isBool(r rune) bool {
	return r == '#'
}

// isComment returns true if the rune is ; (semicolon) indicating the
// beginning of a comment.
func isComment(r rune) bool {
	return r == ';'
}

// isSymbol returns true if no other special character type matches
// the rune.
func isSymbol(r rune) bool {
	return !(isWhitespace(r) || isLParen(r) || isRParen(r) || isString(r) || isNumber(r) || isBool(r) || isComment(r))
}
