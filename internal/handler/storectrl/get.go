package storectrl

import (
	"context"
	"ysf/canoe/model"
	"ysf/canoe/reply"
	"ysf/canoe/server"
)

func (h handler) get(ctx context.Context, req server.Request) server.Response {
	key := req.GetParam("key")

	cmd := model.CommandPayload{
		Operation: "GET",
		Key:       key,
		Value:     nil,
	}

	data, err := h.dep.GetGossip().DoOperation(cmd)
	if err != nil {
		return reply.Error(err.Error())
	}

	return reply.Success(data)
}
