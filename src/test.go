package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func (this *Core) handleTest0(w http.ResponseWriter, r *http.Request) {
	res := this.test0()
	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
	return
}

func (this *Core) handleTest3(w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf("%v", this.cidr)
	io.WriteString(w, str)
}

func (this *Core) handleTest(w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf("goto %v, using \"tail -f %v\" to see details", this.ip, kubelog)
	_ = this.test()
	io.WriteString(w, str)
}

func (this *Core) handleTest6(w http.ResponseWriter, r *http.Request) {
	res := this.test0()
	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
	return
}

func (this *Core) handleTest00(w http.ResponseWriter, r *http.Request) {
	tmp := this.getDmapFromCluster()
	res := this.countMap(tmp)
	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
	return
}

func (this *Core) test() error {
	installkube <- "ok"
	return nil
}

func (this *Core) test4() error {
	f := new(file)
	content := "hahaha"
	path := "/tmp/test.txt"
	f.set(path, content)
	//f.Path = path
	//f.Content = content

	// self
	tools := new(funcs)
	tools.Append(*f)

	// others
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		fmt.Printf("addr: %v\n", addr)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("funcs.Append", *f)
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
	}
	return nil
}

func (this *Core) test0() bool {
	path := "/tmp/iamhelloworld"
	return this.getFromOther(path)
}

func (this *Core) test80() []string {
	var str []string
	str = append(str, fmt.Sprintf("%v", this.ipnet))
	str = append(str, this.cidr)
	str = append(str, this.network)
	return str
}

func (this *Core) test20() map[string]string {
	var ret map[string]string
	ret = make(map[string]string)
	// self
	ret[this.ip] = this.kube.role

	// others
	input := new(args)
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		fmt.Printf("addr: %v\n", addr)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("kube.GetRole", *input)
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
		res := new(string)
		ConvertToNormalType(resBuf, res)
		ret[ip] = *res
	}
	return ret
}

func (this *Core) test10() PairList {
	m := this.getScores()
	//fmt.Println(m)
	p := sortMapByValue(m)
	return p
}

func (this *Core) test2() *float64 {
	ip := this.getOthers()[0]
	addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
	fmt.Printf("addr: %v\n", addr)
	cli := Client{Adress: addr}
	resBuf, err := cli.SendRequest("funcs.Score", args{})
	if err != nil {
		log.Fatalln(err)
	}
	ok := ResultIsOk(resBuf)
	fmt.Println(ok)
	res := new(float64)
	ConvertToNormalType(resBuf, res)
	fmt.Printf("%v\n", *res)
	return res
}

func (this *Core) handleTest1(w http.ResponseWriter, r *http.Request) {
	s := this.test()
	//fmt.Printf("%v\n", *s)
	b, err := json.Marshal(s)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
	return
}

func (this *Core) test1() *systeminfo {
	ip := this.getOthers()[0]
	addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
	fmt.Printf("addr: %v\n", addr)
	cli := Client{Adress: addr}
	resBuf, err := cli.SendRequest("funcs.GetStatus", args{})
	if err != nil {
		log.Fatalln(err)
	}
	ok := ResultIsOk(resBuf)
	fmt.Println(ok)
	res := new(systeminfo)
	ConvertToNormalType(resBuf, res)
	fmt.Printf("%v\n", *res)
	return res
}
