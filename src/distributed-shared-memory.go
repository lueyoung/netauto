package core

import (
	"encoding/json"
	//"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"log"
	"net/http"
	//"reflect"
	"strings"
	"sync"
	"time"
)

type dmap struct {
	data map[string]string
	lock sync.Mutex
}

func (this *dmap) GetMap(kv args) map[string]string {
	return this.data
}

func (this *dmap) Set(kv args) error {
	this.set(kv.K, kv.V)
	return nil
}

func (this *dmap) Delete(kv args) error {
	this.delete(kv.K)
	return nil
}

func (this *dmap) set(k, v string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.data[k] = v
}

func (this *dmap) delete(k string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.data, k)
}

func (this *dmap) get(k string) string {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.data[k]
}

func (this *Core) handleAllDmapRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		b, err := json.Marshal(this.dsm.data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(b))
	}
}

func (this *Core) handleDmapRequest(w http.ResponseWriter, r *http.Request) {
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
		v := this.dsm.data[k]
		b, err := json.Marshal(map[string]string{k: v})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(b))

	case "PUT", "POST":
		// Read the value from the POST body.
		m := make(map[string]string)
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for k, v := range m {
			if this.raft.State() == raft.Leader {
				this.dmapWrite(k, v)
				this.broadcastDmapWrite(k, v)
				b, err := json.Marshal(map[string]string{k: v})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				io.WriteString(w, string(b))
			} else {
				ipport := fmt.Sprintf("%v", this.raft.Leader())
				ip := strings.Split(ipport, ":")[0]
				url := fmt.Sprintf("http://%v:%v/mem", ip, httpPort)
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
	case "DELETE":
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		v := this.dsm.get(k)
		if v == "" {
			res := fmt.Sprintf("no such key: %v", k)
			b, err := json.Marshal(map[string]string{"ret": res})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			io.WriteString(w, string(b))
			return
		}
		if this.raft.State() == raft.Leader {
			this.dmapDelete(k)
			this.broadcastDmapDelete(k)
			b, err := json.Marshal(map[string]string{k: v})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			io.WriteString(w, string(b))
		} else {
			ipport := fmt.Sprintf("%v", this.raft.Leader())
			ip := strings.Split(ipport, ":")[0]
			url := fmt.Sprintf("http://%v:%v/mem/%v", ip, httpPort, k)
			req, err := http.NewRequest("DELETE", url, nil)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			client := &http.Client{}
			resp, err := client.Do(req)
			defer resp.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			//ret := fmt.Sprintf("%v",*resp.Body)
			//io.WriteString(w, ret)
			resp.Write(w)
		}

		/**
		if err := this.Delete(k); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		this.Delete(k)**/

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return

}

func (this *Core) dmapWrite(k, v string) error {
	this.dsm.set(k, v)
	return nil
}

func (this *Core) dmapDelete(k string) error {
	this.dsm.delete(k)
	return nil
}

func (this *Core) broadcastDmapWrite(k, v string) error {
	for _, ip := range this.getOthers() {
		kv := args{}
		kv.K = k
		kv.V = v
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		fmt.Println(addr)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("dmap.Set", kv)
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
	}
	return nil
}

func (this *Core) broadcastDmapDelete(k string) error {
	for _, ip := range this.getOthers() {
		kv := args{}
		kv.K = k
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		fmt.Println(addr)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("dmap.Delete", kv)
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
	}
	return nil
}

func (this *Core) syncDsm() {
	for {
		select {
		case <-time.After(100 * time.Second):
			if this.raft.State() != raft.Leader {
				return
			}
			datax := this.getDmapFromCluster()
			this.syncMap(datax)
		}
	}
}

func (this *Core) countMap(datax map[string]map[string]string) map[string]int {
	counts := make(map[string]int)
	for _, data := range datax {
		for k, _ := range data {
			counts[k] += 1
		}
	}
	return counts
}

func (this *Core) syncMap(datax map[string]map[string]string) {
	counts := make(map[string]int)
	total := len(datax)
	quorum := total/2 + 1
	fmt.Println(quorum)
	bak := make(map[string]string)
	for _, data := range datax {
		for k, v := range data {
			if counts[k] == 0 {
				bak[k] = v
			}
			counts[k] += 1
		}
	}
	for k, n := range counts {
		if n == total {
			continue
		}
		if n < quorum {
			// delete
			for ip, data := range datax {
				if data[k] != "" {
					fmt.Printf("DSM - delete %v@%v\n", k, ip)
					if ip == this.ip {
						this.dsm.delete(k)
					} else {
						kv := args{}
						kv.K = k
						addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
						fmt.Println(addr)
						cli := Client{Adress: addr}
						resBuf, err := cli.SendRequest("dmap.Delete", kv)
						if err != nil {
							log.Fatalln(err)
						}
						ok := ResultIsOk(resBuf)
						fmt.Println(ok)
					}
				}
			}
			continue
		}
		// n >= quorum, rewrite
		for ip, data := range datax {
			if data[k] == "" {
				fmt.Printf("DSM - set %v@%v to %v\n", k, ip, bak[k])
				if ip == this.ip {
					this.dsm.set(k, bak[k])
				} else {
					kv := args{}
					kv.K = k
					kv.V = bak[k]
					addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
					fmt.Println(addr)
					cli := Client{Adress: addr}
					resBuf, err := cli.SendRequest("dmap.Set", kv)
					if err != nil {
						log.Fatalln(err)
					}
					ok := ResultIsOk(resBuf)
					fmt.Println(ok)
				}
			}
		}
	}
}

func (this *Core) getDmapFromCluster() map[string]map[string]string {
	datax := make(map[string]map[string]string)
	datax[this.ip] = this.dsm.data
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("dmap.GetMap", args{})
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
		data := make(map[string]string)
		ConvertToNormalType(resBuf, &data)
		datax[ip] = data
	}
	return datax
}
