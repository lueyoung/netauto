package core

import (
	"fmt"
	"strings"
)

type yamler interface {
	sed() map[string]string
	path() string
	yaml() map[string]string
}

type yamlCreater struct {
	yaml yamler
}

func (this *yamlCreater) set(y yamler) {
	this.yaml = y
}

func (y *yamlCreater) mkdir() {
	path := y.yaml.path()
	log2kubelog <- path
	execCmd(fmt.Sprintf("mkdir -p %v", path))
}

func (y *yamlCreater) make() {
	for name, content := range y.yaml.yaml() {
		name = fmt.Sprintf("%v/%v", y.yaml.path(), name)
		content = y.replace(content)
		f := new(file)
		f.set(name, content)
		f.newFile()
	}
}
func (y *yamlCreater) deploy() error {
	cmd := fmt.Sprintf("%v/kubectl create -f", string(bin))
	for name, _ := range y.yaml.yaml() {
		name = fmt.Sprintf("%v/%v", y.yaml.path(), name)
		_, err := execCmd(fmt.Sprintf("%v %v", cmd, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (y *yamlCreater) replace(content string) string {
	for k, v := range y.yaml.sed() {
		content = strings.Replace(content, k, v, -1)
	}
	return content
}

func (y *yamlCreater) run() error {
	y.mkdir()
	y.make()
	return y.deploy()
}
