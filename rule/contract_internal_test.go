package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A contract may only have Cardinality == MANY on its final Term.
func TestContractCanOnlyHaveCardinalityManyOnLastTerm(t *testing.T) {
	c := Contract{
		OpCode:     "whatever",
		ReturnType: INTEGER,
		Terms: []Term{
			{
				Type:        INTEGER,
				Cardinality: MANY,
				Min:         2,
			},
			{
				Type:        BOOLEAN,
				Cardinality: ONE,
			},
		},
	}

	require.PanicsWithValue(t, `bad contract for "whatever", only the last Term in a Contract may have Cardinality == MANY`, func() { c.GetTerm(0) })
}
