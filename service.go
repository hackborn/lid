package dlock

import (
	"time"
)

// ------------------------------------------------------------
// SERVICE

// Service defines the contract for anything that can perform
// locking operations.
type Service interface {
	// Lock acquires the supplied lock. An error means the lock was not
	// acquired; success could be for various reasons supplied in the response.
	// The lock will be acquired if:
	// * It does not exist
	// * Or it does, and I own it
	// * Or it does, I don't own it, but my lock level is higher
	// * Or it does, I don't own it, but it's expired
	Lock(req LockRequest, opts *LockOpts) (LockResponse, error)

	// Unlock releases the supplied lock. Error is only returned if the lock
	// is owned by another signee; nil error means the lock no longer
	// exists, whether it did before or not.
	// The lock will be released if:
	// * It does not exist
	// * Or it does, and I own it
	Unlock(req UnlockRequest, opts *UnlockOpts) (UnlockResponse, error)
}

// ------------------------------------------------------------
// SERVICE-DEBUG

// ServiceDebug provides debugging functions on services. A
// nice service will only implement during testing.
type ServiceDebug interface {
	SetDuration(time.Duration)
}
