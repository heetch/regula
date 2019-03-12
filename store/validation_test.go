package store

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestValidation(t *testing.T) {
	t.Run("OK - ruleset name", func(t *testing.T) {
		names := []string{
			"path/to/my-ruleset",
			"path/to/my-awesome-ruleset",
			"path/to/my-123-ruleset",
			"path/to/my-ruleset-123",
		}

		for _, n := range names {
			err := ValidatePath(n)
			require.NoError(t, err)
		}
	})

	t.Run("NOK - ruleset name", func(t *testing.T) {
		names := []string{
			"",
			"PATH/TO/MY-RULESET",
			"/path/to/my-ruleset",
			"path/to/my-ruleset/",
			"path/to//my-ruleset",
			"path/to/my_ruleset",
			"1path/to/my-ruleset",
			"path/to/my--ruleset",
		}

		for _, n := range names {
			err := ValidatePath(n)
			require.True(t, IsValidationError(err))
		}
	})

	t.Run("OK - param names", func(t *testing.T) {
		names := []string{
			"a",
			"abc",
			"abc-xyz",
			"abc-123",
			"abc-123-xyz",
		}

		for _, n := range names {
			rs := regula.NewRuleset(
				rule.New(
					rule.BoolParam(n),
					rule.BoolValue(true),
				),
			)

			for _, r := range rs.Rules {
				for _, param := range r.Params() {
					err := ValidateParamName(param.Name)
					require.NoError(t, err)
				}
			}
		}
	})

	t.Run("NOK - param names", func(t *testing.T) {
		names := []string{
			"ABC",
			"abc-",
			"abc_",
			"abc--xyz",
			"abc_xyz",
			"0abc",
		}

		names = append(names, reservedWords...)

		for _, n := range names {
			rs := regula.NewRuleset(
				rule.New(
					rule.BoolParam(n),
					rule.BoolValue(true),
				),
			)

			for _, r := range rs.Rules {
				for _, param := range r.Params() {
					err := ValidateParamName(param.Name)
					require.True(t, IsValidationError(err))
				}
			}
		}
	})
}

func TestValidateRule(t *testing.T) {
	sig := regula.Signature{ReturnType: "bool", Params: map[string]string{"foo": "int64"}}

	tests := []struct {
		name  string
		rule  *rule.Rule
		fails bool
	}{
		{"empty rule", rule.New(nil, nil), true},
		{"no expr", rule.New(nil, rule.BoolValue(true)), true},
		{"no value", rule.New(rule.True(), nil), true},
		{"wrong return type", rule.New(rule.True(), rule.Int64Value(10)), true},
		{"wrong param type", rule.New(rule.Float64Param("foo"), rule.BoolValue(true)), true},
		{"unknown param", rule.New(rule.Int64Param("bar"), rule.BoolValue(true)), true},
		{"no params", rule.New(rule.True(), rule.BoolValue(true)), false},
		{"right param", rule.New(rule.Int64Param("foo"), rule.BoolValue(true)), false},
		{"bad return value", rule.New(rule.True(), &rule.Value{Type: "bool", Kind: "value", Data: "100"}), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.fails, ValidateRule(&sig, test.rule) != nil)
		})
	}
}
