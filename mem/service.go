package lidmem

import (
	"fmt"
	"github.com/hackborn/lid"
	"github.com/micro-go/lock"
	"sync"
	"time"
)

// ------------------------------------------------------------
// MEM-SERVICE

// memService provides an in-memory lid.Service implementation.
type memService struct {
	opts    lid.ServiceOpts
	mutex   sync.RWMutex
	records map[string]*record
}

// NewService constructs a new in-memory locking service.
func NewService(opts lid.ServiceOpts) (lid.Service, error) {
	records := make(map[string]*record)
	return &memService{opts: opts, records: records}, nil
}

func (s *memService) Lock(req lid.LockRequest, opts *lid.LockOpts) (lid.LockResponse, error) {
	if !req.IsValid() {
		return lid.LockResponse{}, lid.ErrBadRequest
	}

	endTimeFn := newEndTimeFn(&s.opts, opts)

	// First try a read
	r := s.find(req.Signature)
	if r != nil {
		return r.lock(req, opts, endTimeFn)
	}

	// Then a write
	defer lock.Write(&s.mutex).Unlock()
	r = s.records[req.Signature]
	if r != nil {
		return r.lock(req, opts, endTimeFn)
	}

	endTime := endTimeFn(time.Now())
	r = &record{signee: req.Signee, level: req.Level, endTime: endTime}
	s.records[req.Signature] = r
	return lid.LockResponse{Status: lid.LockOk}, nil
}

func (s *memService) Unlock(req lid.UnlockRequest, opts *lid.UnlockOpts) (lid.UnlockResponse, error) {
	if !req.IsValid() {
		return lid.UnlockResponse{}, lid.ErrBadRequest
	}

	r := s.find(req.Signature)
	if r == nil {
		return lid.UnlockResponse{Status: lid.UnlockNoLock}, nil
	}
	resp, err := r.unlock(req, opts)
	if resp.Status != lid.UnlockFailed {
		defer lock.Write(&s.mutex).Unlock()
		delete(s.records, req.Signature)
	}
	return resp, err
}

// Check() answers the state of a the requested lock. An error is answered
// if the lock doesn't exist.
// DO NOT USE THIS FUNCTION. It doesn't have much value, but exists as
// I transition a service to this library.
func (s *memService) Check(signature string) (lid.CheckResponse, error) {
	if signature == "" {
		return lid.CheckResponse{}, lid.ErrBadRequest
	}
	r := s.find(signature)
	if r != nil {
		return r.check(signature)
	}
	return lid.CheckResponse{}, lid.ErrNotFound
}

func (s *memService) find(signature string) *record {
	defer lock.Read(&s.mutex).Unlock()
	r, _ := s.records[signature]
	return r
}

// ------------------------------------------------------------
// RECORD

type record struct {
	mutex   sync.Mutex
	signee  string
	level   int
	endTime time.Time
}

func (r *record) lock(req lid.LockRequest, opts *lid.LockOpts, endTimeFn TimeFunc) (lid.LockResponse, error) {
	now := time.Now()
	endTime := endTimeFn(now)
	defer lock.Locker(&r.mutex).Unlock()
	if req.Signee == r.signee {
		r.level = req.Level
		r.endTime = endTime
		return lid.LockResponse{Status: lid.LockRenewed}, nil
	}
	if req.Level > r.level {
		prev := r.signee
		r.signee = req.Signee
		r.level = req.Level
		r.endTime = endTime
		return lid.LockResponse{Status: lid.LockTransferred, PreviousSignee: prev}, nil
	}
	if now.After(r.endTime) {
		prev := r.signee
		r.signee = req.Signee
		r.level = req.Level
		r.endTime = endTime
		return lid.LockResponse{Status: lid.LockTransferred, PreviousSignee: prev}, nil
	}
	return lid.LockResponse{Status: lid.LockFailed}, lid.ErrForbidden
}

func (r *record) unlock(req lid.UnlockRequest, opts *lid.UnlockOpts) (lid.UnlockResponse, error) {
	if r.signee != req.Signee {
		return lid.UnlockResponse{}, lid.ErrForbidden
	}
	return lid.UnlockResponse{Status: lid.UnlockOk}, nil
}

func (r *record) check(signature string) (lid.CheckResponse, error) {
	defer lock.Locker(&r.mutex).Unlock()
	return lid.CheckResponse{r.signee, r.level}, nil
}

// ------------------------------------------------------------
// FUNC

// TimeFunc converts one time to another
type TimeFunc func(time.Time) time.Time

// newEndTimeFn() answers a new TimeFunc that gets the current
// end time after applying all my options.
func newEndTimeFn(opts1 *lid.ServiceOpts, opts2 *lid.LockOpts) TimeFunc {
	return func(t time.Time) time.Time {
		if opts2 != nil && opts2.Duration != emptyDuration {
			return t.Add(opts2.Duration)
		}
		if opts1 != nil && opts1.Duration != emptyDuration {
			return t.Add(opts1.Duration)
		}
		return t
	}
}

// ------------------------------------------------------------
// BOILERPLATE

func serviceFmt() {
	fmt.Println()
}

// ------------------------------------------------------------
// CONST and VAR

var (
	emptyDuration = time.Second * 0
)
