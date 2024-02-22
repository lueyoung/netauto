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
)

func (this *Core) installVIP() error {
	return this.kubeMasterRun(string(installVIP), "")
}

func (this *Core) makeVipMod() error {
	name := "vip-mod"
	// 0 a binary file to opt kernel
	// 0.0 make
	f := new(file)
	path := fmt.Sprintf("%v/%v.sh", string(bin), name)
	f.set(path, string(vipMod))
	err := f.newExec()
	if err != nil {
		return err
	}
	// 0.1 distribute
	err = this.distribute(path)
	if err != nil {
		return err
	}
	// 1 oneshot systemd unit
	// 1.1 make
	content := string(vipModeSvc)
	content = strings.Replace(content, "{{.bin}}", fmt.Sprintf("%v/%v.sh", string(bin), name), -1)
	path = fmt.Sprintf("/etc/systemd/system/%v.service", name)
	f.set(path, content)
	f.newFile()
	// 1.2 distribute
	err = this.distribute(path)
	if err != nil {
		return err
	}
	// 1.3 start
	err = this.clusterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("systemctl daemon-reload && systemctl enable %v && systemctl restart %v", name, name)
	err = this.clusterRun(cmd, "")
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) configHaproxy() error {
	path := "/etc/haproxy/haproxy.cfg"
	// 0 make config file
	var content string
	content += string(haproxyCfg)
	content = strings.Replace(content, "{{.port}}", fmt.Sprintf("%v", kubeMasterPort), -1)
	sep := ""
	for i, ip := range this.getKubeMaster() {
		content += sep
		str := fmt.Sprintf("        server k8s-api-%v %v:%v check inter 10000 fall 2 rise 2 weight 1", i+1, ip, realKubeMasterPort)
		content += str
		sep = "\n"
	}
	content += "\n"
	f := new(file)
	f.set(path, content)
	f.newFile()
	// 1 distribute
	err := this.distribute2master(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) configKeepalived() error {
	src := "/tmp/keepalived.conf"
	dest := "/etc/keepalived/keepalived.conf"
	for i, ip := range this.getKubeMaster() {
		f := new(file)
		var content string
		content += string(keepalivedConf_P0)
		content = strings.Replace(content, "{{.ip}}", ip, -1)
		content = strings.Replace(content, "{{.router.id}}", string(routerID), -1)
		content += fmt.Sprintf("%v\n", ip)
		content += "    unicast_peer {\n"
		for _, j := range this.getKubeMaster() {
			if j != ip {
				content += fmt.Sprintf("        %v\n", j)
			}
		}
		content += "    }\n    # 初始化状态\n"
		if i == 0 {
			content += "    state MASTER\n"
		} else {
			content += "    state BACKUP\n"
		}
		content += string(keepalivedConf_P1)
		content = strings.Replace(content, "{{.virtual.router.id}}", fmt.Sprintf("%v", virtualRouterID), -1)
		if i == 0 {
			content += "    priority 101\n"
		} else {
			content += "    priority 99\n"
		}
		content += string(keepalivedConf_P2)
		content = strings.Replace(content, "{{.vip}}", this.getKubeVIP(), -1)
		if ip == this.ip {
			f.set(dest, content)
			f.newFile()
		} else {
			f.set(src, content)
			f.set(src, content)
			f.newFile()
			scp(src, fmt.Sprintf("root@%v:%v", ip, dest))
		}
	}
	err := this.kubeMasterRun(string(chInterface), "")
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) keepalivedCheckHaproxy() error {
	path := "/etc/keepalived/chk.sh"
	f := new(file)
	f.set(path, string(chkSh))
	f.newExec()
	err := this.distribute(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) startVipComponents() error {
	err := this.clusterRun("systemctl stop keepalived", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl stop haproxy", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload && systemctl enable haproxy && systemctl restart haproxy", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload && systemctl enable keepalived && systemctl restart keepalived", "")
	if err != nil {
		return err
	}
	return nil

}
