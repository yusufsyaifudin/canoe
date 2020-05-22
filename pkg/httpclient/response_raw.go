package httpclient

import (
	"net/http"
)

// ResponseRaw should return raw header and body
// it resemble http.Response
type ResponseRaw struct {
	Status           string      `json:"status"`      // e.g. "200 OK"
	StatusCode       int         `json:"status_code"` // e.g. 200
	Proto            string      `json:"proto"`       // e.g. "HTTP/1.0"
	ProtoMajor       int         `json:"proto_major"` // e.g. 1
	ProtoMinor       int         `json:"proto_minor"` // e.g. 0
	Header           http.Header `json:"header"`
	Body             interface{} `json:"body"` // already in go format
	ContentLength    int64       `json:"content_length"`
	TransferEncoding []string    `json:"transfer_encoding"`
	Uncompressed     bool        `json:"uncompressed"`
	//Trailer          http.Header

	// TLS contains information about the TLS connection on which the
	// response was received. It is nil for unencrypted responses.
	// The pointer is shared between responses and should not be
	// modified.
	//TLS *tls.ConnectionState
}
