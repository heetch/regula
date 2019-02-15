package sexpr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/heetch/regula/rule"
)

var sm *opCodeMap

func init() {
	sm = makeSymbolMap()
}

// prettyPrintOp is the generic printer for operations
func prettyPrintOp(indent, wrap int, e rule.Expr) (string, error) {
	output := ""
	v, ok := e.(rule.Operator)
	if !ok {
		return "", fmt.Errorf("called prettyPrintOp with non-operator expression")
	}
	symbol, err := sm.getSymbolForOpCode(v.Contract().OpCode)
	if err != nil {
		return "", err
	}
	prefix := "(" + symbol
	subdent := indent + len(prefix)
	currentIndent := subdent
	leader := strings.Repeat(" ", subdent)
	output += prefix
	ops := v.Operands()
	for _, op := range ops {
		unindented, err := PrettyPrint(0, wrap, op)
		if err != nil {
			return "", err
		}
		parts := strings.Split(unindented, "\n")
		if (currentIndent + 1 + len(parts[0])) >= wrap {
			output += "\n" + leader
			currentIndent = subdent
		}
		indented, err := PrettyPrint(currentIndent+1, wrap, op)
		if err != nil {
			return "", err
		}
		parts = strings.Split(indented, "\n")
		if len(parts) > 1 {
			output += " " + indented
			currentIndent = len(parts[len(parts)-1])
		} else {
			output += " " + indented
			currentIndent = currentIndent + 1 + len(parts[0])
		}
	}
	output += ")"
	return output, nil
}

// prettyPrintLet is a special case printer for let forms.  The intent is to force lets to conform to the form:
//
//    (let symbol value
//         (body ...))
//
func prettyPrintLet(indent, wrap int, e rule.Expr) (string, error) {
	v, ok := e.(rule.Operator)
	if !ok || v.Contract().OpCode != "let" {
		return "", fmt.Errorf("called prettyPrintLet without a let expression")
	}
	ops := v.Operands()
	subdent := indent + 5 // 5 = len("(let ")
	subleader := strings.Repeat(" ", subdent)
	symbol, err := PrettyPrint(subdent, wrap, ops[0])
	if err != nil {
		return "", err
	}
	value, err := PrettyPrint(subdent+1+len(symbol), wrap, ops[1])
	if err != nil {
		return "", err
	}
	body, err := PrettyPrint(subdent, wrap, ops[2])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(let %s %s\n%s%s)", symbol, value, subleader, body), nil
}

// prettyPrintIf is the special case printer for if forms.  The intent is that if forms should always print with the following outline:
//
//    (if (test)
//        (true-body)
//        (false-body))
//
func prettyPrintIf(indent, wrap int, e rule.Expr) (string, error) {
	v, ok := e.(rule.Operator)
	if !ok || v.Contract().OpCode != "if" {
		return "", fmt.Errorf("called prettyPrintIf without and if expression")
	}
	ops := v.Operands()
	subdent := indent + 4 // 4 = len("(if ")
	subleader := strings.Repeat(" ", subdent)
	test, err := PrettyPrint(subdent, wrap, ops[0])
	if err != nil {
		return "", err
	}
	truePart, err := PrettyPrint(subdent, wrap, ops[1])
	if err != nil {
		return "", err
	}
	falsePart, err := PrettyPrint(subdent, wrap, ops[2])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(if %s\n%s%s\n%s%s)", test, subleader, truePart, subleader, falsePart), nil
}

// PrettyPrint takes a tree of expressions and prints them out as a well-formatted RUSE program.
func PrettyPrint(indent, wrap int, e rule.Expr) (string, error) {
	var line string
	var err error
	switch v := e.(type) {
	case *rule.Value:
		switch v.Type {
		case "int64":
			line = v.Data
		case "float64":
			// For simplicity we used a fixed precision
			// representation internally, but that might
			// pack trailing zeros into its output, so we
			// go back via a real float64 for printing.
			f, err := strconv.ParseFloat(v.Data, 64)
			if err != nil {
				return "", err
			}
			line = strconv.FormatFloat(f, 'f', -1, 64)
		case "bool":
			line = "#" + v.Data
		case "string":
			line = fmt.Sprintf("%q", v.Data)
		}
	case *rule.Param:
		line = v.Name
	case rule.Operator:
		switch v.Contract().OpCode {
		case "if":
			line, err = prettyPrintIf(indent, wrap, e)
			if err != nil {
				return "", err
			}
		case "let":
			line, err = prettyPrintLet(indent, wrap, e)
			if err != nil {
				return "", err
			}
		default:
			line, err = prettyPrintOp(indent, wrap, e)
			if err != nil {
				return "", err
			}
		}
	}
	return line, nil
}
