package regula_test

import (
	"context"
	"testing"
	"time"

	"github.com/heetch/regula"
	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	ctx := context.Background()

	var buf regula.RulesetBuffer

	buf.AddRuleset("match-string-a", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.Eq(regula.StringParam("foo"), regula.StringValue("bar")), regula.ReturnsString("matched a v1")),
		},
	})
	buf.AddRuleset("match-string-a", "2", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.Eq(regula.StringParam("foo"), regula.StringValue("bar")), regula.ReturnsString("matched a v2")),
		},
	})
	buf.AddRuleset("match-string-b", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("matched b")),
		},
	})
	buf.AddRuleset("type-mismatch", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "int", Data: "5"}),
		},
	})
	buf.AddRuleset("no-match", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.Eq(regula.StringValue("foo"), regula.StringValue("bar")), regula.ReturnsString("matched d")),
		},
	})
	buf.AddRuleset("match-bool", "1", &regula.Ruleset{
		Type: "bool",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "bool", Data: "true"}),
		},
	})
	buf.AddRuleset("match-int64", "1", &regula.Ruleset{
		Type: "int64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "int64", Data: "-10"}),
		},
	})
	buf.AddRuleset("match-float64", "1", &regula.Ruleset{
		Type: "float64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "float64", Data: "-3.14"}),
		},
	})
	buf.AddRuleset("match-duration", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("3s")),
		},
	})

	e := regula.NewEngine(&buf)

	t.Run("LowLevel", func(t *testing.T) {
		str, err := e.GetString(ctx, "match-string-a", regula.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, "matched a v2", str)

		str, err = e.GetString(ctx, "match-string-a", regula.Params{
			"foo": "bar",
		}, regula.Version("1"))
		require.NoError(t, err)
		require.Equal(t, "matched a v1", str)

		str, err = e.GetString(ctx, "match-string-b", nil)
		require.NoError(t, err)
		require.Equal(t, "matched b", str)

		b, err := e.GetBool(ctx, "match-bool", nil)
		require.NoError(t, err)
		require.True(t, b)

		i, err := e.GetInt64(ctx, "match-int64", nil)
		require.NoError(t, err)
		require.Equal(t, int64(-10), i)

		f, err := e.GetFloat64(ctx, "match-float64", nil)
		require.NoError(t, err)
		require.Equal(t, -3.14, f)

		_, err = e.GetString(ctx, "match-bool", nil)
		require.Equal(t, regula.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "type-mismatch", nil)
		require.Equal(t, regula.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "no-match", nil)
		require.Equal(t, regula.ErrNoMatch, err)

		_, err = e.GetString(ctx, "not-found", nil)
		require.Equal(t, regula.ErrRulesetNotFound, err)
	})

	t.Run("StructLoading", func(t *testing.T) {
		to := struct {
			StringA  string        `ruleset:"match-string-a"`
			Bool     bool          `ruleset:"match-bool"`
			Int64    int64         `ruleset:"match-int64"`
			Float64  float64       `ruleset:"match-float64"`
			Duration time.Duration `ruleset:"match-duration"`
		}{}

		err := e.LoadStruct(ctx, &to, regula.Params{
			"foo": "bar",
		})

		require.NoError(t, err)
		require.Equal(t, "matched a v2", to.StringA)
		require.Equal(t, true, to.Bool)
		require.Equal(t, int64(-10), to.Int64)
		require.Equal(t, -3.14, to.Float64)
		require.Equal(t, 3*time.Second, to.Duration)
	})

	t.Run("StructLoadingWrongKey", func(t *testing.T) {
		to := struct {
			StringA string `ruleset:"match-string-a,required"`
			Wrong   string `ruleset:"no-exists,required"`
		}{}

		err := e.LoadStruct(ctx, &to, regula.Params{
			"foo": "bar",
		})

		require.Error(t, err)
	})

	t.Run("StructLoadingMissingParam", func(t *testing.T) {
		to := struct {
			StringA string `ruleset:"match-string-a"`
		}{}

		err := e.LoadStruct(ctx, &to, nil)

		require.Error(t, err)
	})
}
