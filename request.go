package tack

// ------------------------------------------------------------
// LOCK-REQUEST

// LockRequest provides the parameters to the Lock function.
type LockRequest struct {
	Signature string `json:"sig,omitempty"` // The ID for this lock
	Signee    string `json:"sin,omitempty"` // The owner requesting the lock
	Level     int    `json:"lvl,omitempty"` // The level of lock requested. Leave this at the default 0 if you don't require levels.
}

func (r LockRequest) isValid() bool {
	return r.Signature != "" && r.Signee != ""
}

// ------------------------------------------------------------
// UNLOCK-REQUEST

// UnlockRequest provides the parameters to the Unlock function.
type UnlockRequest struct {
	Signature string `json:"sig,omitempty"` // The ID for this lock
	Signee    string `json:"sin,omitempty"` // The owner requesting the lock
}

func (r UnlockRequest) isValid() bool {
	return r.Signature != "" && r.Signee != ""
}
