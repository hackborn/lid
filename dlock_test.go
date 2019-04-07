package dlock

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	bootstraps := makeTestServices(t)

	cases := []struct {
		Req      LockRequest
		WantResp LockResponse
		WantErr  error
	}{
		{LockRequest{}, LockResponse{}, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for _, b := range bootstraps {
				runTestLock(t, b, tc.Req, tc.WantResp, tc.WantErr)
			}
		})
	}
}

func runTestLock(t *testing.T, b ServiceBootstrap, req LockRequest, wantResp LockResponse, wantErr error) {
	s := b.OpenService()
	defer b.CloseService()

	haveResp, haveErr := s.Lock(req, nil)
	fmt.Println("haveResp", haveResp, "haveErr", haveErr)
	t.Fatal()
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
	service, err := _newAwsServiceFromSession(b.sess)
	if err != nil {
		panic(err)
	}
	b.service = service
	err = service.createTable(b.tablename)
	// AWS has a variety of error types; some aren't public and you have to probe them.
	// This happens because of the Time to Live, which is not available outside of the managed service.
	if err != nil && !strings.HasPrefix(err.Error(), "UnknownOperationException") {
		panic(err)
	}
	return service
}

func (b *awsServiceBootstrap) CloseService() error {
	b.service.deleteTable(b.tablename)
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