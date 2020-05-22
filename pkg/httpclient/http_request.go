package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"moul.io/http2curl"
)

// DefaultHttpRequester will simplify http request specific to this package's need
type DefaultHttpRequester struct {
	client HttpClient
	hook   []Hook
}

// DefaultClient will do http request using selected client.
// By using this, you can log http
func DefaultClient(client HttpClient) HttpRequester {
	if client == nil {
		panic("cannot use nil http client")
	}

	return &DefaultHttpRequester{
		client: client,
		hook:   make([]Hook, 0),
	}
}

func (r *DefaultHttpRequester) AddHook(h Hook) {
	if h == nil {
		return
	}

	r.hook = append(r.hook, h)
}

func (r DefaultHttpRequester) beforeHook(ctx context.Context, data HookData) {
	for _, hook := range r.hook {
		if hook == nil {
			continue
		}

		hook.BeforeRequest(ctx, data)
	}
}

func (r DefaultHttpRequester) afterHook(ctx context.Context, data HookData) {
	for _, hook := range r.hook {
		if hook == nil {
			continue
		}

		hook.AfterRequest(ctx, data)
	}
}

func (r *DefaultHttpRequester) AddCircuitBreaker(conf CBConfig) {
	r.client = NewCircuitBreaker(conf, r.client)
}

func (r DefaultHttpRequester) Get(ctx context.Context, correlationID, path string, header http.Header) (ret HttpResponse, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get")
	now := time.Now()
	request := &http.Request{}
	requestRaw := HttpRequest{}

	var openTracingRequestID = "no-trace-id"
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		openTracingRequestID = sc.String()
	}

	header.Set(httpHeaderSpanPropagatorKey, openTracingRequestID)

	if header.Get(correlationIDKey) == "" {
		header.Set(correlationIDKey, correlationID)
	}

	defer func() {
		r.afterHook(ctx, HookData{
			Error:         err,
			URL:           path,
			CURL:          ret.CURL,
			StartTime:     now,
			Request:       requestRaw,
			Response:      ret.Raw,
			CorrelationID: correlationID,
		})

		ctx.Done()
		span.Finish()
	}()

	request.Method = "GET"
	request.URL = &url.URL{}
	request.Header = header
	request.Body = ioutil.NopCloser(bytes.NewBuffer(nil))

	ret = HttpResponse{}
	ret.CURL = ""
	ret.Raw = ResponseRaw{}

	requestURL, err := url.Parse(path)
	if err != nil {
		err = fmt.Errorf("fail parse url %s: %s", path, err.Error())
		r.beforeHook(ctx, HookData{
			Error:         err,
			URL:           path,
			CURL:          ret.CURL,
			StartTime:     now,
			Request:       requestRaw,
			Response:      ret.Raw,
			CorrelationID: correlationID,
		})
		return ret, err
	}

	request.URL = requestURL
	request = request.WithContext(ctx)
	command, errCurl := http2curl.GetCurlCommand(request)
	if errCurl == nil {
		ret.CURL = command.String()
	}

	requestRaw = HttpRequest{
		Method:           request.Method,
		URL:              requestURL,
		Proto:            request.Proto,
		ProtoMajor:       request.ProtoMajor,
		ProtoMinor:       request.ProtoMinor,
		Header:           request.Header,
		Body:             "",
		ContentLength:    request.ContentLength,
		TransferEncoding: request.TransferEncoding,
		Close:            request.Close,
		Host:             request.Host,
		Form:             request.Form,
		PostForm:         request.PostForm,
		MultipartForm:    request.MultipartForm,
		Trailer:          request.Trailer,
		RemoteAddr:       request.RemoteAddr,
		RequestURI:       request.RequestURI,
		TLS:              request.TLS,
	}

	r.beforeHook(ctx, HookData{
		Error:         nil,
		URL:           path,
		CURL:          ret.CURL,
		StartTime:     now,
		Request:       requestRaw,
		Response:      ret.Raw,
		CorrelationID: correlationID,
	})

	span.LogFields(
		log.String("curl", ret.CURL),
	)

	resp, errHttp := r.client.Do(request)
	if resp == nil {
		if errHttp != nil {
			err = fmt.Errorf("error response http.Do is nil, err http %s", errHttp.Error())
			return
		}
		err = fmt.Errorf("error response http.Do is nil")
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			span.LogFields(
				log.String("error_close", err.Error()),
			)
			return
		}
	}()

	ret.Raw.Status = resp.Status
	ret.Raw.StatusCode = resp.StatusCode
	ret.Raw.Proto = resp.Proto
	ret.Raw.ProtoMajor = resp.ProtoMajor
	ret.Raw.ProtoMinor = resp.ProtoMinor
	ret.Raw.Header = resp.Header
	ret.Raw.Body = nil
	ret.Raw.ContentLength = resp.ContentLength
	ret.Raw.TransferEncoding = resp.TransferEncoding
	ret.Raw.Uncompressed = resp.Uncompressed

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error read body response %s", err.Error())
		return
	}

	var body interface{}
	ret.Raw.Body = string(bodyBytes)
	if err := json.Unmarshal(bodyBytes, &body); err == nil {
		ret.Raw.Body = body
	}

	ret.RespBody = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	span.LogFields(
		log.Object("response", ret),
	)

	if errHttp != nil {
		err = errHttp
	}

	return
}

func (r DefaultHttpRequester) Post(
	ctx context.Context,
	correlationID,
	path string,
	requestHeader http.Header,
	requestBody []byte,
) (ret HttpResponse, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "post")
	now := time.Now()
	request := &http.Request{}
	requestRaw := HttpRequest{}

	var openTracingRequestID = "no-trace-id"
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		openTracingRequestID = sc.String()
	}

	requestHeader.Set(httpHeaderSpanPropagatorKey, openTracingRequestID)

	if requestHeader.Get(correlationIDKey) == "" {
		requestHeader.Set(correlationIDKey, correlationID)
	}

	defer func() {
		r.afterHook(ctx, HookData{
			Error:         err,
			URL:           path,
			CURL:          ret.CURL,
			StartTime:     now,
			Request:       requestRaw,
			Response:      ret.Raw,
			CorrelationID: correlationID,
		})

		ctx.Done()
		span.Finish()
	}()

	request.Method = "POST"
	request.URL = &url.URL{}
	request.Header = requestHeader
	request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

	ret = HttpResponse{}
	ret.CURL = ""
	ret.Raw = ResponseRaw{}

	requestURL, err := url.Parse(path)
	if err != nil {
		err = fmt.Errorf("fail parse url %s: %s", path, err.Error())

		r.beforeHook(ctx, HookData{
			Error:         err,
			URL:           path,
			CURL:          ret.CURL,
			StartTime:     now,
			Request:       requestRaw,
			Response:      ResponseRaw{},
			CorrelationID: correlationID,
		})
		return ret, err
	}

	request.URL = requestURL
	request = request.WithContext(ctx)
	command, errCurl := http2curl.GetCurlCommand(request)
	if errCurl == nil {
		ret.CURL = command.String()
	}

	var reqBodyInterface interface{}
	if err := json.Unmarshal(requestBody, &reqBodyInterface); err != nil {
		reqBodyInterface = string(requestBody)
	}

	requestRaw = HttpRequest{
		Method:           request.Method,
		URL:              requestURL,
		Proto:            request.Proto,
		ProtoMajor:       request.ProtoMajor,
		ProtoMinor:       request.ProtoMinor,
		Header:           request.Header,
		Body:             reqBodyInterface,
		ContentLength:    request.ContentLength,
		TransferEncoding: request.TransferEncoding,
		Close:            request.Close,
		Host:             request.Host,
		Form:             request.Form,
		PostForm:         request.PostForm,
		MultipartForm:    request.MultipartForm,
		Trailer:          request.Trailer,
		RemoteAddr:       request.RemoteAddr,
		RequestURI:       request.RequestURI,
		TLS:              request.TLS,
	}

	r.beforeHook(ctx, HookData{
		Error:         nil,
		URL:           path,
		CURL:          ret.CURL,
		StartTime:     now,
		Request:       requestRaw,
		Response:      ret.Raw,
		CorrelationID: correlationID,
	})

	span.LogFields(
		log.String("curl", ret.CURL),
	)

	resp, errHttp := r.client.Do(request)
	if resp == nil {
		if errHttp != nil {
			err = fmt.Errorf("error response http.Do is nil, err http %s", errHttp.Error())
			return
		}
		err = fmt.Errorf("error response http.Do is nil")
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			span.LogFields(
				log.String("error_close", err.Error()),
			)
			return
		}
	}()

	ret.Raw.Status = resp.Status
	ret.Raw.StatusCode = resp.StatusCode
	ret.Raw.Proto = resp.Proto
	ret.Raw.ProtoMajor = resp.ProtoMajor
	ret.Raw.ProtoMinor = resp.ProtoMinor
	ret.Raw.Header = resp.Header
	ret.Raw.Body = nil
	ret.Raw.ContentLength = resp.ContentLength
	ret.Raw.TransferEncoding = resp.TransferEncoding
	ret.Raw.Uncompressed = resp.Uncompressed

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error read body response %s", err.Error())
		return
	}

	var body interface{}
	ret.Raw.Body = string(bodyBytes)
	if err := json.Unmarshal(bodyBytes, &body); err == nil {
		ret.Raw.Body = body
	}

	ret.RespBody = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	span.LogFields(
		log.Object("response", ret),
	)

	if errHttp != nil {
		err = errHttp
	}

	return
}

func (r DefaultHttpRequester) Put(
	ctx context.Context,
	correlationID,
	path string,
	requestHeader http.Header,
	requestBody []byte,
) (ret HttpResponse, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "put")
	now := time.Now()
	request := &http.Request{}
	requestRaw := HttpRequest{}

	var openTracingRequestID = "no-trace-id"
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		openTracingRequestID = sc.String()
	}

	requestHeader.Set(httpHeaderSpanPropagatorKey, openTracingRequestID)

	if requestHeader.Get(correlationIDKey) == "" {
		requestHeader.Set(correlationIDKey, correlationID)
	}

	defer func() {
		r.afterHook(ctx, HookData{
			Error:         err,
			URL:           path,
			CURL:          ret.CURL,
			StartTime:     now,
			Request:       requestRaw,
			Response:      ret.Raw,
			CorrelationID: correlationID,
		})

		ctx.Done()
		span.Finish()
	}()

	request.Method = "PUT"
	request.URL = &url.URL{}
	request.Header = requestHeader
	request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

	ret = HttpResponse{}
	ret.CURL = ""
	ret.Raw = ResponseRaw{}

	requestURL, err := url.Parse(path)
	if err != nil {
		err = fmt.Errorf("fail parse url %s: %s", path, err.Error())

		r.beforeHook(ctx, HookData{
			Error:         err,
			URL:           path,
			CURL:          ret.CURL,
			StartTime:     now,
			Request:       requestRaw,
			Response:      ResponseRaw{},
			CorrelationID: correlationID,
		})
		return ret, err
	}

	request.URL = requestURL
	request = request.WithContext(ctx)
	command, errCurl := http2curl.GetCurlCommand(request)
	if errCurl == nil {
		ret.CURL = command.String()
	}

	var reqBodyInterface interface{}
	if err := json.Unmarshal(requestBody, &reqBodyInterface); err != nil {
		reqBodyInterface = string(requestBody)
	}

	requestRaw = HttpRequest{
		Method:           request.Method,
		URL:              requestURL,
		Proto:            request.Proto,
		ProtoMajor:       request.ProtoMajor,
		ProtoMinor:       request.ProtoMinor,
		Header:           request.Header,
		Body:             reqBodyInterface,
		ContentLength:    request.ContentLength,
		TransferEncoding: request.TransferEncoding,
		Close:            request.Close,
		Host:             request.Host,
		Form:             request.Form,
		PostForm:         request.PostForm,
		MultipartForm:    request.MultipartForm,
		Trailer:          request.Trailer,
		RemoteAddr:       request.RemoteAddr,
		RequestURI:       request.RequestURI,
		TLS:              request.TLS,
	}

	r.beforeHook(ctx, HookData{
		Error:         nil,
		URL:           path,
		CURL:          ret.CURL,
		StartTime:     now,
		Request:       requestRaw,
		Response:      ret.Raw,
		CorrelationID: correlationID,
	})

	span.LogFields(
		log.String("curl", ret.CURL),
	)

	resp, errHttp := r.client.Do(request)
	if resp == nil {
		if errHttp != nil {
			err = fmt.Errorf("error response http.Do is nil, err http %s", errHttp.Error())
			return
		}
		err = fmt.Errorf("error response http.Do is nil")
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			span.LogFields(
				log.String("error_close", err.Error()),
			)
			return
		}
	}()

	ret.Raw.Status = resp.Status
	ret.Raw.StatusCode = resp.StatusCode
	ret.Raw.Proto = resp.Proto
	ret.Raw.ProtoMajor = resp.ProtoMajor
	ret.Raw.ProtoMinor = resp.ProtoMinor
	ret.Raw.Header = resp.Header
	ret.Raw.Body = nil
	ret.Raw.ContentLength = resp.ContentLength
	ret.Raw.TransferEncoding = resp.TransferEncoding
	ret.Raw.Uncompressed = resp.Uncompressed

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error read body response %s", err.Error())
		return
	}

	var body interface{}
	ret.Raw.Body = string(bodyBytes)
	if err := json.Unmarshal(bodyBytes, &body); err == nil {
		ret.Raw.Body = body
	}

	ret.RespBody = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	span.LogFields(
		log.Object("response", ret),
	)

	if errHttp != nil {
		err = errHttp
	}

	return
}
