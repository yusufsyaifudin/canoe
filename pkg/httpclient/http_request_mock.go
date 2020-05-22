package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) AddCircuitBreaker(config CBConfig) {
	return
}

// AddHook Do nothing
func (m *Mock) AddHook(Hook) {
	return
}

func (m *Mock) Get(ctx context.Context, correlationID, path string, header http.Header) (ret HttpResponse, err error) {
	args := m.Called(ctx, path, header)

	ret, ok := args.Get(0).(HttpResponse)
	if !ok {
		return HttpResponse{}, fmt.Errorf("not HttpResponse type")
	}

	return ret, args.Error(1)
}

func (m *Mock) Post(
	ctx context.Context,
	correlationID,
	path string,
	requestHeader http.Header,
	requestBody []byte,
) (ret HttpResponse, err error) {
	args := m.Called(ctx, path, requestHeader, requestBody)

	ret, ok := args.Get(0).(HttpResponse)
	if !ok {
		return HttpResponse{}, fmt.Errorf("not HttpResponse type")
	}

	return ret, args.Error(1)
}

func (m *Mock) Put(ctx context.Context, correlationID, path string, requestHeader http.Header, requestBody []byte) (ret HttpResponse, err error) {
	args := m.Called(ctx, path, requestHeader, requestBody)

	ret, ok := args.Get(0).(HttpResponse)
	if !ok {
		return HttpResponse{}, fmt.Errorf("not HttpResponse type")
	}

	return ret, args.Error(1)
}

// NewMock implements AuthVirgoHttpRequester interface
func NewMock() *Mock {
	return &Mock{}
}
