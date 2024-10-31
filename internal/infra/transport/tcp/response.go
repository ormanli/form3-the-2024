//go:generate enumer -type=status -transform=upper

package tcp

import (
	"fmt"
	"strings"
)

type response struct {
	status status
	reason string
}

func (r response) String() string {
	return fmt.Sprintf("RESPONSE|%s|%s", r.status, capitalizeFirst(r.reason))
}

type status int

const (
	Accepted status = iota
	Rejected
)

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
