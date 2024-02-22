package core

import (
	"fmt"
)

type downlink struct {
	url  string
	path string
}

func (d *downlink) check() bool {
	return checkFileExist(d.path)
}

func (d *downlink) download() (string, error) {
	if d.check() {
		return "Already existed", nil
	}
	cmd := new(linuxcmd)
	var arg string
	if d.path == "" {
		arg = ""
	} else {
		arg = fmt.Sprintf("-O %v", d.path)
	}
	cmd.set(fmt.Sprintf("wget -c %v", d.url), arg)
	return cmd.runLogged()
}

func (d *downlink) set(durl, path string) error {
	d.url = durl
	d.path = path
	return nil
}
