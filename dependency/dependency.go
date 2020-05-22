package dependency

import (
	"ysf/canoe/gossip"
)

type Dep struct {
	raft gossip.Service
}

func (d *Dep) GetGossip() gossip.Service {
	return d.raft
}

func NewDep(raft gossip.Service) *Dep {
	return &Dep{
		raft: raft,
	}
}
