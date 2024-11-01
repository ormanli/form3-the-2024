//go:generate enumer -type=status -transform=upper

package tcp

import (
	"fmt"
	"strings"
)

// response represents a structured response containing status and reason.
type response struct {
	status status
	reason string
}

// String returns a formatted string representation of the response.
func (r response) String() string {
	return fmt.Sprintf("RESPONSE|%s|%s", r.status, capitalizeFirstLetter(r.reason))
}

// status is an enumeration type representing different possible states of a response.
type status int

const (
	Accepted status = iota
	Rejected
)

func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
