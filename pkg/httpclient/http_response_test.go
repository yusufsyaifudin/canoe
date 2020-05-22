package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestHttpResponse_To(t *testing.T) {
	convey.Convey("HttpResponse.To", t, func() {
		convey.Convey("HttpResponse is nil", func() {
			resp := &HttpResponse{}
			resp = nil

			var data interface{}
			err := resp.To(context.Background(), &data)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("h is nil"))
		})

		convey.Convey("Response body is nil", func() {
			resp := &HttpResponse{
				RespBody: nil,
			}

			var data interface{}
			err := resp.To(context.Background(), &data)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("response body is nil"))
		})

		convey.Convey("Error on ioutil.ReadAll(h.RespBody)", func() {
			resp := &HttpResponse{
				RespBody: noopCloser(
					&nopReader{
						err: fmt.Errorf("error read body"),
					},
					nil,
				),
			}

			var data interface{}
			err := resp.To(context.Background(), &data)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Error on defer h.RespBody.Close()", func() {
			resp := &HttpResponse{
				RespBody: &nopCloser{
					Reader: bytes.NewReader([]byte(`{"foo": "bar"}`)), // must be json string
					err:    fmt.Errorf("error closing body"),
				},
			}

			var data interface{}
			err := resp.To(context.Background(), &data)

			// should not return error since it just closer error
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Success marshalling to struct", func() {
			resp := &HttpResponse{
				RespBody: &nopCloser{
					Reader: bytes.NewReader([]byte(`{"foo": "bar"}`)), // must be json string
					err:    nil,
				},
			}

			var want = map[string]string{
				"foo": "bar",
			}

			var data map[string]string
			err := resp.To(context.Background(), &data)

			convey.So(data, convey.ShouldResemble, want)
			convey.So(err, convey.ShouldBeNil)
		})

	})
}
