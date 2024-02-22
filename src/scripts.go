package core

const installUseful = `if [ -x "$(command -v yum)" ]; then
  yum makecache fast
  yum install -y epel-release
  yum install -y conntrack ipvsadm ipset jq iptables curl sysstat libseccomp wget
  yum install -y chrony
elif [ -x "$(command -v apt-get)" ]; then
  apt-get update
  apt-get install -y conntrack ipvsadm ipset jq iptables curl sysstat libseccomp wget
  apt-get install -y chrony
fi
/usr/sbin/modprobe ip_vs
`

const shutdownHarmful = `if [ -x "$(command -v yum)" ]; then
  FIREWALL="firewalld"
elif [ -x "$(command -v apt-get)" ]; then
  FIREWALL="ufw"
else
  echo "$(date) - $0 - [ERROR] - unknown Distributor ID."
  exit 1
fi
# firewall
systemctl stop $FIREWALL
systemctl disable $FIREWALL
# swap
swapoff -a
sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab 
# SELinux
## temporary
if [ -x "$(command -v setenforce)" ]; then
  setenforce 0
fi
## for ever
if [ -f "/etc/selinux/config" ]; then
  sed -i s/"SELINUX=enforcing"/"SELINUX=disabled"/g /etc/selinux/config
fi
# Postfix
systemctl disable postfix
systemctl stop postfix
`

const optKernel = `cat > /tmp/kubernetes.conf <<EOF
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
net.ipv4.ip_forward=1
net.ipv4.tcp_tw_recycle=0
vm.swappiness=0 # 禁止使用 swap 空间，只有当系统 OOM 时才允许使用它
vm.overcommit_memory=1 # 不检查物理内存是否够用
vm.panic_on_oom=0 # 开启 OOM
fs.inotify.max_user_instances=8192
fs.inotify.max_user_watches=1048576
fs.file-max=52706963
fs.nr_open=52706963
net.ipv6.conf.all.disable_ipv6=1
net.netfilter.nf_conntrack_max=2310720
EOF
yes | cp /tmp/kubernetes.conf  /etc/sysctl.d/kubernetes.conf
sysctl -p /etc/sysctl.d/kubernetes.conf
`

const makeBaseEnvVariables = `BOOTSTRAP_TOKEN=$(head -c 16 /dev/urandom | od -An -t x | tr -d ' ')
ENCRYPTION_KEY=$(head -c 32 /dev/urandom | base64)
FILE=k8s.env
cat >/tmp/${FILE} <<EOF
BOOTSTRAP_TOKEN="${BOOTSTRAP_TOKEN}"
ENCRYPTION_KEY="ENCRYPTION_KEY"
SERVICE_CIDR="10.254.0.0/16"
CLUSTER_CIDR="{{.cluster.cidr}}"
NODE_PORT_RANGE="10000-32766"
FLANNEL_ETCD_PREFIX="/kubernetes/network"
CLUSTER_KUBERNETES_SVC_IP="10.254.0.1"
CLUSTER_DNS_SVC_IP="10.254.0.2"
CLUSTER_DNS_DOMAIN="cluster.local."
EOF
yes | cp /tmp/${FILE} /etc/kubernetes/env/${FILE}
cat >/tmp/token.csv <<EOF
${BOOTSTRAP_TOKEN},kubelet-bootstrap,10001,"system:kubelet-bootstrap"
EOF
yes | cp /tmp/token.csv /etc/kubernetes/token.csv
`

const makeEnvConf = `FILES=$(find /etc/kubernetes/env -type f -name "*.env")
DEST=/tmp/env.conf
[ -f $DEST ] && rm $DEST
[ -f $DEST ] || touch $DEST
if [ -n "$FILES" ]; then
  for FILE in $FILES; do
    cat $FILE >> $DEST
  done
fi
# renove 'export'
sed -i 's/export //g' $DEST
# del commit
sed -i '/^#/d' $DEST
# del blank
sed -i '/^$/d' $DEST
yes | mv $DEST /etc/kubernetes/env
`

const installVIP = `if [ -x "$(command -v yum)" ]; then
  yum makecache fast
  yum install -y haproxy keepalived
fi
if [ -x "$(command -v apt-get)" ]; then
  apt-get update
  apt-get install -y haproxy keepalived
fi
MODS=" \
net.ipv4.ip_forward^=^1 \
net.ipv4.ip_nonlocal_bind^=^1 \
net.ipv4.conf.lo.arp_ignore^=^1 \
net.ipv4.conf.lo.arp_announce^=^2 \
net.ipv4.conf.all.arp_ignore^=^1 \
net.ipv4.conf.all.arp_announce^=^2"
FILE=/etc/sysctl.d/vip.conf
[ -f $FILE ] && rm -f $FILE
[ -f $FILE ] || touch $FILE
for MOD in $MODS; do
  MOD=$(echo $MOD | tr "^" " ")
  if ! cat $FILE | grep "$MOD"; then
    echo $MOD >> $FILE
  fi
done
sysctl -p ${FILE}
`

const vipMod = `#!/bin/bash
ipvs_modules="ip_vs"
for kernel_module in ${ipvs_modules}; do
  /sbin/modprobe ${kernel_module}
done
lsmod | grep ip_vs
MODS=" \
net.ipv4.ip_forward^=^1 \
net.ipv4.ip_nonlocal_bind^=^1 \
net.ipv4.conf.lo.arp_ignore^=^1 \
net.ipv4.conf.lo.arp_announce^=^2 \
net.ipv4.conf.all.arp_ignore^=^1 \
net.ipv4.conf.all.arp_announce^=^2"
FILE=/etc/sysctl.d/vip.conf
[ -f $FILE ] && rm -f $FILE
[ -f $FILE ] || touch $FILE
for MOD in $MODS; do
  MOD=$(echo $MOD | tr "^" " ")
  if ! cat $FILE | grep "$MOD"; then
    echo $MOD >> $FILE
  fi
done
sysctl -p ${FILE}
`

const chInterface = `FILE=/etc/keepalived/keepalived.conf
IP=$(cat $FILE | grep __IP__) 
IP=${IP##*"__IP__ "}
INTERFACE=$(ip addr | grep $IP)
INTERFACE=${INTERFACE##*" "}
sed -i s/"{{.interface}}"/"${INTERFACE}"/g $FILE
`

const chkSh = `#!/bin/bash
flag=$(systemctl status haproxy &> /dev/null;echo $?)
if [[ $flag != 0 ]]; then
  echo "haproxy is down, close the keepalived"
  systemctl stop keepalived
  exit 1
fi
exit 0
`

const configDocker = `DOCKER="/var/lib/docker"
[ -d "$DOCKER" ] || mkdir -p $DOCKER

if [ -x "$(command -v yum)" ]; then
  yum makecache fast
  yum install -y yum-utils \
    device-mapper-persistent-data \
    lvm2
  yum install -y conntrack
  FIREWALL="firewalld"
elif [ -x "$(command -v apt-get)" ]; then
  apt-get update
  apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common
  apt-get install -y conntrack
  FIREWALL="ufw"
else
  echo "$(date) - $0 - [ERROR] - unknown Distributor ID."
  exit 1
fi

systemctl stop $FIREWALL
systemctl disable $FIREWALL
/sbin/iptables -P FORWARD ACCEPT
/sbin/iptables -F && iptables -X && iptables -F -t nat && iptables -X -t nat

# mk docker-iptables.service
cat > /etc/systemd/system/docker-iptables.service <<"EOF"
[Unit]
Description=Make Iptables Rules for Docker

[Service]
Type=oneshot
ExecStart=/bin/sh \
          -c \
          "sleep 60 && /sbin/iptables -P FORWARD ACCEPT"

[Install]
WantedBy=multi-user.target
EOF

[ -d /etc/docker ] || mkdir -p /etc/docker
cat > /etc/docker/daemon.json << EOF
{
  "data-root": "$DOCKER",
  "registry-mirrors" : [
    "https://nmp34hlf.mirror.aliyuncs.com",
    "https://mirror.ccs.tencentyun.com"
  ],
  "insecure-registries" : [
    "192.168.0.0/16",
    "172.0.0.0/8",
    "10.0.0.0/8"
  ],
  "debug" : true,
  "experimental" : true,
  "max-concurrent-downloads" : 10
}
EOF
`

const configIPVS = `NAME=ipvs-mod
BIN=${NAME}.sh
SVC=${NAME}.service
cat > /usr/local/bin/${BIN} << "EOF"
#!/bin/bash
ipvs_modules="ip_vs ip_vs_lc ip_vs_wlc ip_vs_rr ip_vs_wrr ip_vs_lblc ip_vs_lblcr ip_vs_dh ip_vs_sh ip_vs_nq ip_vs_sed ip_vs_ftp nf_conntrack_ipv4"
for kernel_module in ${ipvs_modules}; do
  /sbin/modprobe ${kernel_module}
done
lsmod | grep ip_vs
MODS=" \
net.ipv4.ip_forward^=^1
net.bridge.bridge-nf-call-iptables^=^1
net.bridge.bridge-nf-call-ip6tables^=^1"
FILE=/etc/sysctl.d/ipvs.conf
[ -f $FILE ] && rm -f $FILE
[ -f $FILE ] || touch $FILE
for MOD in $MODS; do
  MOD=$(echo $MOD | tr "^" " ")
  if ! cat $FILE | grep "$MOD"; then
    echo $MOD >> $FILE
  fi
done
sysctl -p ${FILE}
EOF
chmod +x /usr/local/bin/${BIN}
# mk service 
cat > /etc/systemd/system/${SVC} << EOF
[Unit]
Description=Switch-on Kernel Modules Needed by IPVS 

[Service]
Type=oneshot
ExecStart=/usr/local/bin/${BIN}

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ${SVC} 
systemctl restart ${SVC} 
`

const approveNodeCert = `CSRS=$(kubectl get csr | grep Pending | awk -F ' ' '{print $1}')
if [ -n "$CSRS" ]; then
  for CSR in $CSRS; do
    kubectl certificate approve $CSR
  done
fi
`

const clearKubeNode = `[[ $(systemctl is-active kubelet) == "active" ]] && systemctl stop kubelet
[[ $(systemctl is-active kube-proxy) == "active" ]] && systemctl stop kube-proxy
[[ $(systemctl is-active docker) == "active" ]] && systemctl stop docker
mount | grep '/var/lib/kubelet'| awk '{print $3}'| xargs umount
rm -rf /var/lib/kubelet
rm -rf /var/run/docker/
rm -rf /etc/kubernetes
`

const clearKubeMaster = `[[ $(systemctl is-active kube-apiserver) == "active" ]] && systemctl stop kube-apiserver
[[ $(systemctl is-active kube-controller-manager) == "active" ]] && systemctl stop kube-controller-manager
[[ $(systemctl is-active kube-scheduler) == "active" ]] && systemctl stop kube-scheduler
[[ $(systemctl is-active keepalived) == "active" ]] && systemctl stop keepalived 
[[ $(systemctl is-active haproxy) == "active" ]] && systemctl stop haproxy 
rm -rf /var/run/kubernetes
rm -rf /etc/kubernetes
`

const clearKubeEtcd = `[[ $(systemctl is-active etcd) == "active" ]] && systemctl stop etcd
rm -rf /var/lib/etcd
rm -rf /var/lib/etcd-wal
rm -rf /etc/etcd/ssl/*
`

const getInterfaceSc = `IP={{.ip}}
INTERFACE=$(ip addr | grep $IP)
INTERFACE=${INTERFACE##*" "}
echo ${INTERFACE}
`
