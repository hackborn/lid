package lidaws

import (
	"errors"
)

// ------------------------------------------------------------
// CONST and VAR

var (
	errConditionFailed      = errors.New("Condition failed")
	errDurationRequired     = errors.New("Bad request: Duration required")
	errDynamoRequired       = errors.New("Can't create DynamoDB")
	errInitializationFailed = errors.New("Initialization failed")
	errSessionRequired      = errors.New("Session is required")
	errTableRequired        = errors.New("Bad request: Table name required")
)
