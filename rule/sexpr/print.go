package sexpr

import "github.com/heetch/regula/rule"

func Print(e rule.Expr) string {
	switch v := e.(type) {
	case *rule.Value:
		switch v.Type {
		case "int64", "foat64":
			return v.Data
		}
	}
	return ""
}
