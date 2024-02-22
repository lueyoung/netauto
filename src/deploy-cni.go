package core

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	//"runtime"
	"strings"
	//"sync"
	"bufio"
	"encoding/base64"
	"io"
)

func getKubeClusterCidr() string {
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
		if strings.Contains(line, "CLUSTER_CIDR") {
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

func getKubeEtcdKeyPem() string {
	return fileEncodedByBase64("/etc/kubernetes/ssl/etcd-key.pem")
}

func getKubeEtcdPem() string {
	return fileEncodedByBase64("/etc/kubernetes/ssl/etcd.pem")
}

func getKubeCAPem() string {
	return fileEncodedByBase64("/etc/kubernetes/ssl/ca.pem")
}

func getKubeEtcdEndpoints() string {
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
		if strings.Contains(line, "ETCD_ENDPOINTS") {
			ret = line
			break
		} else {
			fmt.Println(line)
		}
	}
	ret = strings.Trim(ret, "\n")
	ret = strings.Split(ret, "=")[1]
	return ret
}

func fileEncodedByBase64(path string) string {
	fn, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fn.Close()
	b, err := ioutil.ReadAll(fn)
	if err != nil {
		log.Fatal(err)
	}

	return strings.Trim(base64.StdEncoding.EncodeToString(b), "\n")
}

func (this *Core) deployCalico() error {
	var c yamler
	c = new(calico)
	y := new(yamlCreater)
	y.set(c)
	return y.run()
}
