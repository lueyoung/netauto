package core

import (
	//"errors"
	"fmt"
	//"io/ioutil"
	//"log"
	//"os"
	//"runtime"
	//"strings"
	//"sync"
)

func (this *Core) downloadDocker() error {
	tgz := fmt.Sprintf("/tmp/docker-%v.tgz", string(dockerVersion))
	//path := "/tmp/docker"
	if !checkFileExist(tgz) {
		//0 get from other
		if !this.getFromOther(tgz) {
			// 1 download
			durl := fmt.Sprintf("https://download.docker.com/linux/static/stable/x86_64/docker-%v.tgz", string(dockerVersion))
			d := new(downlink)
			d.set(durl, tgz)
			d.download()
		}
	}
	c := new(linuxcmd)
	c.set(fmt.Sprintf("tar -zxvf %v -C /tmp", tgz), "")
	c.run()
	return nil
}

func (this *Core) makeDockerAvailable() error {
	components := []string{"docker", "docker-containerd", "docker-containerd-ctr", "docker-containerd-shim", "dockerd", "docker-init", "docker-proxy", "docker-runc"}
	for _, c := range components {
		src := fmt.Sprintf("/tmp/docker/%v", c)
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

func (this *Core) distributeDocker() error {
	components := []string{"docker", "docker-containerd", "docker-containerd-ctr", "docker-containerd-shim", "dockerd", "docker-init", "docker-proxy", "docker-runc"}
	path := string(bin)
	for _, c := range components {
		f := fmt.Sprintf("%v/%v", path, c)
		err := this.distribute(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Core) configDocker() error {
	return this.clusterRun(string(configDocker), "")
}

func (this *Core) makeDockerConfigAvailable() error {
	err := this.clusterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl daemon-reload && systemctl enable docker-iptables", "")
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) makeAndDistributeDockerSystemdUnit() error {
	path := "/etc/systemd/system/docker.service"
	f := new(file)
	f.set(path, string(dockerSvc))
	f.newFile()
	err := this.distribute(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) startDockerSvc() error {
	err := this.clusterRun("systemctl stop docker", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl daemon-reload && systemctl enable docker && systemctl restart docker", "")
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) makeBootstrapCrb() error {
	cmd := new(linuxcmd)
	c := fmt.Sprintf("%v/kubectl create clusterrolebinding", string(bin))
	a := "kubelet-bootstrap --clusterrole=system:node-bootstrapper --user=kubelet-bootstrap --group=system:bootstrappers"
	cmd.set(c, a)
	cmd.runLogged()
	return nil
}

func (this *Core) makeKubeletKubeconfig() error {
	cmd := new(linuxcmd)
	str := "kubectl config"
	var arg string
	// export KUBE_APISERVER=https://{{.vip}}:{{.port}}
	kubeApiServer := fmt.Sprintf("https://%v:%v", this.kubeclusterinfo.getVIP(), kubeMasterPort)
	// 0
	arg = fmt.Sprintf("set-cluster kubernetes --certificate-authority=/etc/kubernetes/ssl/ca.pem --embed-certs=true --server=%v --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig", kubeApiServer)
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	// 1
	arg = fmt.Sprintf("set-credentials kubelet-bootsrap --token=%v --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig", getBootstrapToekn())
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	// 2
	arg = "set-context default --cluster=kubernetes --user=kubelet-bootsrap --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig"
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	// 3
	arg = "use-context default --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig"
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	return nil
}

func (this *Core) distributeKubeletKubeconfig() error {
	path := "/etc/kubernetes/bootstrap.kubeconfig"
	err := this.distribute(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) makeAndDistributeKubeletSystemdUnit() error {
	path := "/etc/systemd/system/kubelet.service"
	f := new(file)
	f.set(path, string(kubeletSvc))
	f.newFile()
	err := this.distribute(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) startKubeletSvc() error {
	err := this.clusterRun("systemctl stop kubelet", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("[[ -f /etc/kubernetes/kubelet.kubeconfig ]] && rm -f /etc/kubernetes/kubelet.kubeconfig", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("rm -rf /var/lib/kubelet/*", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl daemon-reload && systemctl enable kubelet && systemctl restart kubelet", "")
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) configIPVS() error {
	return this.clusterRun(string(configIPVS), "")
}

func (this *Core) makeProxyPerm() error {
	// 0 make csr
	f := new(file)
	f.set("/etc/kubernetes/ssl/kube-proxy-csr.json", string(kubeProxyCsrJson))
	f.newFile()
	// 1 generate cert
	cmd := new(linuxcmd)
	str := "cd /etc/kubernetes/ssl && cfssl gencert -ca=/etc/kubernetes/ssl/ca.pem -ca-key=/etc/kubernetes/ssl/ca-key.pem -config=/etc/kubernetes/ssl/ca-config.json -profile=kubernetes kube-proxy-csr.json | cfssljson -bare kube-proxy"
	cmd.set(str, "")
	cmd.run()
	return nil
}

func (this *Core) distributeProxyPerm() error {
	components := []string{"kube-proxy-csr.json", "kube-proxy.csr", "kube-proxy-key.pem", "kube-proxy.pem"}
	path := "/etc/kubernetes/ssl"
	for _, c := range components {
		err := this.distribute(fmt.Sprintf("%v/%v", path, c))
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Core) makeProxyKubeconfig() error {
	cmd := new(linuxcmd)
	str := "kubectl config"
	var arg string
	// export KUBE_APISERVER=https://{{.vip}}:{{.port}}
	kubeApiServer := fmt.Sprintf("https://%v:%v", this.kubeclusterinfo.getVIP(), kubeMasterPort)
	// 0
	arg = fmt.Sprintf("set-cluster kubernetes --certificate-authority=/etc/kubernetes/ssl/ca.pem --embed-certs=true --server=%v --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig", kubeApiServer)
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	// 1
	arg = "set-credentials kube-proxy --client-certificate=/etc/kubernetes/ssl/kube-proxy.pem --client-key=/etc/kubernetes/ssl/kube-proxy-key.pem --embed-certs=true --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig"
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	// 2
	arg = "set-context default --cluster=kubernetes --user=kube-proxy --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig"
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	// 3
	arg = "use-context default --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig"
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.run()
	return nil
}

func (this *Core) distributeProxyKubeconfig() error {
	path := "/etc/kubernetes/kube-proxy.kubeconfig"
	err := this.distribute(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) makeAndDistributeProxySystemdUnit() error {
	path := "/etc/systemd/system/kube-proxy.service"
	content := string(kubeProxySvc)
	return this.makeAndDistributeSystemdUnit(path, content)
}

func (this *Core) startProxySvc() error {
	err := this.clusterRun("systemctl stop kube-proxy", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("rm -rf /var/lib/kube-proxy/*", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.clusterRun("systemctl daemon-reload && systemctl enable kube-proxy && systemctl restart kube-proxy", "")
	if err != nil {
		return err
	}
	return nil
}
