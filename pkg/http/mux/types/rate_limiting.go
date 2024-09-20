package types

import (
	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelNet "github.com/Motmedel/utils_go/pkg/net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	insertionIndex   int
	Bucket           []*time.Time
	mutex            sync.Mutex
	NumOccupied      int
	NumSecondsExpiry int
}

func (rateLimiter *RateLimiter) Claim() (*time.Time, bool) {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()

	if rateLimiter.NumOccupied == len(rateLimiter.Bucket) {
		return rateLimiter.Bucket[rateLimiter.insertionIndex], true
	}

	expirationTime := time.Now().Add(time.Duration(rateLimiter.NumSecondsExpiry) * time.Second)

	currentInsertionIndex := rateLimiter.insertionIndex
	rateLimiter.Bucket[currentInsertionIndex] = &expirationTime
	rateLimiter.insertionIndex = (currentInsertionIndex + 1) % len(rateLimiter.Bucket)
	rateLimiter.NumOccupied += 1

	time.AfterFunc(expirationTime.Sub(time.Now()), func() {
		rateLimiter.mutex.Lock()
		defer rateLimiter.mutex.Unlock()

		rateLimiter.Bucket[currentInsertionIndex] = nil
		rateLimiter.NumOccupied -= 1
	})

	return &expirationTime, false
}

func DefaultGetRateLimitingKey(request *http.Request) (string, error) {
	if request == nil {
		return "", nil
	}

	remoteAddr := request.RemoteAddr
	ipAddress, _, err := motmedelNet.SplitAddress(remoteAddr)
	if err != nil {
		return "", &errors.InputError{
			Message: "An error occurred when splitting a remote address.",
			Cause:   err,
			Input:   remoteAddr,
		}
	}

	return ipAddress, nil
}

type TimerRateLimiter struct {
	RateLimiter
	Timer *time.Timer
}

type RateLimitingLookup struct {
	Map   map[string]*TimerRateLimiter
	Mutex sync.Mutex
}

type RateLimitingConfiguration struct {
	NumRequests          int
	NumSecondsExpiration int
	GetKey               func(*http.Request) (string, error)
	Lookup               RateLimitingLookup
}
