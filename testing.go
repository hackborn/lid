package lid

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// RunTestServiceSuite provides scripted testing for the service, allowing
// chained command lists. It's a little painful to write tests, but there
// isn't much value in testing a single locking function.
// This is the main test function for any service implementations: Send
// in a bootstrap on the service for the standard testing.
func RunTestServiceSuite(t *testing.T, suites []ServiceBootstrap) {
	cases := []struct {
		Script   string
		WantResp scriptResponse
	}{
		// Acquire empty lock
		{buildScript(lreq("a", "0", 0, false)), buildResp(lresp(LockOk, "", nil))},
		// Acquire existing lock through higher level
		{buildScript(lreq("a", "0", 0, false), lreq("a", "1", 1, false)), buildResp(lresp(LockOk, "", nil), lresp(LockTransferred, "0", nil))},
		// Renew my existing lock
		{buildScript(durS(-20), lreq("a", "0", 0, false), durS(10), lreq("a", "0", 1, false)), buildResp(lresp(LockOk, "", nil), lresp(LockRenewed, "", nil))},
		// Acquire someone else's expired lock
		{buildScript(durS(-20), lreq("a", "0", 0, false), durS(10), lreq("a", "1", 0, false)), buildResp(lresp(LockOk, "", nil), lresp(LockTransferred, "0", nil))},
		// Fail acquiring existing, valid lock
		{buildScript(lreq("a", "0", 0, false), lreq("a", "1", 0, false)), buildResp(lresp(LockOk, "", nil), lresp(LockFailed, "", ErrForbidden))},
		// Unlock a missing lock
		{buildScript(ulreq("a", "0")), buildResp(ulresp(UnlockNoLock, nil))},
		// Unlock an existing lock
		{buildScript(lreq("a", "0", 0, false), ulreq("a", "0")), buildResp(lresp(LockOk, "", nil), ulresp(UnlockOk, nil))},
		// Fail unlocking someone else's lock
		{buildScript(lreq("a", "0", 0, false), ulreq("a", "1")), buildResp(lresp(LockOk, "", nil), ulresp(UnlockFailed, ErrForbidden))},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for _, b := range suites {
				runTestService(t, b, tc.Script, tc.WantResp)
			}
		})
	}
}

func runTestService(t *testing.T, b ServiceBootstrap, script string, wantResp scriptResponse) {
	s := b.OpenService()
	defer b.CloseService()

	haveResp, err := runScript(script, s)
	MustErr(err)
	if !wantResp.equals(haveResp) {
		fmt.Println("Mismatch have\n", haveResp, "\nwant\n", wantResp)
		t.Fatal()
	}
	// t.Fatal()
}

// ------------------------------------------------------------
// BUILDING

func buildScript(elem ...interface{}) string {
	var b strings.Builder
	for _, e := range elem {
		data, err := json.Marshal(e)
		MustErr(err)
		b.WriteString(string(data))
	}
	return b.String()
}

// durS creates a scripting object that applies a new duration to the service.
func durS(seconds int64) interface{} {
	cmd := make(map[string]interface{})
	cmd[durCmd] = time.Duration(seconds) * time.Second
	return cmd
}

// lreq returns a scripting object to create a lock request.
func lreq(signature, signee string, level int, force bool) interface{} {
	body := make(map[string]interface{})
	body["req"] = LockRequest{signature, signee, level}
	if force {
		body["opts"] = LockOpts{true}
	}
	cmd := make(map[string]interface{})
	cmd[lockCmd] = body
	return cmd
}

// ulreq returns a scripting object to create an unlock request.
func ulreq(signature, signee string) interface{} {
	body := make(map[string]interface{})
	body["req"] = UnlockRequest{signature, signee}
	cmd := make(map[string]interface{})
	cmd[unlockCmd] = body
	return cmd
}

func buildResp(elem ...[]interface{}) scriptResponse {
	resp := scriptResponse{}
	for _, e := range elem {
		resp.History = append(resp.History, e)
	}
	return resp
}

// lresp creates a response for a script lock request.
func lresp(status LockResponseStatus, previousDevice string, err error) []interface{} {
	resp := LockResponse{status, previousDevice}
	return []interface{}{resp, err}
}

// ulresp creates a response for a script unlock request.
func ulresp(status UnlockResponseStatus, err error) []interface{} {
	resp := UnlockResponse{status}
	return []interface{}{resp, err}
}

// ------------------------------------------------------------
// COMPARING

func (a scriptResponse) equals(b scriptResponse) bool {
	if len(a.History) != len(b.History) {
		return false
	}
	for i, ah := range a.History {
		bh := b.History[i]
		if len(ah) != len(bh) {
			return false
		}
		for ii, ahh := range ah {
			if !interfaceEquals(ahh, bh[ii]) {
				return false
			}
		}
	}
	return true
}

func interfaceEquals(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == b {
		return true
	}
	switch aa := a.(type) {
	case error:
		if bb, ok := b.(error); ok {
			return aa.Error() == bb.Error()
		}
	}
	return false
}

// ------------------------------------------------------------
// SERVICE-BOOTSTRAP

// ServiceBootstrap is responsible for initializing and cleaning up a service during testing.
type ServiceBootstrap interface {
	OpenService() Service
	CloseService() error
}
