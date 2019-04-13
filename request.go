package dlock

// ------------------------------------------------------------
// LOCK-REQUEST

type LockRequest struct {
	Signature string `json:"sig,omitempty"` // The ID for this lock
	Signee    string `json:"sin,omitempty"` // The owner requesting the lock
	Level     int    `json:"lvl,omitempty"` // The level of lock requested. Leave this at the default 0 if you don't require levels.
}

// ------------------------------------------------------------
// UNLOCK-REQUEST

type UnlockRequest struct {
	Signature string `json:"sig,omitempty"` // The ID for this lock
	Signee    string `json:"sin,omitempty"` // The owner requesting the lock
}
