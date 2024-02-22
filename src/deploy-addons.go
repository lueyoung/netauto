package core

import (
	//"errors"
	"fmt"
	//"io/ioutil"
	"log"
	"os"
	//"runtime"
	"strings"
	//"sync"
	"bufio"
	//"encoding/base64"
	"io"
)

func getKubeDnsServer() string {
	target := "CLUSTER_DNS_SVC_IP"
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

func getKubeDnsDomain() string {
	target := "CLUSTER_DNS_DOMAIN"
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

func (this *Core) deployKubeDNS() error {
	var c yamler
	c = new(coredns)
	y := new(yamlCreater)
	y.set(c)
	return y.run()
}

func (this *Core) deployKubeDashboard() error {
	var d yamler
	d = new(dashboard)
	y := new(yamlCreater)
	y.set(d)
	return y.run()
}
