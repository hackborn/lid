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

// mergeErr returns the first valid error.
func mergeErr(a ...error) error {
	for _, e := range a {
		if e != nil {
			return e
		}
	}
	return nil
}

// mustErr panics on a non-nil error.
func mustErr(err error) {
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
	errForbidden            = errors.New(forbiddenMsg)
	errBadRequest           = errors.New("Bad request")
	errConditionFailed      = errors.New("Condition failed")
	errDurationRequired     = errors.New("Bad request: Duration required")
	errDynamoRequired       = errors.New("Can't create DynamoDB")
	errInitializationFailed = errors.New("Initialization failed")
	errSessionRequired      = errors.New("Session is required")
	errTableRequired        = errors.New("Bad request: Table name required")
)
