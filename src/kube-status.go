package core

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type kubeclusterstatus struct {
	Summary       string                       `json:"summary"`
	Status        map[string]kubenodestatus    `json:"status"`
	FailureReport map[string]map[string]string `json:"failure-report,omitempty"`
}

func (k *kubeclusterstatus) getFailed() map[string]map[string]string {
	return k.FailureReport
}

func (c *kubeclusterstatus) mkFailureReport() {
	report := make(map[string]map[string]string)
	for ip, s := range c.Status {
		if len(s.Failed) > 0 {
			report[ip] = s.Failed
		}
	}
	c.FailureReport = report
}

type kubenodestatus struct {
	Role                  string            `json:"role"`
	Etcd                  string            `json:"etcd,omitempty"`
	KubeApiserver         string            `json:"kube-apiserver,omitempty"`
	KubeControllerManager string            `json:"kube-controller-manager,omitempty"`
	KubeScheduler         string            `json:"kube-scheduler,omitempty"`
	Kubelet               string            `json:"kubelet,omitempty"`
	KubeProxy             string            `json:"kube-proxy,omitempty"`
	Docker                string            `json:"docker,omitempty"`
	Failed                map[string]string `json:"failed,omitempty"`
}

func (k *kubenodestatus) fine() bool {
	r := reflect.ValueOf(k).Elem()
	if k.Role == "master" {
		masters := []string{"Etcd", "KubeApiserver", "KubeControllerManager", "KubeScheduler"}
		for _, j := range masters {
			if v := r.FieldByName(j).String(); v != "active" {
				return false
			}
		}

	}
	nodes := []string{"Kubelet", "KubeProxy", "Docker"}
	for _, j := range nodes {
		if v := r.FieldByName(j).String(); v != "active" {
			return false
		}
	}
	return true
}

func getKubeNodeStatus() kubenodestatus {
	var masters [4]bool
	ret := new(kubenodestatus)
	r := reflect.ValueOf(ret).Elem()
	failed := make(map[string]string)
	components := make(map[string]string)
	components["Etcd"] = "etcd"
	components["KubeApiserver"] = "kube-apiserver"
	components["KubeControllerManager"] = "kube-controller-manager"
	components["KubeScheduler"] = "kube-scheduler"
	components["Kubelet"] = "kubelet"
	components["KubeProxy"] = "kube-proxy"
	components["Docker"] = "docker"
	masterCount := 0
	nullCount := 0
	for f, u := range components {
		cmd := fmt.Sprintf("systemctl is-active %v", u)
		v, _ := execCmdVerbose(cmd)
		v = strings.Trim(v, "\n")
		if v != "" && v != "unknown" {
			r.FieldByName(f).Set(reflect.ValueOf(v))
			if (f == "Etcd" || f == "KubeApiserver" || f == "KubeControllerManager" || f == "KubeScheduler") && v == "active" {
				masters[masterCount] = true
				masterCount++
			}
		}
		if v != "active" && v != "unknown" {
			failed[u] = v
		}
		if v != "active" {
			nullCount++
		}
	}
	master := masters[0]
	for _, m := range masters[1:] {
		master = master || m
		//master = master && m
	}
	if master {
		ret.Role = "master"
	} else if nullCount >= len(components)-2 {
		ret.Role = "null"
	} else {
		ret.Role = "node"
	}
	ret.Failed = failed
	return *ret
}

func (f *funcs) GetKubeNodeStatus(i args) kubenodestatus {
	return getKubeNodeStatus()
}

func kubeClusterFine(nodes map[string]kubenodestatus) bool {
	for _, k := range nodes {
		if !k.fine() {
			return false
		}
	}
	return true
}

func existKubeCluster(nodes map[string]kubenodestatus) bool {
	for _, k := range nodes {
		if existKubeOnNode(k) {
			return true
		}
	}
	return false
}

func existKubeOnNode(k kubenodestatus) bool {
	r := reflect.ValueOf(&k).Elem()
	components := []string{"KubeApiserver", "KubeControllerManager", "KubeScheduler", "Kubelet", "KubeProxy"}
	for _, c := range components {
		if r.FieldByName(c).String() == "active" {
			return true
		}
	}
	return false
}

func (this *Core) reportKubeStatus() (kubeclusterstatus, bool, bool) {
	ret := new(kubeclusterstatus)
	// get node status
	m := make(map[string]kubenodestatus)
	// self
	k := getKubeNodeStatus()
	m[this.ip] = k
	// others
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("funcs.GetKubeNodeStatus", args{})
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
		res := new(kubenodestatus)
		ConvertToNormalType(resBuf, res)
		m[ip] = *res
	}

	fine := kubeClusterFine(m)
	var exist bool
	if fine {
		ret.Summary = "fine"
		exist = true
	} else {
		ret.Summary = "not fine"
		exist = existKubeCluster(m)
	}
	ret.Status = m
	ret.mkFailureReport()
	return *ret, fine, exist
}

func (this *Core) fixKube() {
	status, fine, exist := this.reportKubeStatus()
	if fine {
		return
	}
	if !exist {
		return
	}
	failed := status.getFailed()
	for ip, kv := range failed {
		if len(kv) > 0 {
			if ip == this.ip {
				// self
				for k, _ := range kv {
					cmd := fmt.Sprintf("systemctl restart %v", k)
					execCmd(cmd)
				}
			} else {
				// other, rpc
				c := new(linuxcmd)
				addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
				fmt.Printf("addr: %v\n", addr)
				cli := Client{Adress: addr}
				for k, _ := range kv {
					c.set("systemctl restart", k)
					resBuf, err := cli.SendRequest("funcs.Run", *c)
					if err != nil {
						log.Fatalln(err)
					}
					ok := ResultIsOk(resBuf)
					fmt.Printf("restart kube componet: %v, on node: %v, result: %v\n", k, ip, ok)
				}
			}
		}
	}
}
