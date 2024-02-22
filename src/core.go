package core

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/raft"
	"memberlist"
)

var (
	mtx        sync.RWMutex
	items      = map[string]string{}
	broadcasts *memberlist.TransmitLimitedQueue
)

type Core struct {
	// basic
	id      string
	ip      string
	cidr    string
	min     bool
	method  string
	join    string
	force   bool
	ipnet   *net.IPNet
	network string
	config  *Config
	typed   string

	// memberlist
	list       *memberlist.Memberlist
	broadcasts *memberlist.TransmitLimitedQueue
	members    map[string]string

	// raft
	raftd      int
	raft       *raft.Raft
	RaftDir    string
	RaftBind   string
	logger     *log.Logger
	mu         sync.Mutex
	inmem      bool
	m          map[string]string
	raftConfig *raft.Config
	transport  *raft.NetworkTransport

	// http db
	addr string
	ln   net.Listener

	// distributed shared memory
	maxmem  int
	memsize int
	mem     *dsm
	dsm     *dmap

	// rpc
	rpc *Server

	// status
	state *systeminfo

	// sort
	sort []string

	// kubernetes
	kube            *kube
	kubeclusterinfo *kubeclusterinfo
	kubelog         *log.Logger
}

func New() (*Core, error) {
	c := &Core{}
	return c, nil
}

func Create(config Config) (*Core, error) {
	this := &Core{
		network: config.Network,
		join:    config.Join,
		force:   config.Force,
		maxmem:  config.Maxmem * 1000 * 1000,
	}
	this.config = &config
	network := config.Network
	join := config.Join
	this.ifVagrant()

	if join == "" {
		this.method = "scan"
	} else {
		if network == "" {
			this.network = strings.Split(join, ",")[0]
		}
		this.method = "join"
		this.join = join
	}
	this.mkNetwork()
	this.systemd()

	// introspection for node type
	this.makeTyped()

	this.configMemberlist()
	node := this.list.LocalNode()
	ip := fmt.Sprintf("%v", node.Addr)
	id := strings.Split(ip, ".")[3]
	if this.network == "" {
		this.ip = ip
		this.id = id
		network := ""
		sep := ""
		nums := strings.Split(ip, ".")
		for i := 0; i < 3; i++ {
			network += sep
			network += nums[i]
			sep = "."
		}
		this.network = network
	} else {
		this.mkThisIp()
	}
	this.members = make(map[string]string)

	// configure raft
	this.configRaft()

	this.addr = fmt.Sprintf("%v:%v", this.ip, httpPort)

	// config distrbuted shared memory
	zero := "null"
	m := &dsm{A: zero, B: zero, C: zero}
	this.memsize = SizeTOf(&m)
	this.mem = m
	dm := new(dmap)
	dm.data = make(map[string]string)
	this.dsm = dm

	// init system info
	s := makeSystemInfo()
	this.state = s

	// for kubernetes
	k := new(kube)
	this.kube = k
	kci := new(kubeclusterinfo)
	this.kubeclusterinfo = kci

	// kubernetes log
	this.kubelog = createKubeLog(string(kubelog))

	// make rpc server
	server := MakeNewServer()
	xx := &xxx{1, 2, "for test"}
	yy := &yyy{0}
	f := &funcs{}
	server.Install(xx)
	server.Install(yy)
	server.Install(m)
	server.Install(f)
	server.Install(k)
	server.Install(kci)
	server.Install(dm)
	this.rpc = server

	// set parameters before running
	this.backupParameter()

	go this.introduce()
	//go this.listenCh()
	go this.listenIDCh()
	//go this.listenDsmCh()
	//go core.startRaft()
	go this.maintainRaftNode()
	go this.addRaftNode()
	go this.leaderinfo()
	//go this.evaluate()
	go this.kubeLooper()
	go this.installKube()
	go this.syncDsm()
	return this, nil
}

func (this *Core) configMemberlist() {
	c := memberlist.DefaultLocalConfig()
	if this.network != "" {
		ip, err := this.getThisIp()
		if err != nil {
			log.Fatal(err)
		}
		c.BindAddr = ip
		fmt.Printf("c.BindAddr: %v\n", c.BindAddr)
	}
	c.Delegate = &delegate{}
	c.Name = fmt.Sprintf("%v-%v", c.Name, c.BindAddr)
	m, err := memberlist.Create(c)
	if err != nil {
		log.Fatal("Failed to create memberlist: " + err.Error())
	}
	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}
	this.list = m
	this.broadcasts = broadcasts
}

func (this *Core) configRaft() {
	this.raftd = 0
	this.m = make(map[string]string)
	this.inmem = false
	prefix := fmt.Sprintf("[%v] ", projectName)
	this.logger = log.New(os.Stderr, prefix, log.LstdFlags)
	this.RaftBind = fmt.Sprintf("%v:%v", this.ip, raftPort)
	fmt.Printf("Raft bind to: %v\n", this.RaftBind)
	this.RaftDir = "/tmp"
	this.createRaft()
	if enableSingle {
		log.Println("start raft bootstrap")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      this.raftConfig.LocalID,
					Address: this.transport.LocalAddr(),
				},
			},
		}
		this.raft.BootstrapCluster(configuration)
		//var servers []raft.Server
	}
}

func (this *Core) reconnect() {
	if this.force {
		return
	}
	configFuture := this.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		this.logger.Printf("failed to get raft configuration: %v", err)
		return
	}
	var join []string
	self := this.ip
	for _, srv := range configFuture.Configuration().Servers {
		ipport := fmt.Sprintf("%v", srv.Address)
		ip := strings.Split(ipport, ":")[0]
		fmt.Printf("ip in reconnect: %v\n", ip)
		if self == ip {
			continue
		}
		fmt.Printf("ip to reconnect: %v\n", ip)
		join = append(join, ip)
	}
	fmt.Printf("reconnect join array: %v\n", join)
	if len(join) == 0 {
		fmt.Printf("empty join array, exit\n")
		return
	}
	_, err := this.list.Join(join)
	if err != nil {
		log.Printf("reconnect cannot join: %v\n", join)
	}
	fmt.Printf("reconnect join: %v\n", join)
	return
}

func (this *Core) Start() error {
	var err error
	switch this.method {
	case "scan":
		err = this.scan()
	case "join":
		err = this.jointo()
	}
	if err != nil {
		return err
	}

	server := http.Server{
		Handler: this,
	}

	//ln, err := net.Listen("tcp", this.addr)
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", httpPort))
	if err != nil {
		return err
	}
	this.ln = ln
	http.Handle("/", this)
	log.Fatal(server.Serve(this.ln))
	return nil
}

func (this *Core) makeScanRange() string {
	mask, _ := strconv.Atoi(strings.Split(this.cidr, "/")[1])
	if mask < 32 {
		return this.cidr
	}
	ip := strings.Split(this.cidr, "/")[0]
	ipSegs := strings.Split(ip, ".")
	return fmt.Sprintf("%v.%v.0.0/16", ipSegs[0], ipSegs[1])
}

func (this *Core) scanV1() error {
	this.reconnect()
	cidr := this.makeScanRange()
	ip := strings.Split(cidr, "/")[0]
	ipSegs := strings.Split(ip, ".")
	maskLen, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
	seg2MinIp, seg2MaxIp := getIpSeg2Range(ipSegs, maskLen)
	seg3MinIp, seg3MaxIp := getIpSeg3Range(ipSegs, maskLen)
	seg4MinIp, seg4MaxIp := getIpSeg4Range(ipSegs, maskLen)
	for i := seg2MinIp; i <= seg2MaxIp; i++ {
		for j := seg3MinIp; j <= seg3MaxIp; j++ {
			for k := seg4MinIp; k <= seg4MaxIp; k++ {
				tmp := fmt.Sprintf("%v.%v.%v.%v", ipSegs[0], i, j, k)
				if this.try2join(tmp) {
					return nil
				}
			}
		}
	}
	log.Println("Start as a cluster head")
	return nil
}

func (this *Core) scan() error {
	this.reconnect()
	var wg sync.WaitGroup
	cidr := this.makeScanRange()
	ip := strings.Split(cidr, "/")[0]
	ipSegs := strings.Split(ip, ".")
	maskLen, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
	seg2MinIp, seg2MaxIp := getIpSeg2Range(ipSegs, maskLen)
	seg3MinIp, seg3MaxIp := getIpSeg3Range(ipSegs, maskLen)
	seg4MinIp, seg4MaxIp := getIpSeg4Range(ipSegs, maskLen)
	joined := new(mymap)
	for i := seg2MinIp; i <= seg2MaxIp; i++ {
		for j := seg3MinIp; j <= seg3MaxIp; j++ {
			/**for k := seg4MinIp; k <= seg4MaxIp; k++ {
			        tmp := fmt.Sprintf("%v.%v.%v.%v", ipSegs[0], i, j, k)
			        if this.try2join(tmp) {
			                return nil
			        }
			}**/
			total := seg4MaxIp - seg4MinIp + 1
			b := total / scanBatch
			r := total % scanBatch
			for k := 0; k < b; k++ {
				runtime.GOMAXPROCS(scanBatch)
				wg.Add(scanBatch)
				joined.data = make(map[string]bool)
				for l := 0; l < scanBatch; l++ {
					go func(l int) {
						defer wg.Done()
						tmp := fmt.Sprintf("%v.%v.%v.%v", ipSegs[0], i, j, k*scanBatch+seg4MinIp+l)
						v := this.try2join(tmp)
						joined.set(tmp, v)
					}(l)
				}
				wg.Wait()
				if res := trueInMap(joined.get()); res {
					return nil
				}
			}
			runtime.GOMAXPROCS(r)
			wg.Add(r)
			joined.data = make(map[string]bool)
			for k := 0; k < r; k++ {
				go func(k int) {
					defer wg.Done()
					tmp := fmt.Sprintf("%v.%v.%v.%v", ipSegs[0], i, j, b*scanBatch+k+seg4MinIp)
					v := this.try2join(tmp)
					joined.set(tmp, v)
				}(k)
			}
			wg.Wait()
			if res := trueInMap(joined.get()); res {
				return nil
			}
		}
	}
	log.Println("Start as a cluster head")
	return nil
}

func (this *Core) try2join(ip string) bool {
	n := len(this.list.Members())
	if n > 1 {
		log.Printf("be found, exit")
		return true
	}
	if ip == this.ip {
		log.Printf("meet self: %v, continue\n", ip)
		return false
	}
	_, err := this.list.Join([]string{ip})
	if err != nil {
		log.Printf("Cannot join the node: %v\n", ip)
		return false
	}
	log.Println("Start as a cluster member")
	return true
}

func (this *Core) jointo() error {
	ip := strings.Split(this.join, ",")
	_, err := this.list.Join(ip)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) mkNetwork() {
	if this.network == "" {
		this.introspection()
	}
	network := this.network
	nums := strings.Split(network, ".")
	if len(nums) < 2 || nums[1] == "" {
		log.Fatalf("set an IP as 1.2.3.4 rather than: %v\n", network)
	}
	c := ""
	sep := ""
	n := len(nums)
	if len(nums) > 3 {
		n = 3
	}
	for i := 0; i < n; i++ {
		c += sep
		c += nums[i]
		sep = "."
	}
	this.network = c
}

func (this *Core) mkThisIp() {
	ips, err := getIntranetIp()
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range ips {
		if strings.Contains(ip, this.network) {
			this.ip = ip
			id := strings.Split(ip, ".")[3]
			this.id = id
			this.ipnet = getIPNet(ip)
			this.cidr = fmt.Sprintf("%v", this.ipnet)
			return
		}
	}
}

func getIntranetIp() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, fmt.Sprintf("%v", ipnet.IP.To4()))
			}

		}
	}
	return ips, nil
}

func (this *Core) getThisIp() (string, error) {
	ips, err := getIntranetIp()
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range ips {
		if strings.Contains(ip, this.network) {
			return ip, nil
		}
	}
	return "", fmt.Errorf("err: not found ip in network %v", this.network)
}

func (this *Core) Addr() net.Addr {
	return this.Addr()
}

func (this *Core) SyncKubeClusterInfo(k kubeclusterinfo) error {
	this.kubeclusterinfo = &k
	return nil
}

func (this *Core) introspection() {
	ipnets := GetIntranetIP(this.config.Vagrant)
	if len(ipnets) > 0 {
		//log.Fatal("cannot find local IP ... exit")
		ipnet := ipnets[0]
		this.ipnet = ipnet
		this.ip = ipnet.IP.String()
		this.network = ipnet.IP.String()
		this.cidr = fmt.Sprintf("%v", ipnet)
		return
	}
	// for container
	ipnets = GetIntranetIP4More()
	if len(ipnets) == 0 {
		log.Fatal("cannot find local IP ... exit")
		return
	}
	this.config.Container = true
	ipnet := ipnets[0]
	this.ipnet = ipnet
	this.ip = ipnet.IP.String()
	this.network = ipnet.IP.String()
	this.cidr = clusterCidr
	if !cidrContains(this.cidr, this.ip) {
		log.Fatal("cannot find local IP ... exit")
		return
	}
	return
}

func (this *Core) ifVagrant() {
	if this.config.Vagrant {
		debug.Println("config.vagrant is true")
		return
	}
	componet := "/usr/sbin/VBoxService"
	cmd := fmt.Sprintf("ps aux | grep -v grep | grep \"%v\"", componet)
	res, _ := execCmd(cmd)
	debug.Printf("cmd res: %v", res)
	if res != "" {
		this.config.Vagrant = true
		debug.Printf("cmd res is not nil: %v", res)
	}
}

func (this *Core) makeTyped() {
	if this.config.Vagrant {
		this.typed = "vm"
		return
	}
	if this.config.Container {
		this.typed = "container"
		return
	}
	this.typed = "physical"
	return
}

func (this *Core) getTyped() string {
	return this.typed
}

func (this *Core) backupParameter() {
	this.kube.Typed = this.typed
}
