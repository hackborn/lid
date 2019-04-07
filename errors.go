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
	conditionFailedErr = errors.New("Condition failed")
	dynamoRequiredErr  = errors.New("Can't create DynamoDB")
	sessionRequiredErr = errors.New("Session is required")
)
