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

func (this *Core) downloadKubernetes() error {
	//repo := fmt.Sprintf("https://gitlab.com/humstarman/kube-bin/raw/%v", string(kubeVersion))
	components := []string{"kubectl", "kube-apiserver", "kube-controller-manager", "kube-scheduler", "kubelet", "kube-proxy"}
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
	pack := "kubernetes-server-linux-amd64.tar.gz"
	path := fmt.Sprintf("/tmp/%v", pack)
	if !checkFileExist(path) {
		//0 get from other
		if !this.getFromOther(path) {
			// 1 download
			repo := fmt.Sprintf("https://dl.k8s.io/%v/%v", string(kubeVersion), pack)
			c := fmt.Sprintf("wget -c %v -O %v", repo, path)
			execCmd(c)
		}
	}
	cmd := fmt.Sprintf("cd /tmp && tar -zxf %v", pack)
	execCmd(cmd)
	path = "/tmp/kubernetes/server/bin"
	dest := "/tmp"
	for _, c := range components {
		src := fmt.Sprintf("%v/%v", path, c)
		j := fmt.Sprintf("yes | cp %v %v", src, dest)
		execCmd(j)
	}
	return nil
}

func (this *Core) makeKubeComponentsAvailable() error {
	components := []string{"kubectl", "kube-apiserver", "kube-controller-manager", "kube-scheduler", "kubelet", "kube-proxy"}
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

func (this *Core) distributeKubeComponents() error {
	components := []string{"kubectl", "kube-apiserver", "kube-controller-manager", "kube-scheduler", "kubelet", "kube-proxy"}
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

func (this *Core) makeKubectlPerm() error {
	f := new(file)
	// 0 generate csr
	path := "/etc/kubernetes/ssl/admin-csr.json"
	content := string(adminCsrJson)
	f.set(path, content)
	f.newFile()
	// 1 generate cert
	cmd := new(linuxcmd)
	str := "cd /etc/kubernetes/ssl && cfssl gencert -ca=/etc/kubernetes/ssl/ca.pem -ca-key=/etc/kubernetes/ssl/ca-key.pem -config=/etc/kubernetes/ssl/ca-config.json -profile=kubernetes admin-csr.json | cfssljson -bare admin"
	cmd.set(str, "")
	cmd.runLogged()
	return nil
}

func (this *Core) distributeKubectlPerm() error {
	components := []string{"admin-csr.json", "admin.csr", "admin-key.pem", "admin.pem"}
	path := "/etc/kubernetes/ssl"
	for _, c := range components {
		f := fmt.Sprintf("%v/%v", path, c)
		err := this.distribute(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func getBootstrapToekn() string {
	fn, err := os.OpenFile("/etc/kubernetes/token.csv", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fn.Close()
	b, err := ioutil.ReadAll(fn)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(strings.Trim(string(b), "\n"), ",")[0]
	//return bootstrapToken
}

func (this *Core) makeKubectlKubeconfig() error {
	cmd := new(linuxcmd)
	str := "kubectl config"
	var arg string
	// export KUBE_APISERVER=https://{{.vip}}:{{.port}}
	kubeApiServer := fmt.Sprintf("https://%v:%v", this.kubeclusterinfo.getVIP(), kubeMasterPort)
	// 0
	arg = fmt.Sprintf("set-cluster kubernetes --certificate-authority=/etc/kubernetes/ssl/ca.pem --embed-certs=true --server=%v --kubeconfig=/tmp/kubectl.kubeconfig", kubeApiServer)
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.runLogged()
	// 1
	arg = fmt.Sprintf("set-credentials admin --client-certificate=/etc/kubernetes/ssl/admin.pem --embed-certs=true --client-key=/etc/kubernetes/ssl/admin-key.pem --token=%v --kubeconfig=/tmp/kubectl.kubeconfig", getBootstrapToekn())
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.runLogged()
	// 2
	arg = "set-context kubernetes --cluster=kubernetes --user=admin --kubeconfig=/tmp/kubectl.kubeconfig"
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.runLogged()
	// 3
	arg = "use-context kubernetes --kubeconfig=/tmp/kubectl.kubeconfig"
	cmd.set(str, arg)
	fmt.Println(cmd.getCmd())
	cmd.runLogged()
	// 4
	src := "/root/.kube/config"
	dest := "/etc/kubernetes/admin.conf"
	str = fmt.Sprintf("yes | cp /tmp/kubectl.kubeconfig %v", src)
	cmd.set(str, "")
	fmt.Println(cmd.getCmd())
	cmd.runLogged()
	str = fmt.Sprintf("yes | cp %v %v", src, dest)
	cmd.set(str, "")
	fmt.Println(cmd.getCmd())
	cmd.runLogged()
	return nil
}

func (this *Core) distributeKubectlKubeconfig() error {
	components := []string{"/root/.kube/config", "/etc/kubernetes/admin.conf"}
	for _, c := range components {
		err := this.distribute(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Core) makeKubectlEnv() error {
	f := new(file)
	f.set("/etc/profile", "source <(kubectl completion bash)")
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
