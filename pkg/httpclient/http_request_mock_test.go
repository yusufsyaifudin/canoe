package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNewMock(t *testing.T) {
	convey.Convey("New httpclient.Mock", t, func() {
		convey.Convey("Return mock object", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)
		})
	})
}

func TestMock_AddCircuitBreaker(t *testing.T) {
	convey.Convey("New httpclient.Mock add hook", t, func() {
		convey.Convey("Return mock object then call add hook", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			// should do nothing
			client.AddCircuitBreaker(CBConfig{})
		})
	})
}

func TestMock_AddHook(t *testing.T) {
	convey.Convey("New httpclient.Mock add hook", t, func() {
		convey.Convey("Return mock object then call add hook", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			// should do nothing
			client.AddHook(nil)
		})
	})
}

func TestMock_Get(t *testing.T) {
	convey.Convey("New httpclient.Mock", t, func() {
		convey.Convey("When call Get then return not HttpResponse object", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			want := map[string]interface{}{
				"foo": "bar",
			}

			ctx := context.Background()
			client.On("Get", ctx, "/", http.Header{}).
				Return(want, nil)

			res, err := client.Get(ctx, "", "/", http.Header{})

			// should return empty HttpResponse on mock
			convey.So(res, convey.ShouldResemble, HttpResponse{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("not HttpResponse type"))
		})

		convey.Convey("When call Get then return expected", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			want := HttpResponse{
				RespBody: ioutil.NopCloser(bytes.NewReader([]byte("hello"))),
				CURL:     "CURL -X GET /",
				Raw:      ResponseRaw{},
			}

			ctx := context.Background()
			client.On("Get", ctx, "/", http.Header{}).
				Return(want, nil)

			res, err := client.Get(ctx, "", "/", http.Header{})
			convey.So(res, convey.ShouldResemble, want)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestMock_Post(t *testing.T) {
	convey.Convey("New httpclient.Mock", t, func() {
		convey.Convey("When call Post then return not HttpResponse object", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			want := map[string]interface{}{
				"foo": "bar",
			}

			ctx := context.Background()
			client.On("Post", ctx, "/", http.Header{}, []byte(nil)).
				Return(want, nil)

			res, err := client.Post(ctx, "", "/", http.Header{}, []byte(nil))

			// should return empty HttpResponse on mock
			convey.So(res, convey.ShouldResemble, HttpResponse{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("not HttpResponse type"))
		})

		convey.Convey("When call Post then return expected", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			want := HttpResponse{
				RespBody: ioutil.NopCloser(bytes.NewReader([]byte("hello"))),
				CURL:     "CURL -X POST /",
				Raw:      ResponseRaw{},
			}

			ctx := context.Background()
			client.On("Post", ctx, "/", http.Header{}, []byte(nil)).
				Return(want, nil)

			res, err := client.Post(ctx, "", "/", http.Header{}, []byte(nil))
			convey.So(res, convey.ShouldResemble, want)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestMock_Put(t *testing.T) {
	convey.Convey("New httpclient.Mock", t, func() {
		convey.Convey("When call Put then return not HttpResponse object", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			want := map[string]interface{}{
				"foo": "bar",
			}

			ctx := context.Background()
			client.On("Put", ctx, "/", http.Header{}, []byte(nil)).
				Return(want, nil)

			res, err := client.Put(ctx, "", "/", http.Header{}, []byte(nil))

			// should return empty HttpResponse on mock
			convey.So(res, convey.ShouldResemble, HttpResponse{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("not HttpResponse type"))
		})

		convey.Convey("When call Put then return expected", func() {
			client := NewMock()
			convey.So(client, convey.ShouldNotBeNil)

			want := HttpResponse{
				RespBody: ioutil.NopCloser(bytes.NewReader([]byte("hello"))),
				CURL:     "CURL -X PUT /",
				Raw:      ResponseRaw{},
			}

			ctx := context.Background()
			client.On("Put", ctx, "/", http.Header{}, []byte(nil)).
				Return(want, nil)

			res, err := client.Put(ctx, "", "/", http.Header{}, []byte(nil))
			convey.So(res, convey.ShouldResemble, want)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
