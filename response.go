package lid

// ------------------------------------------------------------
// LOCK-RESPONSE

// LockResponse provides the output from the Lock function.
type LockResponse struct {
	Status         LockResponseStatus `json:"status,omitempty"`
	PreviousSignee string             `json:"previous_signee,omitempty"` // If I acquired a stale lock, this is the former owner
}

// Ok answers true if the requester has the lock, regardless of
// previous state. See the status for more info.
func (a *LockResponse) Ok() bool {
	return a.Status != LockFailed
}

// Created answers true if this operation actually created the lock. If
// it was already owned, this is false.
func (a *LockResponse) Created() bool {
	return a.Status == LockOk || a.Status == LockTransferred
}

// ------------------------------------------------------------
// UNLOCK-RESPONSE

// UnlockResponse provides the output from the Unlock function.
type UnlockResponse struct {
	Status UnlockResponseStatus `json:"status,omitempty"`
}

// Ok answers true if the lock no longer exists.
func (r *UnlockResponse) Ok() bool {
	return r.Status != UnlockFailed
}

// ------------------------------------------------------------
// CHECK-RESPONSE

// CheckResponse provides the state of a lock.
type CheckResponse struct {
	Signee string `json:"signee,omitempty"` // The owner of the lock.
	Level  int    `json:"level,omitempty"`  // The level of the lock.
}

// ------------------------------------------------------------
// CONST and VAR

// LockResponseStatus defines a status code for the response.
type LockResponseStatus int

// The status codes for the Lock response.
const (
	LockFailed      LockResponseStatus = iota // Someone else owns the lock
	LockOk                                    // The lock was free, now I own it
	LockTransferred                           // Someone else had a stale lock, now I own it
	LockRenewed                               // I previously owned it and still do
)

// UnlockResponseStatus defines a status code for the response.
type UnlockResponseStatus int

// The status codes for the Unlock response.
const (
	UnlockFailed UnlockResponseStatus = iota // Someone else owns the lock
	UnlockOk                                 // The lock was unlocked, no one owns it
	UnlockNoLock                             // Technically I succeeded - there was nothing to unlock.
)
