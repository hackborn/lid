package dlock

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

// TestService() func provides scripted testing for the service, allowing
// complex command lists. It's a little painful to write tests, but there
// isn't much value in testing a single locking function.
func TestService(t *testing.T) {
	bootstraps := makeTestServices(t)

	cases := []struct {
		Script   string
		WantResp ScriptResponse
	}{
		// Successfully acquire empty lock
		{buildScript(lreq("s", "a", 0, false)), buildResp(lresp(LockOk, "", nil))},
		// Fail acquiring existing lock
		{buildScript(lreq("s1", "a", 0, false), lreq("s1", "b", 0, false)), buildResp(lresp(LockOk, "", nil), lresp(LockFailed, "", nil))},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for _, b := range bootstraps {
				runTestService(t, b, tc.Script, tc.WantResp)
			}
		})
	}
}

func runTestService(t *testing.T, b ServiceBootstrap, script string, wantResp ScriptResponse) {
	s := b.OpenService()
	defer b.CloseService()

	haveResp, err := runScript(script, s)
	if err != nil {
		panic("Error running script: " + err.Error())
	}
	if !wantResp.equals(haveResp) {
		fmt.Println("Mismatch have\n", haveResp, "\nwant\n", wantResp)
		t.Fatal()
	}
	//	t.Fatal()
}

// --------------------------------------------------------------------------------------
// BUILDING

func buildScript(elem ...interface{}) string {
	var b strings.Builder
	for _, e := range elem {
		data, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}
		b.WriteString(string(data))
	}
	return b.String()
}

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

func buildResp(elem ...[]interface{}) ScriptResponse {
	resp := ScriptResponse{}
	for _, e := range elem {
		resp.History = append(resp.History, e)
	}
	return resp
}

func lresp(status LockResponseStatus, previousDevice string, err error) []interface{} {
	resp := LockResponse{status, previousDevice}
	return []interface{}{resp, err}
}

// --------------------------------------------------------------------------------------
// COMPARING

func (a ScriptResponse) equals(b ScriptResponse) bool {
	if len(a.History) != len(b.History) {
		return false
	}
	for i, ah := range a.History {
		bh := b.History[i]
		if len(ah) != len(bh) {
			return false
		}
		for ii, ahh := range ah {
			if ahh != bh[ii] {
				return false
			}
		}
	}
	return true
}

// --------------------------------------------------------------------------------------
// TEST-CFG

// ServiceBootstrap is responsible for initializing and cleaning up a service during testing.
type ServiceBootstrap interface {
	OpenService() Service
	CloseService() error
}

type awsServiceBootstrap struct {
	tablename string
	sess      *session.Session
	service   *awsService
}

func (b *awsServiceBootstrap) OpenService() Service {
	// It's tough to test time-to-live:
	// * Not supported in dynalite
	// * Not supported in local dynamodb? (verify; add test if it is)
	// * Hosted dynamo has no guarantee on when the item is actually deleted.
	opts := ServiceOpts{Table: b.tablename, Duration: time.Second * 1}
	service, err := _newAwsServiceFromSession(opts, b.sess)
	if err != nil {
		panic(err)
	}
	b.service = service
	err = service.createTable()
	if err != nil {
		panic(err)
	}
	return service
}

func (b *awsServiceBootstrap) CloseService() error {
	b.service.deleteTable()
	b.service = nil
	return nil
}

// makeTestServices makes the test services for the testing configuration.
func makeTestServices(t *testing.T) []ServiceBootstrap {
	var services []ServiceBootstrap
	if !testing.Short() {
		// Currently we only test against a local dynamo.
		awskey0 := "DLOCK_TESTING_AWS_DYNAMO_ENDPOINT"

		// Check that the system is properly configured
		if os.Getenv(awskey0) == "" {
			fmt.Println("Can't do integration test, must have envvar", awskey0, "(use -short to disable)")
			t.Fatal()
		}

		val0 := os.Getenv(awskey0)

		cfg := &aws.Config{}
		cfg = cfg.WithRegion("us-west-2").WithEndpoint(val0)
		sess := session.Must(session.NewSession(cfg))
		tablename := "dlocktest_" + randomString(12)
		bootstrap := &awsServiceBootstrap{tablename: tablename, sess: sess}
		services = append(services, bootstrap)
	}
	return services
}

func randomString(size int) string {
	rs := rand.NewSource(time.Now().UnixNano())
	r := rand.New(rs)

	b := make([]byte, size)
	if _, err := r.Read(b); err != nil {
		return ""
	}
	ans := fmt.Sprintf("%X", b)
	if len(ans) > size {
		return ans[:size]
	}
	return ans
}
