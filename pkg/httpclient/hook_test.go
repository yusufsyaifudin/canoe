package httpclient

import (
	"context"
	"testing"
)

// this test will do nothing except just call, since no operation happens inside the function
func TestNoopHook_BeforeRequest(t *testing.T) {
	req := new(NoopHook)
	req.BeforeRequest(context.Background(), HookData{})
}

func TestNoopHook_AfterRequest(t *testing.T) {
	req := new(NoopHook)
	req.AfterRequest(context.Background(), HookData{})
}
