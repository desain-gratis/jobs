package main

import (
	"github.com/hashicorp/raft"
)

var _ raft.FSMSnapshot = &snapshotNoop{}

type snapshotNoop struct{}

func (sn *snapshotNoop) Persist(_ raft.SnapshotSink) error { return nil }
func (sn *snapshotNoop) Release() {

}
