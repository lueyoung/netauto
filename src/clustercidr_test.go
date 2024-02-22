package core

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"
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
		line, err = rd.ReadString('\n') // 以'\n'为结束符读入一行

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
	//ret= strings.Trim(ret(, "\n"), "=")
	ret = strings.Split(ret, "=\"")[1]
	//ret= strings.Split(strings.Trim(ret(, "\n"), "=")[0]
	return ret
	//return strings.Split(strings.Trim(string(b), "\n"), ",")[0]
}

func Test_Cidr(t *testing.T) {
	str := getKubeClusterCidr()
	t.Log(str)
}
