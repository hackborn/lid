package dlock

import (
	"errors"
)

// ------------------------------------------------------------
// PACKAGE-ERROR

// PackageError struct provides additional information about an error.
type PackageError struct {
	Msg     string
	Payload interface{}
}

func (e *PackageError) Error() string {
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
	alreadyLockedMsg = "Already locked"
)

var (
	alreadyLockedErr        = errors.New(alreadyLockedMsg)
	badRequestErr           = errors.New("Bad request")
	conditionFailedErr      = errors.New("Condition failed")
	durationRequiredErr     = errors.New("Bad request: Duration required")
	dynamoRequiredErr       = errors.New("Can't create DynamoDB")
	initializationFailedErr = errors.New("Initialization failed")
	sessionRequiredErr      = errors.New("Session is required")
	tableRequiredErr        = errors.New("Bad request: Table name required")
)
