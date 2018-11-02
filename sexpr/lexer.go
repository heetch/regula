package sexpr

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

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isLParen(r rune) bool {
	return r == '('
}

func isRParen(r rune) bool {
	return r == ')'
}

func isString(r rune) bool {
	return r == '"'
}

func isNumber(r rune) bool {
	// Note, although we allow a number to contain a decimal
	// point, it can't start with one so we don't include that in
	// the predicate.
	return r == '-' || (r >= '0' && r <= '9')
}

func isBool(r rune) bool {
	return r == '#'
}

func isComment(r rune) bool {
	return r == ';'
}
