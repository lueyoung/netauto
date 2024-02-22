package core

const etcdSvc = `[Unit]
Description=Etcd Server
After=network.target
After=network-online.target
Wants=network-online.target
Documentation=https://github.com/coreos

[Service]
Type=notify
EnvironmentFile=-/etc/kubernetes/env/env.conf
WorkingDirectory=/var/lib/etcd/
ExecStart=/usr/local/bin/etcd \
  --data-dir=/var/lib/etcd \
  --wal-dir=/var/lib/etcd-wal \
  --name=${NODE_NAME} \
  --cert-file=/etc/etcd/ssl/etcd.pem \
  --key-file=/etc/etcd/ssl/etcd-key.pem \
  --trusted-ca-file=/etc/kubernetes/ssl/ca.pem \
  --peer-cert-file=/etc/etcd/ssl/etcd.pem \
  --peer-key-file=/etc/etcd/ssl/etcd-key.pem \
  --peer-trusted-ca-file=/etc/kubernetes/ssl/ca.pem \
  --peer-client-cert-auth \
  --client-cert-auth \
  --listen-peer-urls=https://${NODE_IP}:2380 \
  --initial-advertise-peer-urls=https://${NODE_IP}:2380 \
  --listen-client-urls=https://${NODE_IP}:2379,http://127.0.0.1:2379 \
  --advertise-client-urls=https://${NODE_IP}:2379 \
  --initial-cluster-token=etcd-cluster-{{.n}} \
  --initial-cluster=${ETCD_NODES} \
  --initial-cluster-state=new \
  --auto-compaction-mode=periodic \
  --auto-compaction-retention=1 \
  --max-request-bytes=33554432 \
  --quota-backend-bytes=6442450944 \
  --heartbeat-interval=250 \
  --election-timeout=2000
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
`

const vipModeSvc = `[Unit]
Description=Switch-on Kernel Modules Needed by IPVS 

[Service]
Type=oneshot
ExecStart={{.bin}}

[Install]
WantedBy=multi-user.target
`

const kubeApiserverSvc = `[Unit]
Description=Kubernetes API Server
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=network.target

[Service]
EnvironmentFile=-/etc/kubernetes/env/env.conf
ExecStart=/usr/local/bin/kube-apiserver \
  --anonymous-auth=false \
  --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota \
  --advertise-address=${NODE_IP} \
  --bind-address=0.0.0.0 \
  --insecure-bind-address=0.0.0.0 \
  --authorization-mode=Node,RBAC \
  --runtime-config=api/all=true \
  --kubelet-https=true \
  --enable-bootstrap-token-auth \
  --token-auth-file=/etc/kubernetes/token.csv \
  --service-cluster-ip-range=${SERVICE_CIDR} \
  --service-node-port-range=${NODE_PORT_RANGE} \
  --tls-cert-file=/etc/kubernetes/ssl/kubernetes.pem \
  --tls-private-key-file=/etc/kubernetes/ssl/kubernetes-key.pem \
  --client-ca-file=/etc/kubernetes/ssl/ca.pem \
  --service-account-key-file=/etc/kubernetes/ssl/ca-key.pem \
  --etcd-cafile=/etc/kubernetes/ssl/ca.pem \
  --etcd-certfile=/etc/kubernetes/ssl/kubernetes.pem \
  --etcd-keyfile=/etc/kubernetes/ssl/kubernetes-key.pem \
  --etcd-servers=${ETCD_ENDPOINTS} \
  --enable-swagger-ui=true \
  --allow-privileged=true \
  --apiserver-count=3 \
  --audit-log-maxage=30 \
  --audit-log-maxbackup=3 \
  --audit-log-maxsize=100 \
  --audit-log-path=/var/lib/audit.log \
  --event-ttl=1h \
  --logtostderr=true \
  --v=2
Restart=on-failure
RestartSec=5
Type=notify
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
`

const kubeControllerManagerSvc = `[Unit]
Description=Kubernetes Controller Manager
Documentation=https://github.com/GoogleCloudPlatform/kubernetes

[Service]
EnvironmentFile=-/etc/kubernetes/env/env.conf
ExecStart=/usr/local/bin/kube-controller-manager \
  --use-service-account-credentials=true \
  --controllers=*,bootstrapsigner,tokencleaner \
  --address=127.0.0.1 \
  --master=http://${NODE_IP}:8080 \
  --allocate-node-cidrs=true \
  --service-cluster-ip-range=${SERVICE_CIDR} \
  --cluster-cidr=${CLUSTER_CIDR} \
  --cluster-name=kubernetes \
  --cluster-signing-cert-file=/etc/kubernetes/ssl/ca.pem \
  --cluster-signing-key-file=/etc/kubernetes/ssl/ca-key.pem \
  --service-account-private-key-file=/etc/kubernetes/ssl/ca-key.pem \
  --root-ca-file=/etc/kubernetes/ssl/ca.pem \
  --leader-elect=true \
  --v=2
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

const kubeSchedulerSvc = `[Unit]
Description=Kubernetes Scheduler
Documentation=https://github.com/GoogleCloudPlatform/kubernetes

[Service]
EnvironmentFile=-/etc/kubernetes/env/env.conf
ExecStart=/usr/local/bin/kube-scheduler \
  --address=127.0.0.1 \
  --master=http://${NODE_IP}:8080 \
  --leader-elect=true \
  --v=2
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

const dockerSvc = `[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target firewalld.service
Wants=network-online.target

[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
ExecStart=/usr/local/bin/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
# Uncomment TasksMax if your systemd version supports it.
# Only systemd 226 and above support this version.
#TasksMax=infinity
TimeoutStartSec=0
# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes
# kill only the docker process, not all processes in the cgroup
KillMode=process
# restart the docker process if it exits prematurely
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
`

const kubeletSvc = `[Unit]
Description=Kubernetes Kubelet
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=docker.service
Requires=docker.service

[Service]
EnvironmentFile=-/etc/kubernetes/env/env.conf
WorkingDirectory=/var/lib/kubelet
ExecStart=/usr/local/bin/kubelet \
  --network-plugin=cni \
  --cni-conf-dir=/etc/cni/net.d \
  --cni-bin-dir=/opt/cni/bin \
  --fail-swap-on=false \
  --pod-infra-container-image=lowyard/pause:latest \
  --cgroup-driver=cgroupfs \
  --address=${NODE_IP} \
  --hostname-override=${NODE_IP} \
  --bootstrap-kubeconfig=/etc/kubernetes/bootstrap.kubeconfig \
  --kubeconfig=/etc/kubernetes/kubelet.kubeconfig \
  --cert-dir=/etc/kubernetes/ssl \
  --cluster-dns=${CLUSTER_DNS_SVC_IP} \
  --cluster-domain=${CLUSTER_DNS_DOMAIN} \
  --hairpin-mode promiscuous-bridge \
  --allow-privileged=true \
  --serialize-image-pulls=false \
  --logtostderr=true \
  --pod-manifest-path=/etc/kubernetes/manifests \
  --feature-gates=RotateKubeletClientCertificate=true,RotateKubeletServerCertificate=true \
  --runtime-cgroups=/systemd/system.slice \
  --kubelet-cgroups=/systemd/system.slice \
  --cgroups-per-qos=false \
  --enforce-node-allocatable= \
  --v=2
ExecStartPost=/sbin/iptables -A INPUT -s 10.0.0.0/8 -p tcp --dport 4194 -j ACCEPT
ExecStartPost=/sbin/iptables -A INPUT -s 172.17.0.0/12 -p tcp --dport 4194 -j ACCEPT
ExecStartPost=/sbin/iptables -A INPUT -s 192.168.1.0/16 -p tcp --dport 4194 -j ACCEPT
ExecStartPost=/sbin/iptables -A INPUT -p tcp --dport 4194 -j DROP
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

const kubeProxySvc = `[Unit]
Description=Kubernetes Kube-Proxy Server
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=network.target

[Service]
EnvironmentFile=-/etc/kubernetes/env/env.conf
WorkingDirectory=/var/lib/kube-proxy
ExecStart=/usr/local/bin/kube-proxy \
  --bind-address=${NODE_IP} \
  --hostname-override=${NODE_IP} \
  --cluster-cidr=${SERVICE_CIDR} \
  --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig \
  --logtostderr=true \
  --proxy-mode=ipvs \
  --ipvs-min-sync-period=5s \
  --ipvs-sync-period=5s \
  --ipvs-scheduler=rr \
  --masquerade-all \
  --v=2
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
`

const projectSvc = `[Unit]
Description={{.name}} Process

[Service]
PIDFile={{.pid}}
ExecStart={{.cmd}}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`
