package dlock

import (
	"time"
)

// ------------------------------------------------------------
// LOCK-STATE

type LockState struct {
	Signature  string    `json:"dsig,omitempty"`     // The ID for this lock
	Signee     string    `json:"dsignee,omitempty"`  // The owner requesting the lock
	Level      int       `json:"dlevel,omitempty"`   // The level of lock requested. Leave this at the default 0 if you don't require levels.
	Expires    time.Time `json:"-"`                  // The time at which this lock expires
	ExpiresInt int64     `json:"dexpires,omitempty"` // The int value of the expiration time.
}
