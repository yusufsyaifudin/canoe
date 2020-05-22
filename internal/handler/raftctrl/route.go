package raftctrl

import (
	"ysf/canoe/dependency"
	"ysf/canoe/server"
)

func Routes(dep *dependency.Dep) []*server.Route {
	h := &handler{
		dep: dep,
	}
	return []*server.Route{
		{
			Path:       "/raft/join",
			Method:     "POST",
			Handler:    h.join,
			Middleware: nil,
		},
		{
			Path:       "/raft/stats",
			Method:     "GET",
			Handler:    h.stats,
			Middleware: nil,
		},
	}
}
