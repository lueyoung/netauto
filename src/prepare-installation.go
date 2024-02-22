package core

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

func checkFileExist(path string) bool {
	exist := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func (this *Core) clearKubeSystem() error {
	this.clearKubeNode()
	this.clearKubeMaster()
	this.clearKubeEtcd()
	return nil
}

func (this *Core) clearKubeNode() {
	this.clusterRun(string(clearKubeNode), "")
}

func (this *Core) clearKubeMaster() {
	this.clusterRun(string(clearKubeMaster), "")
}

func (this *Core) clearKubeEtcd() {
	this.clusterRun(string(clearKubeEtcd), "")
}

func (this *Core) makeLogFile() error {
	cmd := fmt.Sprintf("[[ -f %v ]] || touch %v", string(kubelog), string(kubelog))
	this.clusterRun(cmd, "")
	return nil
}

func (this *Core) makeSshConnection() error {
	// 0 make ssh key
	public := "/root/.ssh/id_rsa.pub"
	if !checkFileExist(public) {
		key := "/root/.ssh/id_rsa"
		cmd := fmt.Sprintf("ssh-keygen -t rsa -P \"\" -f %v", key)
		if _, err := execCmd(cmd); err != nil {
			log.Println(err)
			return err
		}
	}
	// 1 read from '/root/.ssh/id_rsa.pub'
	fn, err := os.OpenFile(public, os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	defer fn.Close()
	raw, err := ioutil.ReadAll(fn)
	if err != nil {
		log.Println(err)
		return err
	}
	content := strings.Replace(string(raw), "\n", "", 1)
	// 2 write to '/root/.ssh/authorized_keys' to all nodes
	authorized := "/root/.ssh/authorized_keys"
	f := new(file)
	f.set(authorized, content)
	//f.Path = authorized
	//f.Content = content
	//fmt.Println(*f)
	// 2.1 self
	tools := new(funcs)
	tools.Try2Append(*f)
	// 2.2 others
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
			fmt.Printf("rpc make ssh: %v\n", ok)
		}(ip, f)
	}
	wg.Wait()
	// 3 handle ECDSA key fingerprint
	// no need, using 'StrictHostKeyChecking no' parameter

	return nil
}

func (this *Core) installUseful() error {
	return this.clusterRun(string(installUseful), "")
}

func (this *Core) configTimeSync() error {
	f := new(file)
	path := "/etc/chrony.conf"
	// self
	nums := strings.Split(this.cidr, ".")
	cidr := fmt.Sprintf("%v.%v.0.0/16", nums[0], nums[1])
	content := fmt.Sprintf("%v", cidr)
	f.set(path, content)
	configNtpServer(*f)
	// others
	server := this.ip
	content = fmt.Sprintf("%v", server)
	f.set(path, content)
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
			resBuf, err := cli.SendRequest("funcs.ConfigNtpClient", *f)
			if err != nil {
				log.Fatalln(err)
			}
			ok := ResultIsOk(resBuf)
			fmt.Printf("rpc config ntp cli: %v\n", ok)
		}(ip, f)
	}
	wg.Wait()
	cmds := []string{"timedatectl set-timezone Asia/Shanghai", "timedatectl set-local-rtc 0", "systemctl restart rsyslog", "systemctl restart crond", "systemctl restart chronyd"}
	for _, cmd := range cmds {
		this.clusterRun(cmd, "")
	}

	return nil
}

func (this *Core) shutdownHarmful() error {
	return this.clusterRun(string(shutdownHarmful), "")
}

func (this *Core) optKernel() error {
	return this.clusterRun(string(optKernel), "")
}

func (this *Core) batchMkdir() error {
	dirs := []string{"/etc/kubernetes/env", "/etc/kubernetes/ssl", "/etc/flanneld/ssl", "/var/lib/kubelet", "/var/lib/kube-proxy", "/etc/kubernetes/manifests", "/etc/calico", "/etc/harbor/ssl", "/etc/etcd/ssl", "/var/lib/etcd", "/var/lib/etcd-wal", "/root/.kube", "/root/.ssh", "/etc/haproxy", "/etc/keepalived"}
	for _, dir := range dirs {
		this.clusterRun("mkdir -p", dir)
	}
	return nil
}

func (this *Core) makeKubeClusterInfo() error {
	// the node run the installation must be a master
	// if total numbers of nodes <= 3, all the nodes should be made masters
	// if total numbers of nodes > 3, besides self choose two with highest scores to be masters
	// vip used the first three num of self, the last one is 240 by default
	total := len(this.getAllMembers())
	this.kubeclusterinfo.Num = total
	if total <= 3 {
		this.kubeclusterinfo.Master = this.getAllMembers()
	} else {
		var master []string
		var node []string
		sorted := sortMapByValue(this.getScores())
		j := 0
		// self
		master = append(master, this.ip)
		// others
		for _, p := range sorted {
			if p.Key != this.ip {
				if j < 2 {
					master = append(master, p.Key)
				} else {
					// add others as node
					node = append(node, p.Key)
				}
				j++
			}
		}
		this.kubeclusterinfo.Master = master
		this.kubeclusterinfo.Node = node
	}
	// vip
	// should double check if another VIP already set in the network
	// a VIP cannot ping, so a used VIP seems to be available to the program
	vip := makeVIP(this.ip)
	if vip == "" {
		log.Fatal("VIP cannot be set")
	}
	this.kubeclusterinfo.VIP = vip
	// sync cluster info
	n := len(this.getOthers())
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, ip := range this.getOthers() {
		go func(ip string, k *kubeclusterinfo) {
			defer wg.Done()
			addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
			fmt.Printf("addr: %v\n", addr)
			cli := Client{Adress: addr}
			resBuf, err := cli.SendRequest("kubeclusterinfo.Copy", *k)
			if err != nil {
				log.Fatalln(err)
			}
			ok := ResultIsOk(resBuf)
			fmt.Printf("rpc make cluster info: %v\n", ok)
		}(ip, this.kubeclusterinfo)
	}
	wg.Wait()

	return nil
}

func (this *Core) makeKubeRole() error {
	// master
	n := len(this.kubeclusterinfo.getMaster())
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	input := new(args)
	input.V = "master"
	for _, ip := range this.kubeclusterinfo.getMaster() {
		go func(ip string, input *args) {
			defer wg.Done()
			addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
			fmt.Printf("addr: %v\n", addr)
			cli := Client{Adress: addr}
			resBuf, err := cli.SendRequest("kube.SetRole", *input)
			if err != nil {
				log.Fatalln(err)
			}
			ok := ResultIsOk(resBuf)
			fmt.Printf("rpc make kube role: %v\n", ok)
		}(ip, input)
	}
	wg.Wait()
	// node
	n = len(this.kubeclusterinfo.getNode())
	if n > 0 {
		runtime.GOMAXPROCS(n)
		var wg sync.WaitGroup
		wg.Add(n)
		input := new(args)
		input.V = "node"
		for _, ip := range this.kubeclusterinfo.getNode() {
			go func(ip string, input *args) {
				defer wg.Done()
				addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
				fmt.Printf("addr: %v\n", addr)
				cli := Client{Adress: addr}
				resBuf, err := cli.SendRequest("kube.SetRole", *input)
				if err != nil {
					log.Fatalln(err)
				}
				ok := ResultIsOk(resBuf)
				fmt.Printf("rpc make kube role: %v\n", ok)
			}(ip, input)
		}
		wg.Wait()
	}
	return nil
}
