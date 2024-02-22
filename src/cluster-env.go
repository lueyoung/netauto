package core

import (
	//"errors"
	"fmt"
	//"io/ioutil"
	"log"
	//"os"
	"runtime"
	"strings"
	"sync"
)

func (this *Core) makeBaseEnvVariables() error {
	cmd := string(makeBaseEnvVariables)
	cmd = strings.Replace(cmd, "{{.cluster.cidr}}", clusterCidr, -1)
	execCmd(cmd)
	err := this.distribute("/etc/kubernetes/env/k8s.env")
	if err != nil {
		return err
	}
	err = this.distribute("/etc/kubernetes/token.csv")
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) makeIpEnvVariables() error {
	f := new(file)
	path := "/etc/kubernetes/env/this-ip.env"
	nums := strings.Split(this.ip, ".")
	netid := fmt.Sprintf("%v.%v.%v", nums[0], nums[1], nums[2])
	vip := this.getKubeVIP()
	tmp := string(thisip)
	tmp = strings.Replace(tmp, "{{.net.id}}", netid, -1)
	tmp = strings.Replace(tmp, "{{.vip}}", vip, -1)
	tmp = strings.Replace(tmp, "{{.port}}", fmt.Sprintf("%v", kubeMasterPort), -1)
	// self
	content := tmp
	content = strings.Replace(content, "{{.this.ip}}", this.ip, -1)
	f.set(path, content)
	f.new()
	// others
	for _, ip := range this.getOthers() {
		content := tmp
		content = strings.Replace(content, "{{.this.ip}}", ip, -1)
		log2kubelog <- fmt.Sprintf("ip: %v,content: %v", ip, content)
		log2kubelog <- fmt.Sprintf("ip: %v,content: %v", ip, content)
		src := fmt.Sprintf("/tmp/this-ip.env.%v", ip)
		fn := new(file)
		fn.set(src, content)
		fn.new()
		dest := fmt.Sprintf("root@%v:%v", ip, path)
		scp(src, dest)
	}
	/**
	n := len(this.getOthers())
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, ip := range this.getOthers() {
		go func(ip string, path string, tmp string, f *file) {
			defer wg.Done()
			content := tmp
			content = strings.Replace(content, "{{.this.ip}}", ip, -1)
			log2kubelog <- fmt.Sprintf("ip: %v,content: %v", ip, content)
			f.set(path, content)
			addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
			fmt.Printf("addr: %v\n", addr)
			cli := Client{Adress: addr}
			resBuf, err := cli.SendRequest("funcs.NewFile", *f)
			if err != nil {
				log.Fatalln(err)
			}
			ok := ResultIsOk(resBuf)
			fmt.Printf("rpc make ip env: %v\n", ok)
		}(ip, path, tmp, f)
	}
	wg.Wait()**/

	return nil
}

func (this *Core) makeEtcdEnvVariables() error {
	f := new(file)
	path := "/etc/kubernetes/env/etcd.env"
	var nodeips, etcdnodes, etcdendpoints string
	sep0 := ""
	sep1 := ""
	for _, ip := range this.kubeclusterinfo.getMaster() {
		name := fmt.Sprintf("etcd-%v", ip)

		nodeips += sep0
		nodeips += ip

		etcdnodes += sep1
		etcdnodes += fmt.Sprintf("%v=https://%v:2380", name, ip)

		etcdendpoints += sep1
		etcdendpoints += fmt.Sprintf("https://%v:2379", ip)

		sep0 = " "
		sep1 = ","
	}
	tmp := masteretcdenv
	tmp = strings.Replace(tmp, "{{.node.ips}}", nodeips, -1)
	tmp = strings.Replace(tmp, "{{.etcd.nodes}}", etcdnodes, -1)
	tmp = strings.Replace(tmp, "{{.etcd.endpoints}}", etcdendpoints, -1)
	node := nodeetcdenv
	node = strings.Replace(node, "{{.etcd.nodes}}", etcdnodes, -1)
	node = strings.Replace(node, "{{.etcd.endpoints}}", etcdendpoints, -1)
	// master
	// self
	content := tmp
	content = strings.Replace(content, "{{.node.name}}", fmt.Sprintf("etcd-%v", this.ip), -1)
	f.set(path, content)
	f.newFile()
	// others
	n := len(this.kubeclusterinfo.getMaster()) - 1
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, ip := range this.kubeclusterinfo.getMaster() {
		if ip != this.ip {
			go func(ip string, path string, content string, f *file) {
				defer wg.Done()
				content = strings.Replace(content, "{{.node.name}}", fmt.Sprintf("etcd-%v", ip), -1)
				f.set(path, content)
				addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
				fmt.Printf("addr: %v\n", addr)
				cli := Client{Adress: addr}
				resBuf, err := cli.SendRequest("funcs.NewFile", *f)
				if err != nil {
					log.Fatalln(err)
				}
				ok := ResultIsOk(resBuf)
				fmt.Printf("rpc make etcd env for master: %v\n", ok)
			}(ip, path, tmp, f)
		}
	}
	wg.Wait()
	// node
	if len(this.kubeclusterinfo.getNode()) == 0 {
		return nil
	}
	f.set(path, node)
	n = len(this.kubeclusterinfo.getNode())
	runtime.GOMAXPROCS(n)
	var wg1 sync.WaitGroup
	wg1.Add(n)
	for _, ip := range this.kubeclusterinfo.getNode() {
		go func(ip string, f *file) {
			defer wg1.Done()
			addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
			fmt.Printf("addr: %v\n", addr)
			cli := Client{Adress: addr}
			resBuf, err := cli.SendRequest("funcs.NewFile", *f)
			if err != nil {
				log.Fatalln(err)
			}
			ok := ResultIsOk(resBuf)
			fmt.Printf("rpc make etcd env for node: %v\n", ok)
		}(ip, f)
	}
	wg1.Wait()

	return nil
}

func (this *Core) makeEnvConf() error {
	return this.clusterRun(string(makeEnvConf), "")
}

func (this *Core) sourceEnv() error {
	f := new(file)
	f.set("/etc/profile", string(sourceEnv))
	// self
	f.try2append()
	// others
	n := len(this.getOthers())
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, ip := range this.getOthers() {
		go func(ip string, f *file) {
			defer wg.Done()
			addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
			fmt.Printf("addr: %v\n", addr)
			cli := Client{Adress: addr}
			resBuf, err := cli.SendRequest("funcs.Try2Append", *f)
			if err != nil {
				log.Fatalln(err)
			}
			ok := ResultIsOk(resBuf)
			fmt.Printf("rpc source env: %v\n", ok)
		}(ip, f)
	}
	wg.Wait()

	return nil
}
