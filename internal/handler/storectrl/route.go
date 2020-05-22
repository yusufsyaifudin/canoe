package storectrl

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
			Path:       "/store/:key",
			Method:     "GET",
			Handler:    h.get,
			Middleware: nil,
		},
		{
			Path:       "/store",
			Method:     "POST",
			Handler:    h.post,
			Middleware: nil,
		},
	}
}
