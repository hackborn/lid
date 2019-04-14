package tack

import (
	"time"
)

// ------------------------------------------------------------
// LOCK-OPTS

// LockOpts provides options for the Lock operation.
type LockOpts struct {
	Force bool `json:"force,omitempty"` // If true then force the lock, even if someone else owns it.
}

// ------------------------------------------------------------
// UNLOCK-OPTS

// UnlockOpts is a placeholder in case we ever have options.
type UnlockOpts struct {
}

// ------------------------------------------------------------
// SERVICE-OPTS

// ServiceOpts provides standard options when constructing a service.
type ServiceOpts struct {
	Table      string        // Name of the table with lock data. NOTE: The package will manage this table, deleting it at will.
	Duration   time.Duration // The duration before the lock expires.
	TimeToLive time.Time     // Non-empty values will set a time to live on the lock table.
}
