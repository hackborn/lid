package dlock

import (
	"time"
)

// ------------------------------------------------------------
// LOCK-OPTS

type LockOpts struct {
	Force bool // If true then force the lock, even if someone else owns it.
}

// ------------------------------------------------------------
// SERVICE-OPTS

type ServiceOpts struct {
	Table      string        // Name of the table with lock data. NOTE: The package will manage this table, deleting it at will.
	Duration   time.Duration // The duration before the lock expires.
	TimeToLive time.Time     // Non-empty values will set a time to live on the lock table.
}
