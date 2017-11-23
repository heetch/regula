package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		tests := []struct {
			node Node
			ctx  map[string]string
		}{
			{Eq(ValStr("foo"), ValStr("foo")), nil},
			{Eq(ValStr("foo"), VarStr("bar")), map[string]string{"bar": "foo"}},
			{In(ValStr("foo"), VarStr("bar")), map[string]string{"bar": "foo"}},
			{
				Eq(
					Eq(ValStr("bar"), ValStr("bar")),
					Eq(ValStr("foo"), ValStr("foo")),
				),
				nil,
			},
			{True(), nil},
		}

		for _, test := range tests {
			r := New(test.node, ReturnsStr("matched"))
			res, err := r.Eval(test.ctx)
			require.NoError(t, err)
			require.Equal(t, "matched", res.Value)
			require.Equal(t, "string", res.Type)
		}
	})

	t.Run("Invalid return", func(t *testing.T) {
		tests := []struct {
			node Node
			ctx  map[string]string
		}{
			{ValStr("foo"), nil},
			{VarStr("bar"), map[string]string{"bar": "foo"}},
		}

		for _, test := range tests {
			r := New(test.node, ReturnsStr("matched"))
			_, err := r.Eval(test.ctx)
			require.Error(t, err)
		}
	})
}
