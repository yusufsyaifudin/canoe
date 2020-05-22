package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// HttpResponse mock http.Response
type HttpResponse struct {
	RespBody io.ReadCloser
	CURL     string
	Raw      ResponseRaw
}

// To unmarshal body to struct
func (h *HttpResponse) To(ctx context.Context, out interface{}) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "HttpResponse.To")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	if h == nil {
		return fmt.Errorf("h is nil")
	}

	if h.RespBody == nil {
		return fmt.Errorf("response body is nil")
	}

	bodyBytes, err := ioutil.ReadAll(h.RespBody)
	if err != nil {
		return fmt.Errorf("error read body response %s", err.Error())
	}

	// close after read
	if err := h.RespBody.Close(); err != nil {
		span.LogFields(log.String("error closing body", err.Error()))
	}

	// reuse response body for future read
	// http://blog.manugarri.com/how-to-reuse-http-response-body-in-golang/
	h.RespBody = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	err = json.Unmarshal(bodyBytes, out)
	return err
}
