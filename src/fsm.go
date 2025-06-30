package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

var _ raft.FSM = &kvFsm{}

type kvFsm struct {
	db *sync.Map
}

func (kf *kvFsm) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		var sp setPayload
		err := json.Unmarshal(log.Data, &sp)
		if err != nil {
			return fmt.Errorf("Could not parse payload: %s", err)
		}

		kf.db.Store(sp.Key, sp.Value)
	default:
		return fmt.Errorf("Unknown raft log type: %#v", log.Type)
	}

	return nil
}

func (kf *kvFsm) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshotNoop{}, nil
}

func (kf *kvFsm) Restore(snapshot io.ReadCloser) error {
	return nil
}
