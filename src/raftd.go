package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	//"io/ioutil"
	"log"
	"os"
	//"path"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

func (this *Core) startRaft() {
	for {
		select {
		case <-newbie:
			if this.raftd == 1 {

				fmt.Printf("raftd: %v\n", this.raftd)
				continue
			}
			fmt.Printf("members len: %v\n", len(this.list.Members()))
			if len(this.list.Members()) < 3 {
				continue
			}
			if this.min == false {
				continue
			}
			if enableSingle {
				log.Println("start raft bootstrap")
				configuration := raft.Configuration{
					Servers: []raft.Server{
						{
							ID:      this.raftConfig.LocalID,
							Address: this.transport.LocalAddr(),
						},
					},
				}
				this.raft.BootstrapCluster(configuration)
				//var servers []raft.Server
				this.raftd = 1
			}
		}
	}
}

/*
func (this *Core) maintainRaftNode() {
	i := 0
	for {
		select {
		case <-maintainraft:
			if this.raft.State() == raft.Follower {
				continue
			}
			membersinlist := ""
			sep := ""
			for _, member := range this.list.Members() {
				membersinlist += sep
				membersinlist += fmt.Sprintf("%v", member.Addr)
				sep = ","
			}
			fmt.Printf("Members in list: %v\n", membersinlist)
			if this.raft.State() == raft.Leader {
				err := this.raftremove(membersinlist)
				if err != nil {
					continue
				}
			}
			if this.raft.State() == raft.Candidate {
				del, err := this.raftcheck(membersinlist)
				if err != nil {
					continue
				}
				if del {
					i++
					fmt.Printf("Candidate promote index: %v\n", i)
				} else {
					i = 0
				}
				if i > raftMaintainTotalTry {
					fmt.Printf("excced total try: %v, a Candidate would be promoted to maintain Raft net\n", i)
					err := this.raftsuicide()
					if err != nil {
						log.Printf("error in suicide: %v\n", err)
						continue
					}
					i = 0
				}
			}
		}
	}
}*/
func (this *Core) maintainRaftNode() {
	for {
		select {
		case <-maintainraft:
			if this.raft.State() != raft.Leader {
				continue
			}
			membersinlist := ""
			sep := ""
			for _, member := range this.list.Members() {
				membersinlist += sep
				membersinlist += fmt.Sprintf("%v", member.Addr)
				sep = ","
			}
			fmt.Printf("Members in list: %v\n", membersinlist)
			err := this.raftremove(membersinlist)
			if err != nil {
				continue
			}
		}
	}
}

func (this *Core) raftsuicide() error {
	this.createRaft()
	if enableSingle {
		log.Println("restart raft bootstrap")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      this.raftConfig.LocalID,
					Address: this.transport.LocalAddr(),
				},
			},
		}
		this.raft.BootstrapCluster(configuration)
	}
	return nil
}

func (this *Core) raftcheck(ips string) (bool, error) {
	configFuture := this.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		this.logger.Printf("failed to get raft configuration: %v", err)
		return false, err
	}
	for _, srv := range configFuture.Configuration().Servers {
		ipport := fmt.Sprintf("%v", srv.Address)
		ip := strings.Split(ipport, ":")[0]
		if strings.Contains(ips, ip) {
			continue
		}
		break
	}
	return true, nil
}

func (this *Core) raftremove(ips string) error {
	configFuture := this.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		this.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}
	//log.Printf("raft remove: %v\n", addr)
	for _, srv := range configFuture.Configuration().Servers {
		ipport := fmt.Sprintf("%v", srv.Address)
		ip := strings.Split(ipport, ":")[0]
		fmt.Printf("ip in remove: %v\n", ip)
		if strings.Contains(ips, ip) {
			continue
		}
		fmt.Printf("ip to remove: %v\n", ip)
		future := this.raft.RemoveServer(srv.ID, 0, 0)
		if err := future.Error(); err != nil {
			return fmt.Errorf("error removing existing node %v at %v: %v", id, ip, err)
		}
	}
	return nil
}

func (this *Core) addRaftNode() {
	for {
		select {
		case <-addraft:
			fmt.Println(this.raft.State())
			if this.raft.State() != raft.Leader {
				continue
			}
			for _, member := range this.list.Members() {
				ip := fmt.Sprintf("%v", member.Addr)
				if ip == this.ip {
					continue
				}
				id := strings.Split(ip, ".")[3]
				addr := fmt.Sprintf("%v:%v", ip, raftPort)
				err := this.raftjoin(id, addr)
				if err != nil {
					continue
				}
			}
		}
	}
}

func (this *Core) raftjoin(id, addr string) error {
	configFuture := this.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		this.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}
	log.Printf("raft add: %v\n", addr)
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(id) || srv.Address == raft.ServerAddress(addr) {
			if srv.ID == raft.ServerID(id) && srv.Address == raft.ServerAddress(addr) {
				this.logger.Printf("node %v at %v already member of cluster, ignoring", id, addr)
				return nil
			}
			future := this.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %v at %v: %v", id, addr, err)
			}
		}
	}
	future := this.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0)
	if future.Error != nil {
		return future.Error()
	}
	this.logger.Printf("node %s at %s joined successfully\n", id, addr)
	return nil
}

func (this *Core) createRaft() {
	config := raft.DefaultConfig()
	//config := raft.Config.DefaultConfig()
	config.LocalID = raft.ServerID(this.id)
	this.raftConfig = config
	addr, err := net.ResolveTCPAddr("tcp", this.RaftBind)
	if err != nil {
		log.Fatal(err)
	}
	//transport, err := raft.NetworkTransport.NewTCPTransport(this.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	transport, err := raft.NewTCPTransport(this.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
	this.transport = transport

	snapshots, err := raft.NewFileSnapshotStore(this.RaftDir, retainSnapshotCount, os.Stderr)
	//snapshots, err := raft.FileSnapshotStore.NewFileSnapshotStore(this.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		log.Fatalf("file snapshot store: %s\n", err)
	}

	var logStore raft.LogStore
	var stableStore raft.StableStore
	if this.inmem {
		//logStore = raft.InmemStore.NewInmemStore()
		logStore = raft.NewInmemStore()
		stableStore = raft.NewInmemStore()
		//stableStore = raft.InmemStore.NewInmemStore()
	} else {
		boltDB, err := raftboltdb.NewBoltStore(filepath.Join(this.RaftDir, "raft.db"))
		if err != nil {
			log.Fatalf("new bolt store: %\ns", err)
		}
		logStore = boltDB
		stableStore = boltDB
	}

	ra, err := raft.NewRaft(config, (*fsm)(this), logStore, stableStore, snapshots, transport)
	if err != nil {
		log.Fatalf("new raft: %s\n", err)
	}
	this.raft = ra
}

type fsm Core

// Apply applies a Raft log entry to the key-value store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "set":
		return f.applySet(c.Key, c.Value)
	case "delete":
		return f.applyDelete(c.Key)
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}
}

// Snapshot returns a snapshot of the key-value store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clone the map.
	o := make(map[string]string)
	for k, v := range f.m {
		o[k] = v
	}
	return &fsmSnapshot{store: o}, nil
}

// Restore stores the key-value store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	f.m = o
	return nil
}

func (f *fsm) applySet(key, value string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[key] = value
	return nil
}

func (f *fsm) applyDelete(key string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, key)
	return nil
}

type fsmSnapshot struct {
	store map[string]string
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}
