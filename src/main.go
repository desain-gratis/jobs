package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func main() {

	nodeID := ""
	raftAddress := ""

	dataDir := "data"

	flag.StringVar(&nodeID, "node-id", "", "raft node id")
	flag.StringVar(&raftAddress, "raft-addr", "", "raft address")

	flag.Parse()

	log.Println(nodeID, raftAddress)
	r, _ := setupRaft(path.Join(dataDir, "raft-"+nodeID), nodeID, raftAddress, &kvFsm{})

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		token := sc.Text()
		tokens := strings.Split(token, " ")

		fmt.Println("Adding as voter:", tokens[0], tokens[1])
		err := r.AddVoter(raft.ServerID(tokens[0]), raft.ServerAddress(tokens[1]), 0, 0).Error()
		if err != nil {
			log.Println("err add voter", err)
		}
	}
}

func setupRaft(dir, nodeId, raftAddress string, kf *kvFsm) (*raft.Raft, error) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Could not create data directory: %s", err)
	}

	store, err := raftboltdb.NewBoltStore(path.Join(dir, "bolt"))
	if err != nil {
		return nil, fmt.Errorf("Could not create bolt store: %s", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(path.Join(dir, "snapshot"), 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("Could not create snapshot store: %s", err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftAddress)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve address: %s", err)
	}

	transport, err := raft.NewTCPTransport(raftAddress, tcpAddr, 10, time.Second*10, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("Could not create tcp transport: %s", err)
	}

	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(nodeId)

	r, err := raft.NewRaft(raftCfg, kf, store, store, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("Could not create raft instance: %s", err)
	}

	// Cluster consists of unjoined leaders. Picking a leader and
	// creating a real cluster is done manually after startup.
	r.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nodeId),
				Address: transport.LocalAddr(),
			},
		},
	})

	return r, nil
}
