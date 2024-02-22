package main

import (
	"core"
	"flag"
	"log"
)

var (
	network   = flag.String("network", "", "Specify the CIDR, or define an example IP")
	join      = flag.String("join", "", "Specify the IP or IPs (in term of CSV) to join")
	force     = flag.Bool("force", false, "Force to boot as a new node")
	maxmem    = flag.Int("maxmem", 100, "Max space of distributed shared memory, in term of MB")
	vagrant   = flag.Bool("vagrant", false, "Specify if the program resides in a Vagrant VM")
	nosystemd = flag.Bool("nosystemd", false, "Specify if using Systemd as wrapper")
)

func init() {
	flag.Parse()
}

func main() {
	config := core.Config{
		Network:   *network,
		Join:      *join,
		Force:     *force,
		Maxmem:    *maxmem,
		Vagrant:   *vagrant,
		Nosystemd: *nosystemd,
	}
	c, err := core.Create(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(c.Start())
}
