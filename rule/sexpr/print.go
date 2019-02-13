package sexpr

import (
	"strconv"

	"github.com/heetch/regula/rule"
)

func Print(e rule.Expr) (string, error) {
	switch v := e.(type) {
	case *rule.Value:
		switch v.Type {
		case "int64":
			return v.Data, nil
		case "float64":
			// For simplicity we used a fixed precision
			// representation internally, but that might
			// pack trailing zeros into its output, so we
			// go back via a real float64 for printing.
			f, err := strconv.ParseFloat(v.Data, 64)
			if err != nil {
				return "", err
			}
			return strconv.FormatFloat(f, 'f', -1, 64), nil
		}
	}
	return "", nil
}
