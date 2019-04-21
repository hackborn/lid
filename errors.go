package lid

import (
	"errors"
)

// ------------------------------------------------------------
// ERROR

// Error struct provides additional information about an error.
type Error struct {
	Code    int
	Msg     string
	Payload interface{}
}

func (e *Error) Error() string {
	return e.Msg
}

// ------------------------------------------------------------
// UTIL

// MergeErr returns the first valid error.
// Clearly this belongs in a richer error handling package.
func MergeErr(a ...error) error {
	for _, e := range a {
		if e != nil {
			return e
		}
	}
	return nil
}

// MustErr panics on a non-nil error.
// Clearly this belongs in a richer error handling package.
func MustErr(err error) {
	if err != nil {
		panic(err)
	}
}

// ------------------------------------------------------------
// CONST and VAR

const (
	// Forbidden describes a lock that exists but is owned by another signee.
	Forbidden = iota

	forbiddenMsg = "Forbidden"
)

var (
	ErrForbidden  = &Error{Forbidden, forbiddenMsg, nil}
	ErrBadRequest = errors.New("Bad request")
)
