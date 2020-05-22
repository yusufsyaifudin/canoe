package gossip

import (
	"ysf/canoe/model"
)

type Service interface {
	Join(nodeID, addr string) error
	Stats() map[string]string
	DoOperation(payload model.CommandPayload) (value interface{}, err error)
	Shutdown() error
}
