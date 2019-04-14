package dlock

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

// mustErr() is a simple utility to panic on errors.
func mustErr(err error) {
	if err != nil {
		panic(err)
	}
}

// ------------------------------------------------------------
// CONST and VAR

const (
	// AlreadyLocked is the code for the already locked error.
	AlreadyLocked = iota

	alreadyLockedMsg = "Already locked"
)

var (
	errAlreadyLocked        = errors.New(alreadyLockedMsg)
	errBadRequest           = errors.New("Bad request")
	errConditionFailed      = errors.New("Condition failed")
	errDurationRequired     = errors.New("Bad request: Duration required")
	errDynamoRequired       = errors.New("Can't create DynamoDB")
	errInitializationFailed = errors.New("Initialization failed")
	errSessionRequired      = errors.New("Session is required")
	errTableRequired        = errors.New("Bad request: Table name required")
)
