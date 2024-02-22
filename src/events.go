package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/raft"
)

func (this *Core) sayhello() {
	node := this.list.LocalNode()
	str := fmt.Sprintf("hello from: %v:%v", node.Addr, node.Port)
	this.broadcasts.QueueBroadcast(&broadcast{
		msg:    []byte(str),
		notify: nil,
	})
}

func (this *Core) introduce() {
	for {
		select {
		case <-time.After(30 * time.Second):
			str := fmt.Sprintf("i %v:%v", this.ip, this.id)
			this.broadcasts.QueueBroadcast(&broadcast{
				msg:    []byte(str),
				notify: nil,
			})
		}
	}
}

func (this *Core) leaderinfo() {
	for {
		select {
		case <-time.After(15 * time.Second):
			fmt.Println(this.raft.Leader())
			if this.raft.State() != raft.Leader {
				continue
			}
			for _, member := range this.list.Members() {
				fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
			}
		}
	}
}

func (this *Core) listenCh() {
	for {
		select {
		case msg := <-id:
			//fmt.Println("ch got msg")
			this.handleID(msg)
		case msg := <-dsmsync:
			//fmt.Println("dsmsync ch got msg")
			this.handleDsm(msg)
		}
	}
}

func (this *Core) listenIDCh() {
	for {
		select {
		case msg := <-id:
			//fmt.Println("ch got msg")
			this.handleID(msg)
		}
	}
}

func (this *Core) listenDsmCh() {
	for {
		select {
		case msg := <-dsmsync:
			//fmt.Println("ch got msg")
			this.handleDsm(msg)
		}
	}
}

func (this *Core) listenReportCh() {
	for {
		select {
		case msg := <-report:
			//fmt.Println("ch got msg")
			this.handleReport(msg)
		}
	}
}

func (this *Core) chkMembers() {
	members := ""
	sep := ""
	for _, member := range this.list.Members() {
		members += sep
		members += fmt.Sprintf("%v", member.Addr)
		sep = ","
	}
	for id, ip := range this.members {
		if strings.Contains(members, ip) {
			continue
		} else {
			delete(this.members, id)
		}
	}
}

func (this *Core) handleID(b []byte) {
	msg := string(b)
	//fmt.Println("here")
	ipid := strings.Split(msg, ":")
	ip := ipid[0]
	id := ipid[1]
	//fmt.Printf("ip in handle: %v\n", ip)
	//fmt.Printf("id in handle: %v\n", id)
	//_, nodeid := this.members[ip]
	/*
		if nodeid {
			return
		}*/
	this.members[ip] = id
	//newbie <- "okay"
	maintainraft <- "okay"
	addraft <- "okay"
}

func (this *Core) handleDsm(b []byte) {
	msg := string(b)
	kv := strings.Split(msg, ",")
	k := kv[0]
	v := kv[1]
	this.dsmWrite(k, v)
}

func (this *Core) handleReport(b []byte) {
	//msg := string(b)
}

func (this *Core) evaluate() {
	for {
		select {
		case <-time.After(30 * time.Second):
			if this.raft.State() == raft.Leader {
				//m := this.getScores()
				//p := sortMapByValue(m)
			}
		}
	}
}

func (this *Core) listenKubelogCh() {
	for {
		select {
		case msg := <-log2kubelog:
			this.kubelog.Println(msg)
		}
	}
}
