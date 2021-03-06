package lid

import (
	"time"
)

// ------------------------------------------------------------
// LOCK-OPTS

// LockOpts provides options for the Lock operation.
type LockOpts struct {
	Force      bool          `json:"force,omitempty"` // If true then force the lock, even if someone else owns it.
	Duration   time.Duration // Override the service default
	TimeToLive time.Duration // Override the service default
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
	TimeToLive time.Duration // Non-empty values will enable time to live on the lock table and expire items after the duration.
}
