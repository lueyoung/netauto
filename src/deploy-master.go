package core

import (
	//"errors"
	"bufio"
	"fmt"
	"io"
	//"io/ioutil"
	"log"
	"os"
	//"runtime"
	"strings"
	//"sync"
)

func (this *Core) makeKubernetesPerm() error {
	// 0 make csr
	path := "/etc/kubernetes/ssl/kubernetes-csr.json"
	content := string(kubernetesCsrJson_P0)
	content += fmt.Sprintf("    \"%v\",\n", this.getKubeVIP())
	for _, ip := range this.getKubeMaster() {
		content += fmt.Sprintf("    \"%v\",\n", ip)
	}
	content += string(kubernetesCsrJson_P1)
	f := new(file)
	f.set(path, content)
	f.newFile()
	// 1 generate pem
	cmd := new(linuxcmd)
	str := "cd /etc/kubernetes/ssl && cfssl gencert -ca=/etc/kubernetes/ssl/ca.pem -ca-key=/etc/kubernetes/ssl/ca-key.pem -config=/etc/kubernetes/ssl/ca-config.json -profile=kubernetes kubernetes-csr.json | cfssljson -bare kubernetes"
	cmd.set(str, "")
	cmd.runLogged()
	// 1 distribute
	components := []string{"kubernetes-csr.json", "kubernetes.csr", "kubernetes-key.pem", "kubernetes.pem"}
	path = "/etc/kubernetes/ssl"
	for _, c := range components {
		f := fmt.Sprintf("%v/%v", path, c)
		err := this.distribute(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func getEncryptionKey() string {
	target := "ENCRYPTION_KEY"
	fn, err := os.OpenFile("/etc/kubernetes/env/env.conf", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fn.Close()
	rd := bufio.NewReader(fn)
	var line, ret string
	for {
		line, err = rd.ReadString('\n')

		if err != nil || io.EOF == err {
			break
		}
		if strings.Contains(line, target) {
			ret = line
			break
		} else {
			fmt.Println(line)
		}
	}
	ret = strings.Trim(ret, "\n")
	ret = strings.Trim(ret, "\"")
	ret = strings.Split(ret, "=\"")[1]
	return ret
}

func (this *Core) prepare4master() error {
	return nil
}

func (this *Core) makeAndDistributeKubeApiserverSystemdUnit() error {
	path := "/etc/systemd/system/kube-apiserver.service"
	f := new(file)
	f.set(path, string(kubeApiserverSvc))
	f.newFile()
	err := this.distribute2master(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) startKubeApiserverSvc() error {
	err := this.clusterRun("systemctl stop kube-apiserver", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload && systemctl enable kube-apiserver && systemctl restart kube-apiserver", "")
	if err != nil {
		return err
	}
	return nil
}

func authorizeKube2Kubelet() error {
	cmd := new(linuxcmd)
	c := fmt.Sprintf("%v/kubectl create clusterrolebinding", string(bin))
	a := "kube-apiserver:kubelet-apis --clusterrole=system:kubelet-api-admin --user kubernetes"
	cmd.set(c, a)
	cmd.runLogged()
	return nil
}

func (this *Core) makeAndDistributeKubeControllerManagerSystemdUnit() error {
	path := "/etc/systemd/system/kube-controller-manager.service"
	f := new(file)
	f.set(path, string(kubeControllerManagerSvc))
	f.newFile()
	err := this.distribute2master(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) startKubeControllerManagerSvc() error {
	err := this.clusterRun("systemctl stop kube-controller-manager", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload && systemctl enable kube-controller-manager && systemctl restart kube-controller-manager", "")
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) makeAndDistributeKubeSchedulerSystemdUnit() error {
	path := "/etc/systemd/system/kube-scheduler.service"
	f := new(file)
	f.set(path, string(kubeSchedulerSvc))
	f.newFile()
	err := this.distribute2master(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) startKubeSchedulerSvc() error {
	err := this.clusterRun("systemctl stop kube-scheduler", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload", "")
	if err != nil {
		return err
	}
	err = this.kubeMasterRun("systemctl daemon-reload && systemctl enable kube-scheduler && systemctl restart kube-scheduler", "")
	if err != nil {
		return err
	}
	return nil
}
