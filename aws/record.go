package lidaws

import (
	"time"
)

// ------------------------------------------------------------
// AWS-RECORD

// awsRecord stores a single entry in the lock table.
type awsRecord struct {
	Signature    string    `json:"lsig"`           // The ID for this lock. MUST MATCH awsSignatureKey
	Signee       string    `json:"lsignee"`        // The owner requesting the lock. MUST MATCH awsSigneeKey
	Level        int       `json:"llevel"`         // The level of lock requested. Leave this at the default 0 if you don't require levels. MUST MATCH awsLevelKey
	ExpiresEpoch int64     `json:"lexpires"`       // The time at which this lock expires (epoch). MUST MATCH awsExpiresKey
	Ttl          int64     `json:"lttl,omitempty"` // The TTL. Epoch seconds.
	Expires      time.Time `json:"-"`              // The time at which this lock expires. Convenience for clients.
}
