package gossip

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
	"ysf/canoe/fsm"
	"ysf/canoe/model"
	"ysf/canoe/repo"

	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/hashicorp/raft"
)

const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
)

type handle struct {
	raft *raft.Raft
}

func New(nodeID, raftBindAddress, raftDir string, dataRepo repo.Service) (*handle, error) {
	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(nodeID)
	raftConf.SnapshotThreshold = 1024

	// For this example, we use in-memory database
	// Vault using BoltDb: https://www.vaultproject.io/docs/internals/integrated-storage
	// https://github.com/hashicorp/vault/blob/8813dc7363/physical/raft/fsm.go#L450
	// https://github.com/hashicorp/vault/blob/8813dc7363/physical/raft/fsm.go#L80
	// https://github.com/hashicorp/vault/blob/8813dc7363/physical/raft/fsm.go#L620-L632
	// Consul using MemDB https://www.consul.io/docs/internals/consensus.html#raft-protocol-overview
	fsmStore, err := fsm.NewFSM(dataRepo)
	if err != nil {
		return nil, err
	}

	// Create the backend raft store for logs and stable storage.
	// https://github.com/hashicorp/consul/blob/aa121bc8d2b270c836b58e548e1cc8989b2ef921/agent/consul/server.go#L690-L702
	// Vault also use like this,
	// https://github.com/hashicorp/vault/blob/8813dc7363fab378f9019e78c14118facac110cf/physical/raft/raft.go#L242-L257
	store, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft.dataRepo"))
	if err != nil {
		return nil, err
	}

	// Wrap the store in a LogCache to improve performance.
	cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", raftBindAddress)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(raftBindAddress, addr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(raftConf, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	// always start single server as a leader
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nodeID),
				Address: transport.LocalAddr(),
			},
		},
	}

	r.BootstrapCluster(configuration)

	return &handle{
		raft: r,
	}, nil
}

// Join handle when raft join
func (h handle) Join(nodeID, addr string) error {
	if h.raft.State() != raft.Leader {
		return fmt.Errorf("not a leader")
	}

	configFuture := h.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		fmt.Printf("failed to get raft configuration: %v\n", err)
		return err
	}

	for _, raftServer := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if raftServer.ID == raft.ServerID(nodeID) || raftServer.Address == raft.ServerAddress(addr) {
			// However if both the ID and the address are the same,
			// then nothing not even a join operation is needed.
			if raftServer.Address == raft.ServerAddress(addr) && raftServer.ID == raft.ServerID(nodeID) {
				fmt.Printf("node %s at %s already member of cluster, ignoring join request\n", nodeID, addr)
				return nil
			}

			future := h.raft.RemoveServer(raftServer.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	// This must be run on the leader or it will fail.
	f := h.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	fmt.Printf("node %s at %s joined successfully\n", nodeID, addr)
	return nil
}

// Join handle when raft join
func (h handle) Remove(nodeID, addr string) error {
	if h.raft.State() != raft.Leader {
		return fmt.Errorf("not a leader")
	}

	configFuture := h.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		fmt.Printf("failed to get raft configuration: %v\n", err)
		return err
	}

	for _, raftServer := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if raftServer.ID == raft.ServerID(nodeID) || raftServer.Address == raft.ServerAddress(addr) {
			// However if both the ID and the address are the same,
			// then nothing not even a join operation is needed.
			if raftServer.Address == raft.ServerAddress(addr) && raftServer.ID == raft.ServerID(nodeID) {
				fmt.Printf("node %s at %s already member of cluster, ignoring join request\n", nodeID, addr)
				return nil
			}

			future := h.raft.RemoveServer(raftServer.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	// This must be run on the leader or it will fail.
	f := h.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	fmt.Printf("node %s at %s joined successfully\n", nodeID, addr)
	return nil
}

func (h handle) Stats() map[string]string {
	return h.raft.Stats()
}

func (h handle) DoOperation(payload model.CommandPayload) (value interface{}, err error) {
	if h.raft.State() != raft.Leader {
		return nil, fmt.Errorf("not leader")
	}

	cmd, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// This must be run on the leader or it will fail.
	// https://github.com/hashicorp/raft/blob/v1.1.2/api.go#L669-L676
	future := h.raft.Apply(cmd, 1*time.Second)
	if future.Error() != nil {
		return nil, future.Error()
	}

	return future.Response(), nil
}

func (h handle) Shutdown() error {
	return h.raft.Shutdown().Error()
}
