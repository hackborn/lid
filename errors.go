package dlock

import (
	"errors"
)

// mustErr() is a simple utility to panic on errors.
func mustErr(err error) {
	if err != nil {
		panic(err)
	}
}

// ------------------------------------------------------------
// CONST and VAR

var (
	conditionFailedErr      = errors.New("Condition failed")
	durationRequiredErr     = errors.New("Bad request: Duration required")
	dynamoRequiredErr       = errors.New("Can't create DynamoDB")
	initializationFailedErr = errors.New("Initialization failed")
	sessionRequiredErr      = errors.New("Session is required")
	tableRequiredErr        = errors.New("Bad request: Table name required")
)
