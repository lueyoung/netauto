package core

import (
	//"encoding/json"
	//"errors"
	"fmt"
	//"github.com/hashicorp/raft"
	//"io"
	//"io/ioutil"
	"log"
	//"net/http"
	//"reflect"
	"os"
	"runtime"
	"strings"
	"sync"
)

func createKubeLog(path string) *log.Logger {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(f, "", log.LstdFlags)
}

func scp(src, dest string) error {
	cmd := fmt.Sprintf("scp -o \"StrictHostKeyChecking no\" %v %v", src, dest)
	_, err := execCmdLogged(cmd)
	if err != nil {
		log.Println(err.Error())
	}
	return nil
}

func (this *Core) distribute(path string) error {
	n := len(this.getOthers())
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, ip := range this.getOthers() {
		go func(ip, path string) {
			defer wg.Done()
			cmd := fmt.Sprintf("scp -o \"StrictHostKeyChecking no\" %v root@%v:%v", path, ip, path)
			_, err := execCmdLogged(cmd)
			if err != nil {
				log.Println(err.Error())
			}
		}(ip, path)
	}
	wg.Wait()
	return nil
}

func (this *Core) distribute2master(path string) error {
	n := len(this.kubeclusterinfo.getMaster()) - 1
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, ip := range this.kubeclusterinfo.getMaster() {
		if ip != this.ip {
			go func(ip, path string) {
				defer wg.Done()
				cmd := fmt.Sprintf("scp -o \"StrictHostKeyChecking no\" %v root@%v:%v", path, ip, path)
				_, err := execCmdLogged(cmd)
				if err != nil {
					log.Println(err.Error())
				}
			}(ip, path)
		}
	}
	wg.Wait()
	return nil
}

func (this *Core) clusterRun(cmd, arg string) error {
	c := new(linuxcmd)
	c.set(cmd, arg)
	n := len(this.getOthers()) + 1
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	// 2.1 self
	go func(c *linuxcmd) {
		defer wg.Done()
		c.runLogged()
	}(c)
	// 2.2 others
	for _, ip := range this.getOthers() {
		go func(ip string, c *linuxcmd) {
			defer wg.Done()
			addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
			fmt.Printf("addr: %v\n", addr)
			cli := Client{Adress: addr}
			resBuf, err := cli.SendRequest("funcs.RunLogged", *c)
			if err != nil {
				log.Fatalln(err)
			}
			ok := ResultIsOk(resBuf)
			fmt.Println(ok)
		}(ip, c)
	}
	wg.Wait()

	return nil
}

func (this *Core) kubeMasterRun(cmd, arg string) error {
	c := new(linuxcmd)
	c.set(cmd, arg)
	n := len(this.kubeclusterinfo.getMaster())
	runtime.GOMAXPROCS(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, ip := range this.kubeclusterinfo.getMaster() {
		if ip != this.ip {
			go func(ip string, c *linuxcmd) {
				defer wg.Done()
				addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
				fmt.Printf("addr: %v\n", addr)
				cli := Client{Adress: addr}
				resBuf, err := cli.SendRequest("funcs.Run", *c)
				if err != nil {
					log.Fatalln(err)
				}
				ok := ResultIsOk(resBuf)
				fmt.Println(ok)
			}(ip, c)
		} else {
			go func(c *linuxcmd) {
				defer wg.Done()
				c.run()
			}(c)
		}
	}
	wg.Wait()

	return nil
}

func configNtpServer(f file) error {
	path, cidr := f.get()
	conf := ntpServerConf
	conf = strings.Replace(conf, "{{.cidr}}", cidr, -1)
	f.set(path, conf)
	return f.newFile()
}

func (this *Core) makeScriptThenRun(method string) error {
	path := fmt.Sprintf("/tmp/%v.sh", method)
	f := new(file)
	switch method {
	case "optKernel":
		f.set(path, string(optKernel))
	case "shutdownHarmful":
		f.set(path, string(shutdownHarmful))
	case "installUseful":
		f.set(path, string(installUseful))
	case "makeBaseEnvVariables":
		f.set(path, string(makeBaseEnvVariables))
	case "makeEnvConf":
		f.set(path, string(makeEnvConf))
	case "sourceEnv":
		f.set(path, string(sourceEnv))
	case "installVIP":
		f.set(path, string(installVIP))
	}
	err := f.newExec()
	if err != nil {
		return err
	}
	this.distribute(path)
	this.clusterRun(path, "")
	return nil
}

func (this *Core) makeScriptThenKubeMasterRun(method string) error {
	path := fmt.Sprintf("/tmp/%v.sh", method)
	f := new(file)
	switch method {
	case "optKernel":
		f.set(path, string(optKernel))
	case "shutdownHarmful":
		f.set(path, string(shutdownHarmful))
	case "installUseful":
		f.set(path, string(installUseful))
	case "makeBaseEnvVariables":
		f.set(path, string(makeBaseEnvVariables))
	case "makeEnvConf":
		f.set(path, string(makeEnvConf))
	case "sourceEnv":
		f.set(path, string(sourceEnv))
	case "installVIP":
		f.set(path, string(installVIP))
	}
	err := f.newExec()
	if err != nil {
		return err
	}
	this.distribute2master(path)
	this.kubeMasterRun(path, "")
	return nil
}

func (this *Core) run2makeFileThenDistrbute(method string) error {
	/**
	path := fmt.Sprintf("/tmp/%v.sh", method)
	f := new(file)
	switch method {
	case "optKernel":
		f.set(path, string(optKernel))
	case "shutdownHarmful":
		f.set(path, string(shutdownHarmful))
	case "installUseful":
		f.set(path, string(installUseful))
	case "makeBaseEnvVariables":
		f.set(path, string(makeBaseEnvVariables))
	case "sourceEnv":
		f.set(path, string(sourceEnv))
	}
	err := f.newExec()
	if err != nil {
		return err
	}
	_, err = execCmd(path)
	if err != nil {
		return err
	}
	switch method {
	case "makeBaseEnvVariables":
		this.distribute("/etc/kubernetes/env/k8s.env")
		this.distribute("/etc/kubernetes/token.csv")
	}
	return nil**/
	c := new(linuxcmd)
	switch method {
	case "optKernel":
		c.set(string(optKernel), "")
	case "shutdownHarmful":
		c.set(string(shutdownHarmful), "")
	case "installUseful":
		c.set(string(installUseful), "")
	case "makeBaseEnvVariables":
		c.set(string(makeBaseEnvVariables), "")
	case "sourceEnv":
		c.set(string(sourceEnv), "")
	}
	_, err := c.runLogged()
	if err != nil {
		return err
	}
	switch method {
	case "makeBaseEnvVariables":
		err := this.distribute("/etc/kubernetes/env/k8s.env")
		if err != nil {
			return err
		}
		err = this.distribute("/etc/kubernetes/token.csv")
		if err != nil {
			return err
		}
	}
	return nil

}

/**
func (this *Core) distributeWrite2Profile() error {
	path := "/tmp/toSource"
	f := new(file)
	f.set(path, string(toSource))
	err := f.newFile()
	if err != nil {
		return err
	}
	this.distribute(path)
	return nil
}**/

func (this *Core) getKubeMaster() []string {
	return this.kubeclusterinfo.getMaster()
}

func (this *Core) getKubeNode() []string {
	return this.kubeclusterinfo.getNode()
}

func (this *Core) getKubeVIP() string {
	return this.kubeclusterinfo.getVIP()
}

func (this *Core) makeAndDistributeSystemdUnit(path, content string) error {
	f := new(file)
	f.set(path, content)
	f.newFile()
	err := this.distribute(path)
	if err != nil {
		return err
	}
	return nil
}

func (this *Core) getFromOther(path string) bool {
	f := new(file)
	f.set(path, "")
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("funcs.IfExist", *f)
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
		res := new(bool)
		ConvertToNormalType(resBuf, res)
		if *res {
			scp(fmt.Sprintf("root@%v:%v", ip, path), path)
			return true
		}
	}
	return false
}
