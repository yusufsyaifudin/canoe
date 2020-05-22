package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"ysf/canoe/model"
	"ysf/canoe/repo"

	"github.com/hashicorp/raft"
)

type FSM struct {
	db repo.Service
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (s FSM) Apply(log *raft.Log) interface{} {
	switch log.Type {
	case raft.LogCommand:
		var payload = model.CommandPayload{}
		if err := json.Unmarshal(log.Data, &payload); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error marshalling store payload %s\n", err.Error())
			return nil
		}

		op := strings.ToUpper(strings.TrimSpace(payload.Operation))
		switch op {
		case "SET":
			if err := s.db.Set(payload.Key, payload.Value); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error save data %s\n", err.Error())
				return nil
			}
			return payload.Value
		case "GET":
			return s.db.Get(payload.Key)
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "not raft log command type\n")
	return nil
}

// snapshotNoop will be called during make snapshotNoop.
// snapshotNoop is used to support log compaction.
// No need to call snapshot since it already persisted in disk (using BoltDb) when raft calling Apply function.
// TODO: using badger db to save snapshot in different directory?
func (s FSM) Snapshot() (raft.FSMSnapshot, error) {
	return newSnapshotNoop()
}

// Restore is used to restore an FSM from a snapshotNoop. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
// Restore will update all data in BadgerDB
func (s FSM) Restore(rClose io.ReadCloser) error {
	defer func() {
		if err := rClose.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[FINALLY RESTORE] close error %s\n", err.Error())
		}
	}()

	_, _ = fmt.Fprintf(os.Stdout, "[START RESTORE] read all message from snapshot\n")
	var totalRestored int

	decoder := json.NewDecoder(rClose)
	for decoder.More() {
		var data = &model.CommandPayload{}
		err := decoder.Decode(data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error decode data %s\n", err.Error())
			return err
		}

		if err := s.db.Set(data.Key, data.Value); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error persist data %s\n", err.Error())
			return err
		}

		totalRestored++
	}

	// read closing bracket
	_, err := decoder.Token()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error %s\n", err.Error())
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] success restore %d messages in snapshot\n", totalRestored)
	return nil
}

// NewFSM return implemented interface of raft.FSM
// FSM is to manage replicated state machines.
// Finite State Machine (FSM) provides an interface that can be implemented by
// clients to make use of the replicated log.
// This is use BadgerDB. You can change it using other persistent database.
func NewFSM(db repo.Service) (raft.FSM, error) {
	return &FSM{
		db: db,
	}, nil
}
