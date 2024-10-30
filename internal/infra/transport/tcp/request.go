package tcp

import (
	"strconv"
	"strings"

	"github.com/ormanli/form3-te/internal/app/simulator"
)

type request struct {
	amount int
}

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
