package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
)

func (this *Core) handleDsmMax(w http.ResponseWriter, r *http.Request) {
	v := fmt.Sprintf("%v", this.maxmem)
	k := "MaxMemory"
	b, err := json.Marshal(map[string]string{k: v})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}

func (this *Core) handleDsmSize(w http.ResponseWriter, r *http.Request) {
	v := fmt.Sprintf("%v", this.memsize)
	k := "SizeOf"
	b, err := json.Marshal(map[string]string{k: v})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}

func (this *Core) handleDsmRequest(w http.ResponseWriter, r *http.Request) {
	getKey := func() string {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			return ""
		}
		return parts[2]
	}

	switch r.Method {
	case "GET":
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
		}
		var v string
		r := reflect.ValueOf(this.mem).Elem()
		switch k {
		case "size":
			v = fmt.Sprintf("%v Bytes", this.memsize)
		case "max", "maxmem":
			v = fmt.Sprintf("%v MB", this.maxmem/1000/1000)
		default:
			//v = r.FieldByName("A").String()
			//v = fmt.Sprintf("%v",t.FieldByName("A"))
			//v = r.FieldByName("A").String()
			v = r.FieldByName(strings.ToUpper(k)).String()
		}
		b, err := json.Marshal(map[string]string{k: v})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(b))

	case "PUT", "POST":
		// Read the value from the POST body.
		m := map[string]string{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for k, v := range m {
			if this.raft.State() == raft.Leader {
				err := this.validation(k, v)
				if err != nil {
					b, err0 := json.Marshal(map[string]string{"err": fmt.Sprintf("%v", err)})
					if err0 != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					io.WriteString(w, string(b))
					return
				}
				this.dsmWrite(k, v)
				this.broadcastDsmWrite(k, v)
				b, err := json.Marshal(map[string]string{k: v})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				io.WriteString(w, string(b))
			} else {
				ipport := fmt.Sprintf("%v", this.raft.Leader())
				ip := strings.Split(ipport, ":")[0]
				url := fmt.Sprintf("http://%v:%v/dsm", ip, httpPort)
				j := fmt.Sprintf("{\"%v\":\"%v\"}", k, v)
				typed := "application/json"
				fmt.Printf("redirect url: %v\n", url)
				fmt.Printf("redirect json: %v\n", j)
				client := &http.Client{}
				resp, err := client.Post(url, typed, strings.NewReader(j))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				//ret := fmt.Sprintf("%v",*resp.Body)
				//io.WriteString(w, ret)
				resp.Write(w)
			}
		}

	}
}

func (this *Core) validation(k, v string) error {
	if v == "" {
		return errors.New(fmt.Sprintf("v is null"))
	}
	maxsize := this.maxmem
	currentsize := this.memsize
	//maxsize := 20
	//currentsize := 15
	r := reflect.ValueOf(this.mem).Elem()
	v0 := r.FieldByName(strings.ToUpper(k)).String()
	str := "invalid Value"
	if strings.Contains(v0, "invalid Value") {
		return errors.New(str)
	}
	size0 := SizeOf(v0)
	size := SizeOf(v)
	if size-size0 <= 0 {
		return nil
	}
	if currentsize+size-size0 >= maxsize {
		err := fmt.Sprintf("memory overflow, max:%v, current:%v, to set:%v", maxsize, currentsize, size-size0)
		return errors.New(err)
		//return errors.New(fmt.Sprintf("memory overflow, max:%v, current:%v, to set:%v", maxsize, currentsize, size-size0))
	}
	return nil
}

func (this *Core) dsmWrite(k, v string) error {
	r := reflect.ValueOf(this.mem).Elem()
	r.FieldByName(strings.ToUpper(k)).Set(reflect.ValueOf(v))
	this.memsize = SizeTOf(&this.mem)
	//r0 := reflect.ValueOf(this.mem).Elem()
	//v0 := r0.FieldByName(strings.ToUpper(k)).String()
	//r0.FieldByName(strings.ToUpper(k)).Set(reflect.ValueOf(v))
	//r1 := reflect.ValueOf(this.mem).Elem()
	//v1 := r1.FieldByName(strings.ToUpper(k)).String()
	//fmt.Printf("previoous: %v, current: %v\n", v0, v1)
	return nil
}

func (this *Core) broadcastDsmWrite(k, v string) error {
	for _, ip := range this.getOthers() {
		kv := args{}
		kv.K = k
		kv.V = v
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("dsm.Set", kv)
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
	}
	return nil
}

/*
func (this *Core) broadcastDsmWrite(k, v string) error {
	str := fmt.Sprintf("d %v,%v", k, v)
	//fmt.Printf("str to broadcast: %v\n", str)
	this.broadcasts.QueueBroadcast(&broadcast{
		msg:    []byte(str),
		notify: nil,
	})
	return nil
}*/

func (this *Core) getOthers() []string {
	var others []string
	for _, member := range this.list.Members() {
		ip := fmt.Sprintf("%v", member.Addr)
		if ip != this.ip {
			others = append(others, ip)
		}
	}
	return sortString(others)
}

func (this *Core) getAllMembers() []string {
	var all []string
	for _, member := range this.list.Members() {
		ip := fmt.Sprintf("%v", member.Addr)
		all = append(all, ip)
	}
	return sortString(all)
}
