package tcp

import (
	"strconv"
	"strings"

	"github.com/ormanli/form3-te/internal/app/simulator"
)

// request represents a payment request with an associated amount.
type request struct {
	amount int
}

// parseRequest parses a string representation of a payment and returns a request object along with any error encountered during parsing.
func parseRequest(s string) (request, error) {
	parts := strings.Split(s, "|")
	if len(parts) != 2 || parts[0] != "PAYMENT" {
		return request{}, simulator.ErrInvalidRequest
	}

	amount, err := strconv.Atoi(parts[1])
	if err != nil {
		return request{}, simulator.ErrInvalidAmount
	}

	return request{amount: amount}, nil
}
