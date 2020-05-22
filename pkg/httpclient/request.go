package httpclient

import (
	"crypto/tls"
	"mime/multipart"
	"net/http"
	"net/url"
)

// HttpRequestRaw resemble http.Request
type HttpRequest struct {
	Method           string               `json:"method"`
	URL              *url.URL             `json:"url"`
	Proto            string               `json:"proto"`       // "HTTP/1.0"
	ProtoMajor       int                  `json:"proto_major"` // 1
	ProtoMinor       int                  `json:"proto_minor"` // 0
	Header           http.Header          `json:"header"`
	Body             interface{}          `json:"body"`
	ContentLength    int64                `json:"content_length"`
	TransferEncoding []string             `json:"transfer_encoding"`
	Close            bool                 `json:"close"`
	Host             string               `json:"host"`
	Form             url.Values           `json:"form"`
	PostForm         url.Values           `json:"post_form"`
	MultipartForm    *multipart.Form      `json:"multipart_form"`
	Trailer          http.Header          `json:"trailer"`
	RemoteAddr       string               `json:"remote_addr"`
	RequestURI       string               `json:"request_uri"`
	TLS              *tls.ConnectionState `json:"tls"`
}
