package core

import (
	//"encoding/json"
	//"errors"
	//"fmt"
	//"github.com/hashicorp/raft"
	//"io"
	//"log"
	//"net/http"
	//"reflect"
	//"strings"
	"sync"
)

type mymap struct {
	data map[string]bool
	lock sync.Mutex
}

func (this *mymap) set(k string, v bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.data[k] = v
}

func (this *mymap) get() map[string]bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.data
}
