package dlock

import (
	"time"
)

// ------------------------------------------------------------
// SERVICE

// Service defines the contract for anything that can perform
// locking operations.
type Service interface {
	Lock(req LockRequest, opts *LockOpts) (LockResponse, error)
}

// ------------------------------------------------------------
// SERVICE-DEBUG

// ServiceDebug provides debugging functions on services. A
// nice service will only implement during testing.
type ServiceDebug interface {
	SetDuration(time.Duration)
}
