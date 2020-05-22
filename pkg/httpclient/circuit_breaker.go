package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sony/gobreaker"
)

type CBConfig struct {
	IsActive        bool
	Timeout         int
	IntervalTimeout int
	Threshold       int
	Paths           []string
}

type CircuitBreaker struct {
	client         HttpClient
	useBreaker     bool
	breaker        *gobreaker.CircuitBreaker
	whitelistPaths []string
}

var readyToTrip = func(threshold uint32) func(gobreaker.Counts) bool {
	return func(counts gobreaker.Counts) bool {
		return counts.TotalFailures >= threshold
	}
}

func NewCircuitBreaker(conf CBConfig, client HttpClient) *CircuitBreaker {
	cb := new(CircuitBreaker)
	cb.client = client
	cb.useBreaker = conf.IsActive
	cb.whitelistPaths = conf.Paths
	cb.breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Interval:    time.Duration(conf.IntervalTimeout) * time.Second,
		Timeout:     time.Duration(conf.Timeout) * time.Second,
		ReadyToTrip: readyToTrip(uint32(conf.Threshold)),
	})

	return cb
}

func (cb *CircuitBreaker) Do(request *http.Request) (*http.Response, error) {
	if cb.useBreaker && cb.pathInWhitelist(request.URL.String()) {
		cbResp, err := cb.breaker.Execute(func() (interface{}, error) {
			resp, err := cb.client.Do(request)
			if err != nil {
				return resp, err
			}

			if resp.StatusCode >= http.StatusInternalServerError {
				return resp, fmt.Errorf("error request to bank http status: %d", resp.StatusCode)
			}

			return resp, nil
		})

		var resp *http.Response
		if cbResp != nil {
			resp = cbResp.(*http.Response)
		}

		return resp, err
	}

	return cb.client.Do(request)
}

func (cb *CircuitBreaker) pathInWhitelist(path string) bool {
	for _, wp := range cb.whitelistPaths {
		u, _ := url.Parse(path)
		path = u.Path
		if wp == path {
			return true
		}
	}

	return false
}
