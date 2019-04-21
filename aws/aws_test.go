package lidaws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hackborn/lid"
	"math/rand"
	"os"
	"testing"
	"time"
)

// TestService provides scripted testing for the service, allowing
// chained command lists. It's a little painful to write tests, but there
// isn't much value in testing a single locking function.
func TestService(t *testing.T) {
	suites := makeTestServices(t)

	lid.RunTestServiceSuite(t, suites)
}

// ------------------------------------------------------------
// SERVICE DEBUG

func (s *awsService) SetDuration(d time.Duration) {
	s.opts.Duration = d
}

// ------------------------------------------------------------
// TEST-CFG

type awsServiceBootstrap struct {
	tablename string
	sess      *session.Session
	service   *awsService
}

func (b *awsServiceBootstrap) OpenService() lid.Service {
	// It's tough to test time-to-live:
	// * Not supported in dynalite
	// * Not supported in local dynamodb? (verify; add test if it is)
	// * Hosted dynamo has no guarantee on when the item is actually deleted.
	opts := lid.ServiceOpts{Table: b.tablename, Duration: time.Second * 10}
	service, err := _newAwsServiceFromSession(opts, b.sess)
	lid.MustErr(err)
	b.service = service
	err = service.createTable()
	lid.MustErr(err)
	return service
}

func (b *awsServiceBootstrap) CloseService() error {
	b.service.deleteTable()
	b.service = nil
	return nil
}

// makeTestServices makes the test services for the testing configuration.
func makeTestServices(t *testing.T) []lid.ServiceBootstrap {
	var services []lid.ServiceBootstrap
	if !testing.Short() {
		// Currently we only test against a local dynamo.
		awskey0 := "LID_TESTING_AWS_DYNAMO_ENDPOINT"

		// Check that the system is properly configured
		if os.Getenv(awskey0) == "" {
			fmt.Println("Can't do integration test, must have envvar", awskey0, "(use -short to disable)")
			t.Fatal()
		}

		val0 := os.Getenv(awskey0)

		cfg := &aws.Config{}
		cfg = cfg.WithRegion("us-west-2").WithEndpoint(val0)
		sess := session.Must(session.NewSession(cfg))
		tablename := "lidtest_" + randomString(12)
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
