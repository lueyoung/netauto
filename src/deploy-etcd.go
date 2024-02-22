package core

import (
	//"errors"
	"fmt"
	//"io/ioutil"
	//"log"
	//"os"
	//"runtime"
	"strings"
	//"sync"
	"math/rand"
	"time"
)

func (this *Core) downloadEtcd() error {
	components := []string{"etcdctl", "etcd"}
	n := len(components)
	i := 0
	for _, c := range components {
		p := fmt.Sprintf("/tmp/%v", c)
		if checkFileExist(p) {
			i++
		}
	}
	if i == n {
		return nil
	}
	pack := fmt.Sprintf("etcd-%v-linux-amd64.tar.gz", string(etcdVersion))
	path := fmt.Sprintf("/tmp/%v", pack)
	if !checkFileExist(path) {
		if !this.getFromOther(path) {
			repo := fmt.Sprintf("https://github.com/etcd-io/etcd/releases/download/%v/%v", string(etcdVersion), pack)
			c := fmt.Sprintf("wget -c %v -O %v", repo, path)
			execCmd(c)
		}
	}
	cmd := fmt.Sprintf("cd /tmp && tar -zxf %v", pack)
	execCmd(cmd)
	path = fmt.Sprintf("/tmp/etcd-%v-linux-amd64", string(etcdVersion))
	dest := "/tmp"
	for _, c := range components {
		src := fmt.Sprintf("%v/%v", path, c)
		j := fmt.Sprintf("yes | cp %v %v", src, dest)
		execCmd(j)
	}
	return nil
}

func (this *Core) makeEtcdAvailable() error {
	components := []string{"etcdctl", "etcd"}
	for _, c := range components {
		src := fmt.Sprintf("/tmp/%v", c)
		dest := fmt.Sprintf("%v/%v", string(bin), c)
		if !checkFileExist(dest) {
			cmd := new(linuxcmd)
			str := fmt.Sprintf("chmod a+x %v", src)
			cmd.set(str, "")
			cmd.run()
			str = fmt.Sprintf("yes | cp %v %v", src, dest)
			cmd.set(str, "")
			cmd.run()
		}
	}
	return nil
}

func (this *Core) distributeEtcdComponents() error {
	components := []string{"etcd", "etcdctl"}
	for _, c := range components {
		err := this.distribute2master(fmt.Sprintf("%v/%v", string(bin), c))
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Core) makeEtcdPerm() error {
	// 0 generate csr
	var csr string
	csr += string(etcdCsrJsonP0)
	sep := ""
	for _, ip := range this.kubeclusterinfo.getMaster() {
		csr += sep
		csr += fmt.Sprintf("    \"%v\"", ip)
		sep = ",\n"
	}
	csr += string(etcdCsrJsonP1)
	f := new(file)
	f.set("/etc/kubernetes/ssl/etcd-csr.json", csr)
	f.newFile()
	// 1 generate cert
	cmd := new(linuxcmd)
	str := "cd /etc/kubernetes/ssl && cfssl gencert -ca=/etc/kubernetes/ssl/ca.pem -ca-key=/etc/kubernetes/ssl/ca-key.pem -config=/etc/kubernetes/ssl/ca-config.json -profile=kubernetes etcd-csr.json | cfssljson -bare etcd"
	cmd.set(str, "")
	cmd.runLogged()
	return nil
}

func (this *Core) distributeEtcdPerm() error {
	components := []string{"etcd-csr.json", "etcd.csr", "etcd-key.pem", "etcd.pem"}
	src := "/etc/kubernetes/ssl"
	dest := "/etc/etcd/ssl"
	for _, c := range components {
		ssl0 := fmt.Sprintf("%v/%v", src, c)
		ssl1 := fmt.Sprintf("%v/%v", dest, c)
		cmd := new(linuxcmd)
		str := fmt.Sprintf("cp %v %v", ssl0, ssl1)
		cmd.set(str, "")
		cmd.run()
		err := this.distribute(ssl0)
		if err != nil {
			return err
		}
		err = this.distribute(ssl1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Core) makeAndDistributeEtcdSystemdUnit() error {
	path := "/etc/systemd/system/etcd.service"
	content := string(etcdSvc)
	n := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	content = strings.Replace(content, "{{.n}}", n, -1)
	f := new(file)
	f.set(path, content)
	f.newFile()
	err := this.distribute2master(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) startEtcdSvc() error {
	err := this.clusterRun("systemctl stop etcd", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("rm -rf /var/lib/etcd/*", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload && systemctl enable etcd && systemctl restart etcd", "")
	if err != nil {
		return err
	}
	return nil
	//return this.kubeMasterRun("rm -rf /var/lib/etcd/* && systemctl daemon-reload && systemctl enable etcd && systemctl restart etcd", "")
}
