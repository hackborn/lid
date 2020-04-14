package lidmem

import (
	"github.com/hackborn/lid"
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

func (s *memService) SetDuration(d time.Duration) {
	s.opts.Duration = d
}

// ------------------------------------------------------------
// TEST-CFG

type memServiceBootstrap struct {
	service lid.Service
}

func (b *memServiceBootstrap) OpenService() lid.Service {
	opts := lid.ServiceOpts{Duration: time.Second * 10}
	service, err := NewService(opts)
	lid.MustErr(err)
	b.service = service
	return service
}

func (b *memServiceBootstrap) CloseService() error {
	b.service = nil
	return nil
}

// makeTestServices makes the test services for the testing configuration.
func makeTestServices(t *testing.T) []lid.ServiceBootstrap {
	var services []lid.ServiceBootstrap
	bootstrap := &memServiceBootstrap{}
	services = append(services, bootstrap)
	return services
}
