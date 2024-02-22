package core

import (
	//"encoding/json"
	//"errors"
	"fmt"
	//"github.com/hashicorp/raft"
	//"io"
	//"io/ioutil"
	"log"
	//"net/http"
	//"reflect"
	//"runtime"
	//"strings"
	//"sync"
	"time"
)

func (this *Core) installKube() {
	for {
		select {
		case msg := <-installkube:
			go this.listenKubelogCh()
			start := time.Now()
			fmt.Println(msg)
			this.kubelog.Println("recv signal, start to install Kubernetes ...")
			// 0 prepare
			this.kubelog.Println("stage 0: prepare for installation")
			// 0.0 stop existing system
			this.kubelog.Println("stage 0.0: stop existing Kubernetes system")
			err := this.clearKubeSystem()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.1 make ssh connection
			this.kubelog.Println("stage 0.1: make ssh connection")
			err = this.makeSshConnection()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.2 install useful components
			this.kubelog.Println("stage 0.2: install useful components")
			// 0.2.1 install useful components
			this.kubelog.Println("stage 0.2.1: update and install, and may take a long time ...")
			err = this.installUseful()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.2.2 config time zone and sync
			this.kubelog.Println("stage 0.2.2: configure time zone and sync time amongst hosts")
			err = this.configTimeSync()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.3 shutdown harmful components
			this.kubelog.Println("stage 0.3: shutdown harmful components")
			err = this.shutdownHarmful()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.4 optimize kernel parameters
			this.kubelog.Println("stage 0.4: optimize kernel parameters")
			err = this.optKernel()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.5 batch mkdir
			this.kubelog.Println("stage 0.5: batch mkdir")
			err = this.batchMkdir()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.6 make masters and nodes
			this.kubelog.Println("stage 0.6: make masters and nodes for Kubernetes")
			// 0.6.1 generate cluster information of Kubernetes
			this.kubelog.Println("stage 0.6.1: generate cluster information for Kubernetes")
			err = this.makeKubeClusterInfo()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 0.6.2 choose Kubernetes roles
			this.kubelog.Println("stage 0.6.2: choose Kubernetes roles")
			err = this.makeKubeRole()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 1 make cluster enviroment variables
			this.kubelog.Println("stage 1: make cluster enviroment variables")
			// 1.1 /etc/kubernetes/env/k8s.env & /etc/kubernetes/token.csv
			this.kubelog.Println("stage 1.1: make basic enviroment variables")
			err = this.makeBaseEnvVariables()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 1.2 /etc/kubernetes/env/this-ip.env
			this.kubelog.Println("stage 1.2: make enviroment variables about IP")
			err = this.makeIpEnvVariables()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 1.3 /etc/kubernetes/env/etcd.env
			this.kubelog.Println("stage 1.3: make enviroment variables about ETCD")
			err = this.makeEtcdEnvVariables()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 1.4 /etc/kubernetes/env/env.conf
			this.kubelog.Println("stage 1.4: integrate enviroment variables")
			err = this.makeEnvConf()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 1.5 source env
			this.kubelog.Println("stage 1.5: source enviroment variables")
			err = this.sourceEnv()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 2 generate CA pem
			this.kubelog.Println("stage 2: generate CA perm")
			// 2.1 download cfssl
			this.kubelog.Println("stage 2.1: download CFSSL components")
			err = this.downloadCfssl()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 2.2 chmod & mv
			this.kubelog.Println(fmt.Sprintf("stage 2.2: chmod & move CFSSL components from /tmp to %v", string(bin)))
			err = this.makeCfsslAvailable()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 2.3 generate root CA
			this.kubelog.Println("stage 2.3: generate root CA")
			err = this.generateRootCA()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 2.4 distribute root CA
			this.kubelog.Println("stage 2.4: distribute root CA")
			err = this.distributeRootCA()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 3 deploy kubectl
			this.kubelog.Println("stage 3: deploy kubectl")
			// 3.1 prepare kubernetes
			this.kubelog.Println("stage 3.1: get Kubernetes components")
			err = this.downloadKubernetes()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 3.2 make kubernetes components avaiable
			this.kubelog.Println(fmt.Sprintf("stage 3.2: move Kubernetes components from /tmp to %v", string(bin)))
			err = this.makeKubeComponentsAvailable()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 3.3 distribute kubernetes components
			this.kubelog.Println("stage 3.3: distribute Kubernetes components")
			err = this.distributeKubeComponents()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 3.4 generate kubectl perm
			this.kubelog.Println("stage 3.4: generate perm for kubectl")
			err = this.makeKubectlPerm()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 3.5 distribute kubectl perm
			this.kubelog.Println("stage 3.5: distribute kubectl perm")
			err = this.distributeKubectlPerm()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 3.6 generate kubectl kubeconfig
			this.kubelog.Println("stage 3.6: generate kubeconfig for kubectl")
			err = this.makeKubectlKubeconfig()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 3.7 distribute kubectl kubeconfig
			this.kubelog.Println("stage 3.5: distribute kubectl kubeconfig")
			err = this.distributeKubectlKubeconfig()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 3.8 make env for kubectl
			this.kubelog.Println("stage 3.8: make enviroment variables about kubectl")
			err = this.makeKubectlEnv()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 4 deploy etcd
			this.kubelog.Println("stage 4: deploy etcd")
			// 4.1 download etcd
			this.kubelog.Println("stage 4.1: download etcd components")
			err = this.downloadEtcd()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 4.2 make etcd avaiable
			this.kubelog.Println(fmt.Sprintf("stage 4.2: move etcd components from /tmp to %v", string(bin)))
			err = this.makeEtcdAvailable()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 4.3 distribute etcd components
			this.kubelog.Println("stage 4.3: distribute etcd components")
			err = this.distributeEtcdComponents()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 4.4 generate etcd perm
			this.kubelog.Println("stage 4.4: generate perm for etcd")
			err = this.makeEtcdPerm()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 4.5 distribute etcd perm
			this.kubelog.Println("stage 4.5: distribute etcd perm")
			err = this.distributeEtcdPerm()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 4.6 distribute etcd systemd unit
			this.kubelog.Println("stage 4.6: make and distribute systemd unit for etcd")
			err = this.makeAndDistributeEtcdSystemdUnit()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 4.7 start etcd service
			this.kubelog.Println("stage 4.7: start etcd service")
			err = this.startEtcdSvc()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 5 deploy vip
			this.kubelog.Println("stage 5: deploy vip")
			// 5.1 install haproxy & keepalived
			this.kubelog.Println("stage 5.1: install Haproxy & Keepalived")
			err = this.installVIP()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 5.2 make oneshot systemd unit to optimize kernel for vip
			this.kubelog.Println("stage 5.2: make oneshot systemd unit to optimize kernel for VIP")
			err = this.makeVipMod()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 5.3 config harproxy
			this.kubelog.Println("stage 5.3: configure Haproxy")
			err = this.configHaproxy()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 5.4 config keepalived
			this.kubelog.Println("stage 5.4: configure Keepalived")
			err = this.configKeepalived()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 5.5 make script for keepalived to check haproxy
			this.kubelog.Println("stage 5.5: make script for Keepalived to check the status of Haproxy")
			err = this.keepalivedCheckHaproxy()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 5.6 start keepalived & haproxy
			this.kubelog.Println("stage 5.6: start Haproxy & Keepalived service")
			err = this.startVipComponents()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 6 deploy master components
			this.kubelog.Println("stage 6: deploy master components")
			// 6.1 make kubernetes perm.
			this.kubelog.Println("stage 6.1: make kubernetes perm")
			err = this.makeKubernetesPerm()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.2 prepare for master
			this.kubelog.Println("stage 6.2: make encryption-config, audit-policy and metrics-server for master")
			err = this.prepare4master()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.3 deploy kube-apiserver
			this.kubelog.Println("stage 6.3: deploy kube-apiserver component")
			// 6.3.1 distribute kube-apiserver systemd unit
			this.kubelog.Println("stage 6.3.1: make and distribute systemd unit for kube-apiserver")
			err = this.makeAndDistributeKubeApiserverSystemdUnit()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.3.2 start kube-apiserver service
			this.kubelog.Println("stage 6.3.2: start kube-apiserver service")
			err = this.startKubeApiserverSvc()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.3.3 authorize kubernetes perm to visit kubelet API
			this.kubelog.Println("stage 6.3.3: authorize kubernetes perm to visit kubelet API")
			err = authorizeKube2Kubelet()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.4 deploy kube-controller-manager
			this.kubelog.Println("stage 6.4: deploy kube-controller-manager component")
			// 6.4.1 distribute kube-controller-manager systemd unit
			this.kubelog.Println("stage 6.4.1: make and distribute systemd unit for kube-controller-manager")
			err = this.makeAndDistributeKubeControllerManagerSystemdUnit()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.4.2 start kube-controller-manager service
			this.kubelog.Println("stage 6.4.2: start kube-controller-manager service")
			err = this.startKubeControllerManagerSvc()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.5 deploy kube-scheduler
			this.kubelog.Println("stage 6.5: deploy kube-scheduler component")
			// 6.5.1 distribute kube-scheduler systemd unit
			this.kubelog.Println("stage 6.5.1: make and distribute systemd unit for kube-scheduler")
			err = this.makeAndDistributeKubeSchedulerSystemdUnit()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 6.5.2 start kube-scheduler service
			this.kubelog.Println("stage 6.5.2: start kube-scheduler service")
			err = this.startKubeSchedulerSvc()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 7 deploy node components
			this.kubelog.Println("stage 7: deploy node components")
			// 7.1 deploy docker
			this.kubelog.Println("stage 7.1: deploy docker")
			// 7.1.1 download docker
			this.kubelog.Println("stage 7.1.1: download docker")
			err = this.downloadDocker()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.1.2 move docker components from /tmp to PATH
			err = this.makeDockerAvailable()
			this.kubelog.Println(fmt.Sprintf("stage 7.1.2: move docker components from /tmp to %v", string(bin)))
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.1.3 distribute docker
			this.kubelog.Println("stage 7.1.3: distribute docker")
			err = this.distributeDocker()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.1.4 config docker
			this.kubelog.Println("stage 7.1.4: configure docker, and may take some time ...")
			err = this.configDocker()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.1.5 make the config available
			this.kubelog.Println("stage 7.1.5: make the configuration available")
			err = this.makeDockerConfigAvailable()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.1.6 distribute docker systemd unit
			this.kubelog.Println("stage 7.1.6: make and distribute systemd unit for docker")
			err = this.makeAndDistributeDockerSystemdUnit()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.1.7 start docker service
			this.kubelog.Println("stage 7.1.7: start docker service")
			err = this.startDockerSvc()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.2 deploy kubelet
			this.kubelog.Println("stage 7.2: deploy kubelet")
			// 7.2.1 generate clusterrolebinding for bootstrapper
			this.kubelog.Println("stage 7.2.1: generate clusterrolebinding for bootstrapper")
			err = this.makeBootstrapCrb()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.2.2 generate kubelet bootstrap kubeconfig
			this.kubelog.Println("stage 7.2.2: generate kubelet bootstrap kubeconfig")
			err = this.makeKubeletKubeconfig()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.2.3 distribute kubelet bootstrap kubeconfig
			this.kubelog.Println("stage 7.2.3: distribute kubelet bootstrap kubeconfig")
			err = this.distributeKubeletKubeconfig()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.2.4 make and distribute kubelet systemd unit
			this.kubelog.Println("stage 7.2.4: make and distribute kubelet systemd unit")
			err = this.makeAndDistributeKubeletSystemdUnit()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.2.5 start kubelet service
			this.kubelog.Println("stage 7.2.5: start kubelet service")
			err = this.startKubeletSvc()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.3 deploy kube-proxy
			this.kubelog.Println("stage 7.3: deploy kube-proxy")
			// 7.3.1 configure for ipvs
			this.kubelog.Println("stage 7.3.1: configure for ipvs")
			err = this.configIPVS()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.3.2 generate kube-proxy perm
			this.kubelog.Println("stage 7.3.2: generate kube-proxy perm")
			err = this.makeProxyPerm()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.3.3 distribute kube-proxy perm
			this.kubelog.Println("stage 7.3.3: distribute kube-proxy perm")
			err = this.distributeProxyPerm()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.3.4 generate kubelet bootstrap kubeconfig
			this.kubelog.Println("stage 7.3.4: generate kube-proxy kubeconfig")
			err = this.makeProxyKubeconfig()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.3.5 distribute kube-proxy kubeconfig
			this.kubelog.Println("stage 7.3.5: distribute kube-proxy kubeconfig")
			err = this.distributeProxyKubeconfig()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.3.6 make and distribute kube-proxy systemd unit
			this.kubelog.Println("stage 7.3.6: make and distribute kube-proxy systemd unit")
			err = this.makeAndDistributeProxySystemdUnit()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 7.3.7 start kube-proxy service
			this.kubelog.Println("stage 7.3.7: start kube-proxy service")
			err = this.startProxySvc()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 8 deploy cni
			this.kubelog.Println("stage 8: deploy calico cni")
			err = this.deployCalico()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 9 approve node certificates
			this.kubelog.Println("stage 9: approve node certificates")
			err = this.approveNodeCert()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			// 10 install addons
			this.kubelog.Println("stage 10: install addons")
			// 10.1 install kube dns
			this.kubelog.Println("stage 10.1: deploy Kubernetes DNS")
			err = this.deployKubeDNS()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}
			// 10.2 install kube dashboard
			this.kubelog.Println("stage 10.2: deploy Kubernetes dashboard")
			err = this.deployKubeDashboard()
			if err != nil {
				this.kubelog.Println(err)
				log.Fatal(err)
			}

			this.kubelog.Println("finished installation of Kubernetes")
			this.kubeSummary(start)
		}
	}
}
