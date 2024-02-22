package core

import (
	"time"
)

func (this *Core) kubeSummary(start time.Time) {
	this.kubelog.Println("Summary of Kubernetes installation:")
	nmaster := len(this.getKubeMaster())
	nnode := len(this.getKubeNode())
	var sep string
	// 0 master
	this.kubelog.Printf("number of masters: %v\n", nmaster)
	var masters string
	sep = ""
	for _, m := range this.getKubeMaster() {
		masters += sep
		masters += m
		sep = ", "
	}
	this.kubelog.Printf("which are: %v\n", masters)
	// 1 node
	if nnode == 0 {
		this.kubelog.Println("no node installed")
	} else {
		this.kubelog.Printf("number of nodes: %v\n", nnode)
		var nodes string
		sep = ""
		for _, n := range this.getKubeNode() {
			nodes += sep
			nodes += n
			sep = ", "
		}
		this.kubelog.Printf("which are: %v\n", nodes)
	}
	this.kubelog.Printf("VIP: %v\n", this.getKubeVIP())
	// 2 duration
	t := time.Now()
	elapsed := t.Sub(start)
	this.kubelog.Printf("installation elapsed: %v\n", elapsed)
}
