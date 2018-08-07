package etcd

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
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
			err := validateRulesetName(n)
			require.NoError(t, err)
		}
	})

	t.Run("NOK - ruleset name", func(t *testing.T) {
		names := []string{
			"PATH/TO/MY-RULESET",
			"/path/to/my-ruleset",
			"path/to/my-ruleset/",
			"path/to//my-ruleset",
			"path/to/my_ruleset",
			"1path/to/my-ruleset",
			"path/to/my--ruleset",
		}

		for _, n := range names {
			err := validateRulesetName(n)
			require.Equal(t, store.ErrBadRulesetName, err)
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
			rs, _ := regula.NewBoolRuleset(
				rule.New(
					rule.BoolParam(n),
					rule.BoolValue(true),
				),
			)

			err := validateParamNames(rs)
			require.NoError(t, err)
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

		for _, n := range names {
			rs, _ := regula.NewBoolRuleset(
				rule.New(
					rule.BoolParam(n),
					rule.BoolValue(true),
				),
			)

			err := validateParamNames(rs)
			require.Equal(t, store.ErrBadParameterName, err)
		}

		// For the following tests, we are just testing if the recursion and the type assertion work well.
		// The validation is already tested on the rule package.
		rs, _ := regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.True(),
					rule.BoolParam("foo_"),
				),
				rule.BoolValue(true),
			),
		)

		err := validateParamNames(rs)
		require.Equal(t, store.ErrBadParameterName, err)

		rs, _ = regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.True(),
					rule.New(
						rule.Eq(
							rule.BoolParam("foo"),
							rule.BoolParam("1baz"), // bad param name
						),
						rule.BoolValue(true),
					),
					rule.BoolParam("foo"),
				),
				rule.BoolValue(true),
			),
		)

		err = validateParamNames(rs)
		require.Equal(t, store.ErrBadParameterName, err)

		rs, _ = regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.True(),
					rule.New(
						rule.Eq(
							rule.New(
								rule.Eq(
									rule.BoolParam("foo"),
									rule.BoolParam("foo--bar"), // bad param name
									rule.True(),
								),
								rule.BoolValue(true),
							),
							rule.True(),
						),
						rule.BoolValue(true),
					),
				),
				rule.BoolValue(true),
			),
		)

		err = validateParamNames(rs)
		require.Equal(t, store.ErrBadParameterName, err)
	})
}
