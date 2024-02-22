package core

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func getIPNet(ip string) *net.IPNet {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.String() == ip {
			return ipnet
		}
	}
	return nil
}

func Test_Scan(t *testing.T) {
	ip := "192.168.100.167"
	ipnet := getIPNet(ip)
	c := fmt.Sprintf("%v", ipnet)
	t.Log(c)
	mask := 24
	cidr := fmt.Sprintf("%v/%v", ip, mask)
	n := getCidrHostNum(mask)
	t.Logf("hosts num: %v\n", n)
	t.Logf("cidr: %v\n", cidr)
	looper(cidr)
	minIp, maxIp := getCidrIpRange(cidr)
	fmt.Println("CIDR最小IP：", minIp, " CIDR最大IP：", maxIp)
}
