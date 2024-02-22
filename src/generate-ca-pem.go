package core

import (
	//"errors"
	"fmt"
	//"io/ioutil"
	//"log"
	//"os"
	//"runtime"
	//"strings"
	//"sync"
)

func (this *Core) downloadCfssl() error {
	urls := []string{fmt.Sprintf("https://pkg.cfssl.org/%v/cfssl_linux-amd64", cfsslVersion), fmt.Sprintf("https://pkg.cfssl.org/%v/cfssljson_linux-amd64", cfsslVersion), fmt.Sprintf("https://pkg.cfssl.org/%v/cfssl-certinfo_linux-amd64", cfsslVersion)}
	paths := []string{"/tmp/cfssl", "/tmp/cfssljson", "/tmp/cfssl-certinfo"}
	for i := 0; i < len(paths); i++ {
		path := paths[i]
		if !checkFileExist(path) {
			//0 get from other
			if !this.getFromOther(path) {
				// 1 download
				d := new(downlink)
				d.set(urls[i], path)
				d.download()
			}
		}

	}
	return nil
}

func (this *Core) makeCfsslAvailable() error {
	// chmod, and then mv cfssl to ${PATH}
	components := []string{"cfssl", "cfssljson", "cfssl-certinfo"}
	n := len(components)
	var src, dest []string
	for _, c := range components {
		src = append(src, fmt.Sprintf("/tmp/%v", c))
		dest = append(dest, fmt.Sprintf("%v/%v", string(bin), c))
	}
	for i := 0; i < n; i++ {
		str0 := fmt.Sprintf("chmod a+x %v", src[i])
		str1 := fmt.Sprintf("yes | cp %v %v", src[i], dest[i])
		fmt.Println(str0)
		fmt.Println(str1)
		cmd := new(linuxcmd)
		cmd.set(str0, "")
		cmd.run()
		cmd.set(str1, "")
		cmd.run()
	}
	return nil
}

func (this *Core) generateRootCA() error {
	f := new(file)
	var path, content string
	// 0 generate config
	path = "/etc/kubernetes/ssl/ca-config.json"
	content = string(caConfigJson)
	f.set(path, content)
	f.newFile()
	// 1 generate csr
	path = "/etc/kubernetes/ssl/ca-csr.json"
	content = string(caCsrJson)
	f.set(path, content)
	f.newFile()
	// 2 generate cert
	cmd := new(linuxcmd)
	str := "cd /etc/kubernetes/ssl && cfssl gencert -initca ca-csr.json | cfssljson -bare ca"
	cmd.set(str, "")
	cmd.runLogged()
	return nil
}

func (this *Core) distributeRootCA() error {
	components := []string{"ca-config.json", "ca.csr", "ca-csr.json", "ca-key.pem", "ca.pem"}
	path := "/etc/kubernetes/ssl"
	for _, c := range components {
		f := fmt.Sprintf("%v/%v", path, c)
		err := this.distribute(f)
		if err != nil {
			return err
		}
	}
	return nil
}
