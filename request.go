package dlock

// ------------------------------------------------------------
// LOCK-REQUEST

type LockRequest struct {
	Signature string // The ID for this lock
	Signee    string // The owner requesting the lock
	Level     int    // The level of lock requested. Leave this at the default 0 if you don't require levels.
}

// ------------------------------------------------------------
// UNLOCK-REQUEST

type UnlockRequest struct {
	Signature string // The ID for this lock
	Signee    string // The owner requesting the lock
}
