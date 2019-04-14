package dlock

// ------------------------------------------------------------
// LOCK-RESPONSE

type LockResponse struct {
	Status         LockResponseStatus `json:"status,omitempty"`
	PreviousSignee string             `json:"previous_signee,omitempty"` // If I acquired a stale lock, this is the former owner
}

// Ok() answers true if the requester has the lock, regardless of
// previous state. See the status for more info.
func (a *LockResponse) Ok() bool {
	return a.Status != LockFailed
}

// Created() answers true if this operation actually created the lock. If
// it was already owned, this is false.
func (a *LockResponse) Created() bool {
	return a.Status == LockOk || a.Status == LockTransferred
}

// ------------------------------------------------------------
// UNLOCK-RESPONSE

type UnlockResponse struct {
	Status UnlockResponseStatus
}

func (r *UnlockResponse) Ok() bool {
	return r.Status != UnlockFailed
}

// ------------------------------------------------------------
// STATUS-RESPONSE

type StatusResponse struct {
	Signee string
	Level  string
}

// ------------------------------------------------------------
// CONST and VAR

type LockResponseStatus int

const (
	LockFailed      LockResponseStatus = iota // Someone else owns the lock
	LockOk                                    // The lock was free, now I own it
	LockTransferred                           // Someone else had a stale lock, now I own it
	LockRenewed                               // I previously owned it and still do
)

type UnlockResponseStatus int

const (
	UnlockFailed UnlockResponseStatus = iota // Someone else owns the lock
	UnlockOk                                 // The lock was unlocked, no one owns it
	UnlockNoLock                             // Technically I succeeded - there was nothing to unlock.
)
