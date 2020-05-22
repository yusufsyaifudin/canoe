package httpclient

import (
	"context"
	"time"
)

type Hook interface {
	BeforeRequest(ctx context.Context, data HookData)
	AfterRequest(ctx context.Context, data HookData)
}

type HookData struct {
	Error         error       `json:"error"`
	URL           string      `json:"url"`
	CURL          string      `json:"curl"`
	StartTime     time.Time   `json:"start_time"`
	Request       HttpRequest `json:"request"`
	Response      ResponseRaw `json:"response"`
	CorrelationID string      `json:"correlation_id"`
}

type NoopHook struct{}

func (NoopHook) BeforeRequest(_ context.Context, _ HookData) {
	return
}

func (NoopHook) AfterRequest(_ context.Context, _ HookData) {
	return
}
