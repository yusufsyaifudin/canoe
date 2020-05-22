package raftctrl

import (
	"context"
	"ysf/canoe/reply"
	"ysf/canoe/server"
)

func (h handler) stats(ctx context.Context, req server.Request) server.Response {
	return reply.Success(server.ReplyStructure{
		Type: "Join",
		Data: h.dep.GetGossip().Stats(),
	})
}
