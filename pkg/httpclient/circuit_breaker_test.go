package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func Test_readyToTrip(t *testing.T) {
	convey.Convey("Ready to trip function", t, func() {
		convey.Convey("Should return true", func() {
			rtt := readyToTrip(0)
			convey.So(rtt, convey.ShouldNotBeNil)
		})
	})
}

func TestNewCircuitBreaker(t *testing.T) {
	convey.Convey("New Circuit Breaker", t, func() {
		convey.Convey("Should return not error", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			cb := NewCircuitBreaker(CBConfig{}, testClient)
			convey.So(cb, convey.ShouldNotBeNil)
		})
	})
}

func TestCircuitBreaker_Do(t *testing.T) {
	convey.Convey("New Circuit Breaker", t, func() {
		request := &http.Request{}
		request.URL = &url.URL{
			Path: "/",
		}

		convey.Convey("Not using circuit breaker", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			cb := NewCircuitBreaker(CBConfig{}, testClient)
			convey.So(cb, convey.ShouldNotBeNil)

			resp, err := cb.Do(request)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Using circuit breaker, return success", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			cb := NewCircuitBreaker(CBConfig{
				IsActive: true,
				Paths:    []string{"/"},
			}, testClient)
			convey.So(cb, convey.ShouldNotBeNil)

			resp, err := cb.Do(request)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Using circuit breaker, return error", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					return &http.Response{
						StatusCode: http.StatusOK,
					}, fmt.Errorf("error http")
				},
			}

			cb := NewCircuitBreaker(CBConfig{
				IsActive: true,
				Paths:    []string{"/"},
			}, testClient)
			convey.So(cb, convey.ShouldNotBeNil)

			resp, err := cb.Do(request)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Using circuit breaker, return server 500", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
					}, nil
				},
			}

			cb := NewCircuitBreaker(CBConfig{
				IsActive: true,
				Paths:    []string{"/"},
			}, testClient)
			convey.So(cb, convey.ShouldNotBeNil)

			resp, err := cb.Do(request)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

	})
}

func Test_pathInWhitelist(t *testing.T) {
	convey.Convey("New Circuit Breaker", t, func() {
		convey.Convey("Should return true when path is in list", func() {

			cb := &CircuitBreaker{
				whitelistPaths: []string{"/path"},
			}
			convey.So(cb, convey.ShouldNotBeNil)

			pathInWhiteList := cb.pathInWhitelist("/path")
			convey.So(pathInWhiteList, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when path is not in list", func() {

			cb := &CircuitBreaker{
				whitelistPaths: []string{"/path"},
			}
			convey.So(cb, convey.ShouldNotBeNil)

			pathInWhiteList := cb.pathInWhitelist("/")
			convey.So(pathInWhiteList, convey.ShouldBeFalse)
		})
	})
}
