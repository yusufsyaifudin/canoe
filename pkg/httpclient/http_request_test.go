package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// implements io.Reader
type nopReader struct {
	err error
}

func (r *nopReader) Read(p []byte) (n int, err error) {
	return len(p), r.err
}

// implements io.ReadCloser
type nopCloser struct {
	io.Reader
	err error
}

func (n *nopCloser) Close() error {
	return n.err
}

// noopCloser returns a ReadCloser with a no-op Close method wrapping
// the provided Reader r.
func noopCloser(r io.Reader, err error) io.ReadCloser {
	return &nopCloser{
		Reader: r,
		err:    err,
	}
}

type mockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	// just in case you want default correct return value
	return &http.Response{}, nil
}

func TestDefaultClient(t *testing.T) {
	convey.Convey("Test default client", t, func() {
		convey.Convey("When *http.Client is nil, should return panic", func() {
			defer func() {
				if r := recover(); r == nil {
					// function must be panic
					t.Errorf("It must be panic since *http.Client is nil, but it's not!")
					return
				}
			}()

			DefaultClient(nil)
		})

		convey.Convey("When *http.Client is not nil, it must return the instance", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)
		})
	})
}

func TestDefaultHttpRequester_AddHook(t *testing.T) {
	convey.Convey("Test AddHook", t, func() {
		convey.Convey("Don't add hook if it nil", func() {
			req := new(DefaultHttpRequester)

			// function don't return anything
			req.AddHook(nil)
		})

		convey.Convey("Add hook if it not nil", func() {
			req := new(DefaultHttpRequester)
			req.AddHook(new(NoopHook))
		})

		convey.Convey("Skip hook if nil, but call it func if not nil", func() {
			// this must call before and after hook, and call the function if it exist in slice of hook
			// skip it if in slice contains null (nil) value
			req := &DefaultHttpRequester{
				hook: []Hook{nil, new(NoopHook)},
			}

			req.beforeHook(context.Background(), HookData{})
			req.afterHook(context.Background(), HookData{})
		})
	})
}

func TestDefaultHttpRequester_AddCircuitBreaker(t *testing.T) {
	convey.Convey("Add new circuit breaker", t, func() {
		convey.Convey("Add circuit breaker should return no error", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			client.AddCircuitBreaker(CBConfig{})
			convey.So(client, convey.ShouldNotBeNil)
		})
	})
}

func TestDefaultHttpRequester_Get(t *testing.T) {
	convey.Convey("Test Get", t, func() {

		convey.Convey("Error url.Parse since it contain new line", func() {
			// to see what causing url.Parse return error, see https://golang.org/doc/go1.12#net/url
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Get(context.Background(), "", "http://example.com\n", http.Header{})
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Get should be success", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want

					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Get(context.Background(), "", "http://example.com/", http.Header{})
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Error client.Do request", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want

					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, fmt.Errorf("error client.Do request")
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Get(context.Background(), "", "http://example.com/", http.Header{})
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error response is nil", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return nil, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Get(context.Background(), "", "http://example.com/", http.Header{})
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error response is nil and error returned", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("error return when response nil")
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Get(context.Background(), "", "http://example.com/", http.Header{})
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error body close, but not return error in function (only in defer mode)", func() {
			// test this line:
			// if err := resp.Body.Close(); err != nil {

			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(bytes.NewReader([]byte("hello world")), fmt.Errorf("error closing body"))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Get(context.Background(), "", "http://example.com/", http.Header{})
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Error on ioutil.ReadAll body, but success closing body", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(
						&nopReader{
							err: fmt.Errorf("error read body"),
						},
						fmt.Errorf("error closing body"),
					)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Get(context.Background(), "", "http://example.com/", http.Header{})
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

	})
}

func TestDefaultHttpRequester_Post(t *testing.T) {
	convey.Convey("Test Post", t, func() {

		convey.Convey("Error url.Parse since it contain new line", func() {
			// to see what causing url.Parse return error, see https://golang.org/doc/go1.12#net/url
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Post(context.Background(), "", "http://example.com\n", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Post should be success", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want

					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Post(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Error client.Do request", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want

					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, fmt.Errorf("error client.Do request")
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Post(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error response is nil", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return nil, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Post(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error response is nil and error returned", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("error return when response nil")
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Post(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error body close, but not return error in function (only in defer mode)", func() {
			// test this line:
			// if err := resp.Body.Close(); err != nil {

			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(bytes.NewReader([]byte("hello world")), fmt.Errorf("error closing body"))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Post(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Error on ioutil.ReadAll body, but success closing body", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(
						&nopReader{
							err: fmt.Errorf("error read body"),
						},
						fmt.Errorf("error closing body"),
					)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Post(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

	})
}

func TestDefaultHttpRequester_Put(t *testing.T) {
	convey.Convey("Test Put", t, func() {

		convey.Convey("Error url.Parse since it contain new line", func() {
			// to see what causing url.Parse return error, see https://golang.org/doc/go1.12#net/url
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Put(context.Background(), "", "http://example.com\n", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Put should be success", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want

					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Put(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Error client.Do request", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want

					body := noopCloser(bytes.NewReader([]byte("hello world")), nil)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, fmt.Errorf("error client.Do request")
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Put(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error response is nil", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return nil, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Put(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error response is nil and error returned", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("error return when response nil")
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Put(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error body close, but not return error in function (only in defer mode)", func() {
			// test this line:
			// if err := resp.Body.Close(); err != nil {

			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(bytes.NewReader([]byte("hello world")), fmt.Errorf("error closing body"))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Put(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Error on ioutil.ReadAll body, but success closing body", func() {
			testClient := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// do whatever you want
					body := noopCloser(
						&nopReader{
							err: fmt.Errorf("error read body"),
						},
						fmt.Errorf("error closing body"),
					)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil
				},
			}

			client := DefaultClient(testClient)
			convey.So(client, convey.ShouldNotBeNil)

			resp, err := client.Put(context.Background(), "", "http://example.com/", http.Header{}, nil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

	})
}
