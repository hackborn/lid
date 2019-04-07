package dlock

// ------------------------------------------------------------
// LOCK-OPTS

type LockOpts struct {
	Force bool // If true then force the lock, even if someone else owns it.
}
