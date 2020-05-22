package raftctrl

import (
	"context"
	"fmt"
	"ysf/canoe/reply"
	"ysf/canoe/server"
)

type requestJoin struct {
	NodeId      string `json:"node_id"`
	RaftAddress string `json:"raft_address"`
}

func (h handler) join(ctx context.Context, req server.Request) server.Response {
	form := &requestJoin{}
	_ = req.Bind(form)

	if form.NodeId == "" {
		return reply.Error(server.ReplyStructure{
			Error: &server.ReplyErrorStructure{
				Code:    "",
				Title:   "Error join to leader",
				Message: "empty node id",
			},
			Type: server.ReplyError,
			Data: nil,
		})
	}

	if form.RaftAddress == "" {
		return reply.Error(server.ReplyStructure{
			Error: &server.ReplyErrorStructure{
				Code:    "",
				Title:   "Error join to leader",
				Message: "empty raft address",
			},
			Type: server.ReplyError,
			Data: nil,
		})
	}

	err := h.dep.GetGossip().Join(form.NodeId, form.RaftAddress)
	if err != nil {
		return reply.Error(server.ReplyStructure{
			Error: &server.ReplyErrorStructure{
				Code:    "",
				Title:   "Error join to leader",
				Message: fmt.Sprintf("%s", err.Error()),
			},
			Type: server.ReplyError,
			Data: nil,
		})
	}

	return reply.Success(server.ReplyStructure{
		Type: "Join",
		Data: map[string]interface{}{},
	})
}
