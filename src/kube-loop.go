package core

import (
	//"fmt"
	"github.com/hashicorp/raft"
	"time"
)

func (this *Core) kubeLooper() {
	for {
		select {
		case <-time.After(1 * time.Minute):
			if this.raft.State() == raft.Leader {
				_, _, exist := this.reportKubeStatus()
				if exist {
					debug.Println("auto install kube, exist: true")
					continue
				}
				if this.check2installKube() {
					installkube <- "ok"
				}
			}
		}
	}
}

func (this *Core) check2installKube() bool {
	total := 3
	//count := 0
	debug.Printf("kube needed: %v\n", kubeneeded)
	for i := 0; i < total; i++ {
		debug.Printf("check: %v\n", i)
		// n => some num, which is const value of kubeneeded
		n := len(this.list.Members())
		debug.Printf("number of members: %v\n", n)
		if n < kubeneeded {
			return false
		}
		j := 0
		for _, other := range this.getOthers() {
			if cidrContains(this.cidr, other) {
				j++
			}
		}
		debug.Printf("number in the same CIDR - %v: %v\n", this.cidr, j)
		if j < kubeneeded-1 {
			return false
		}
		if i < total-1 {
			time.Sleep(time.Minute)
		}
	}
	debug.Println("check return true")
	return true
}
