package dlock

import (
	"time"
)

// ------------------------------------------------------------
// SERVICE

// Service defines the contract for anything that can perform
// locking operations.
type Service interface {
	// Acquire the supplied lock. An error means the lock was not
	// acquired; success could be for various reasons supplied in the response.
	Lock(req LockRequest, opts *LockOpts) (LockResponse, error)

	// Release the supplied lock. Error is only returned if the lock
	// is owned by another signee; nil error means the lock no longer
	// exists, whether it did before or not.
	Unlock(req UnlockRequest, opts *UnlockOpts) (UnlockResponse, error)
}

// ------------------------------------------------------------
// SERVICE-DEBUG

// ServiceDebug provides debugging functions on services. A
// nice service will only implement during testing.
type ServiceDebug interface {
	SetDuration(time.Duration)
}
