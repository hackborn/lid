package dlock

import ()

// Service defines the contract for anything that can perform
// locking operations.
type Service interface {
	Lock(req LockRequest, opts *LockOpts) (LockResponse, error)
}
