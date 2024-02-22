package core

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/raft"
)

func (this *Core) Get(key string) (string, error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.m[key], nil
}

func (this *Core) Set(key, value string) error {
	if this.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}
	c := &command{
		Op:    "set",
		Key:   key,
		Value: value,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f := this.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (this *Core) Delete(key string) error {
	if this.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}
	c := &command{
		Op:  "set",
		Key: key,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f := this.raft.Apply(b, raftTimeout)
	return f.Error()
}
