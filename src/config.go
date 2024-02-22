package core

const ntpServerConf = `# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).
server 0.centos.pool.ntp.org iburst
server 1.centos.pool.ntp.org iburst
server 2.centos.pool.ntp.org iburst
server 3.centos.pool.ntp.org iburst

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Allow the system clock to be stepped in the first three updates
# if its offset is larger than 1 second.
makestep 1.0 3

# Enable kernel synchronization of the real-time clock (RTC).
rtcsync

# Enable hardware timestamping on all interfaces that support it.
#hwtimestamp *

# Increase the minimum number of selectable sources required to adjust
# the system clock.
#minsources 2

# Allow NTP client access from local network.
allow {{.cidr}} 

# Serve time even if not synchronized to a time source.
#local stratum 10

# Specify file containing keys for NTP authentication.
#keyfile /etc/chrony.keys

# Specify directory for log files.
logdir /var/log/chrony

# Select which information is logged.
#log measurements statistics tracking
`

const ntpClientConf = `# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).
#server 0.centos.pool.ntp.org iburst
#server 1.centos.pool.ntp.org iburst
#server 2.centos.pool.ntp.org iburst
#server 3.centos.pool.ntp.org iburst
server {{.ntp.server}} iburst

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Allow the system clock to be stepped in the first three updates
# if its offset is larger than 1 second.
makestep 1.0 3

# Enable kernel synchronization of the real-time clock (RTC).
rtcsync

# Enable hardware timestamping on all interfaces that support it.
#hwtimestamp *

# Increase the minimum number of selectable sources required to adjust
# the system clock.
#minsources 2

# Allow NTP client access from local network.
#allow 192.168.0.0/16

# Serve time even if not synchronized to a time source.
#local stratum 10

# Specify file containing keys for NTP authentication.
#keyfile /etc/chrony.keys

# Specify directory for log files.
logdir /var/log/chrony

# Select which information is logged.
#log measurements statistics tracking
`

const thisip = `export NET_ID={{.net.id}}
export THIS_IP={{.this.ip}}
export NODE_IP={{.this.ip}}
export MASTER_IP={{.vip}}
export VIP={{.vip}}
export KUBE_APISERVER=https://{{.vip}}:{{.port}}
`

const masteretcdenv = `export NODE_NAME={{.node.name}}
export NODE_IPS="{{.node.ips}}"
export ETCD_NODES={{.etcd.nodes}}
export ETCD_ENDPOINTS={{.etcd.endpoints}}
`

const nodeetcdenv = `export ETCD_NODES={{.etcd.nodes}}
export ETCD_ENDPOINTS={{.etcd.endpoints}}
`

const haproxyCfg = `global
        log /dev/log    local0
        log /dev/log    local1 notice
        chroot /var/lib/haproxy
        stats timeout 30s
        daemon
        nbproc 1
defaults
        log     global
        timeout connect 5000
        timeout client  50000
        timeout server  50000
listen kube-master
    	bind 0.0.0.0:{{.port}}
        mode tcp
        option tcplog
        balance roundrobin
`

const keepalivedConf_P0 = `! Configuration File for keepalived
#__IP__ {{.ip}} 
global_defs {
    notification_email {
    }
    router_id {{.router.id}} 
}

vrrp_script check_haproxy {
    # 自身状态检测
    script "/etc/keepalived/chk.sh"
    interval 3
    weight 5
}

vrrp_instance haproxy-vip {
    # 使用单播通信，默认是组播通信
    unicast_src_ip `

const keepalivedConf_P1 = `    # 虚拟ip 绑定的网卡 （这里根据你自己的实际情况选择网卡）
    interface {{.interface}} 
    #use_vmac
    # 此ID要配置一致
    virtual_router_id {{.virtual.router.id}} 
    # 默认启动优先级，Master要比Backup大点，但要控制量，保证自身状态检测生效
`

const keepalivedConf_P2 = `    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1111
    }
    virtual_ipaddress {
        # 虚拟ip 地址
        {{.vip}} 
    }
    track_script {
        check_haproxy
    }
}
`

const kubernetesCsrJson_P0 = `{
  "CN": "kubernetes",
  "hosts": [
    "127.0.0.1",
`

const kubernetesCsrJson_P1 = `    "10.254.0.1",
    "kubernetes",
    "kubernetes.default",
    "kubernetes.default.svc",
    "kubernetes.default.svc.cluster",
    "kubernetes.default.svc.cluster.local"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "BeiJing",
      "L": "BeiJing",
      "O": "k8s",
      "OU": "System"
    }
  ]
}
`
