package tcp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ormanli/form3-te/internal/app/simulator"
)

func Test_parseRequest(t *testing.T) {
	tests := []struct {
		input      string
		assertFunc func(*testing.T, request, error)
	}{
		{
			input: "PAYMENT|1",
			assertFunc: func(t *testing.T, r request, err error) {
				assert.NoError(t, err)
				assert.EqualValues(t, request{amount: 1}, r)
			},
		},
		{
			input: "PAYMENT",
			assertFunc: func(t *testing.T, r request, err error) {
				assert.ErrorIs(t, err, simulator.ErrInvalidRequest)
				assert.Empty(t, r)
			},
		},
		{
			input: "PAYMENT|1|2",
			assertFunc: func(t *testing.T, r request, err error) {
				assert.ErrorIs(t, err, simulator.ErrInvalidRequest)
				assert.Empty(t, r)
			},
		},
		{
			input: "PAYMENT|A",
			assertFunc: func(t *testing.T, r request, err error) {
				assert.ErrorIs(t, err, simulator.ErrInvalidAmount)
				assert.Empty(t, r)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			r, err := parseRequest(test.input)
			test.assertFunc(t, r, err)
		})
	}
}
