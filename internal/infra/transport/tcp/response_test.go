package tcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_response_String(t *testing.T) {
	tests := []struct {
		name     string
		response response
		expected string
	}{
		{
			name: "Accepted",
			response: response{
				status: Accepted,
				reason: "payment accepted",
			},
			expected: "RESPONSE|ACCEPTED|Payment accepted",
		},
		{
			name: "Rejected",
			response: response{
				status: Rejected,
				reason: "payment rejected",
			},
			expected: "RESPONSE|REJECTED|Payment rejected",
		},
		{
			name:     "Empty",
			response: response{},
			expected: "RESPONSE|ACCEPTED|",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.EqualValues(t, test.expected, test.response.String())
		})
	}
}
