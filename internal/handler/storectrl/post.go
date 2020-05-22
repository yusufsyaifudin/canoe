package storectrl

import (
	"context"
	"ysf/canoe/model"
	"ysf/canoe/reply"
	"ysf/canoe/server"
)

type requestPost struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (h handler) post(ctx context.Context, req server.Request) server.Response {
	dataToSave := &requestPost{}
	_ = req.Bind(dataToSave)

	cmd := model.CommandPayload{
		Operation: "SET",
		Key:       dataToSave.Key,
		Value:     dataToSave.Value,
	}

	data, err := h.dep.GetGossip().DoOperation(cmd)
	if err != nil {
		return reply.Error(err.Error())
	}

	return reply.Success(data)
}
