package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getAbsPath() string {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	f := os.Args[0]
	l := strings.Split(f, "/")
	name := l[len(l)-1]
	return fmt.Sprintf("%v/%v", path, name)
}

func (this *Core) systemd() {
	if this.config.Nosystemd || this.config.Container {
		return
	}
	// 0 make systemd unit
	path := fmt.Sprintf("/etc/systemd/system/%v.service", projectName)
	if checkFileExist(path) {
		checkPid()
		return
	}
	content := string(projectSvc)
	content = strings.Replace(content, "{{.name}}", strings.ToUpper(projectName), -1)
	pid := fmt.Sprintf("%v/%v.pid", pidpath, projectName)
	content = strings.Replace(content, "{{.pid}}", pid, -1)
	cmd := fmt.Sprintf("%v/%v", string(bin), projectName)
	args := os.Args[1:]
	arg := strings.Join(args, " ")
	content = strings.Replace(content, "{{.cmd}}", fmt.Sprintf("%v %v", cmd, arg), -1)
	f := new(file)
	f.set(path, content)
	f.newFile()
	// 1 mv binary
	src := getAbsPath()
	dest := cmd
	//execCmd(fmt.Sprintf("echo \"%v\" > /tmp/tmp.log", src))
	//execCmd(fmt.Sprintf("echo \"%v\" >> /tmp/tmp.log", dest))
	execCmd(fmt.Sprintf("yes | cp %v %v", src, dest))
	execCmd("chmod a+x dest")
	// 2 start service
	execCmd("systemctl daemon-reload")
	execCmd(fmt.Sprintf("systemctl enable %v", projectName))
	execCmd(fmt.Sprintf("systemctl restart %v", projectName))
	// exit
	os.Exit(0)
	return
}

func checkPid() {
	path := fmt.Sprintf("%v/%v.pid", pidpath, projectName)
	pid := os.Getpid()
	if !checkFileExist(path) {
		f := new(file)
		f.set(path, fmt.Sprintf("%v", pid))
		f.newFile()
		return
	}
	fn, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer fn.Close()
	raw, err := ioutil.ReadAll(fn)
	if err != nil {
		log.Println(err)
	}
	content := strings.Replace(string(raw), "\n", "", 1)
	i, err := strconv.Atoi(content)
	if err != nil {
		log.Fatal(err)
	}
	if pid == i {
		return
	}
	log.Fatalf("already exists: %v, this one exits: %v\n", i, pid)
	os.Exit(1)
}
