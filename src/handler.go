package core

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"html/template"
	"io"
	"log"
	"os"
	//"net/http"
	"io/ioutil"
	"net/http"
	"strings"
)

func (this *Core) handleStatus(w http.ResponseWriter, r *http.Request) {
	n := len(this.list.Members())
	ret := status{
		TotalCount: n,
	}
	var members []staff
	leader := fmt.Sprintf("%v", this.raft.Leader())
	for _, member := range this.list.Members() {
		//fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
		i := staff{
			Name: member.Name,
			Addr: fmt.Sprintf("%v", member.Addr),
		}
		if strings.Contains(leader, i.Addr) {
			i.Role = "Leader"
		} else {
			i.Role = "Follower"
		}
		members = append(members, i)
	}
	ret.Members = &members
	t := template.Must(template.New("demo").Parse(templ))
	err := t.Execute(w, ret)
	//err := t.ExecuteTemplate(w, templ, ret)
	if err != nil {
		log.Fatal(err)
	}
}

func (this *Core) handleLeader(w http.ResponseWriter, r *http.Request) {
	raftportaddr := fmt.Sprintf("%v", this.raft.Leader())
	ipport := strings.Split(raftportaddr, ":")
	ip := ipport[0]
	leader := fmt.Sprintf("%v:%v", ip, httpPort)
	b, err := json.Marshal(map[string]string{"Leader": leader})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}

func (this *Core) handleKeyRequest(w http.ResponseWriter, r *http.Request) {
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
		v, err := this.Get(k)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(map[string]string{k: v})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(b))

	case "POST":
		// Read the value from the POST body.
		m := map[string]string{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for k, v := range m {
			if this.raft.State() == raft.Leader {
				if err := this.Set(k, v); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				ipport := fmt.Sprintf("%v", this.raft.Leader())
				ip := strings.Split(ipport, ":")[0]
				url := fmt.Sprintf("http://%v:%v/key", ip, httpPort)
				j := fmt.Sprintf("{\"%v\":\"%v\"}", k, v)
				typed := "application/json"
				fmt.Printf("redirect url: %v\n", url)
				fmt.Printf("redirect json: %v\n", j)
				client := &http.Client{}
				_, err := client.Post(url, typed, strings.NewReader(j))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

	case "DELETE":
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := this.Delete(k); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		this.Delete(k)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func (this *Core) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	// self
	ret := make(map[string]systeminfo)
	tmp := new(systeminfo)
	tmp = makeSystemInfo()
	//fmt.Println(*tmp)
	//res[this.ip] = systeminfo{}
	ret[this.ip] = *tmp
	//fmt.Println(ret)

	// others
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("funcs.GetStatus", args{})
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
		res := new(systeminfo)
		ConvertToNormalType(resBuf, res)
		ret[ip] = *res
	}

	b, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}

func (this *Core) handleScores(w http.ResponseWriter, r *http.Request) {
	// self
	ret := make(map[string]float64)
	ret[this.ip] = score()

	// others
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("funcs.Score", args{})
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
		res := new(float64)
		ConvertToNormalType(resBuf, res)
		ret[ip] = *res
	}

	b, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}

func (this *Core) handleKubeRequest(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		k := "foo"
		v := "bar"
		b, err := json.Marshal(map[string]string{k: v})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(b))

	case "PUT", "POST":
		// Read the value from the POST body.
		k := new(kube)
		j, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		err = json.Unmarshal(j, k)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if this.raft.State() == raft.Leader {
			//this.installKube()
			ok := true
			if err != nil {
				ok = false
				k.Error = fmt.Sprintf("%v", err)
			}
			k.OK = ok
			b, err := json.Marshal(k)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			io.WriteString(w, string(b))
		} else {
			ipport := fmt.Sprintf("%v", this.raft.Leader())
			ip := strings.Split(ipport, ":")[0]
			url := fmt.Sprintf("http://%v:%v/kube", ip, httpPort)
			typed := "application/json"
			fmt.Printf("redirect url: %v\n", url)
			fmt.Printf("redirect json: %v\n", j)
			client := &http.Client{}
			resp, err := client.Post(url, typed, strings.NewReader(string(j)))
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

func (this *Core) handleKubeStatusRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ret, _, _ := this.reportKubeStatus()
		b, err := json.Marshal(ret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(b))
	}
}

func (this *Core) handleKubeExistRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _, ret := this.reportKubeStatus()
		io.WriteString(w, fmt.Sprintf("%v", ret))
	}
}

func (this *Core) handleKubeFixRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		this.fixKube()
		ret := "restart failed services, wait a moment, and check the status page again"
		io.WriteString(w, fmt.Sprintf("%v", ret))
	}
}

func (this *Core) handleTypedRequest(w http.ResponseWriter, r *http.Request) {
	// self
	ret := make(map[string]string)
	//fmt.Println(*tmp)
	//res[this.ip] = systeminfo{}
	ret[this.ip] = this.getTyped()
	//fmt.Println(ret)

	// others
	for _, ip := range this.getOthers() {
		addr := fmt.Sprintf("http://%v:%v/v1/api", ip, httpPort)
		cli := Client{Adress: addr}
		resBuf, err := cli.SendRequest("kube.GetTyped", args{})
		if err != nil {
			log.Fatalln(err)
		}
		ok := ResultIsOk(resBuf)
		fmt.Println(ok)
		res := new(string)
		ConvertToNormalType(resBuf, res)
		ret[ip] = *res
	}

	b, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}

func handleExitRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ret := make(map[string]string)
		ret["PodIP"] = os.Getenv("POD_IP")
		b, err := json.Marshal(ret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(b))
		os.Exit(1)
	}
}

func (this *Core) handleCountDmapRequest(w http.ResponseWriter, r *http.Request) {
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

func (this *Core) handleIPRequest(w http.ResponseWriter, r *http.Request) {
	ret := make(map[string]string)
	ips := this.getAllMembers()
	str := ""
	sep := ""
	for _, ip := range ips {
		str += sep
		str += ip
		sep = ","
	}
	ret["ret"] = str
	b, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}
