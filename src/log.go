package core

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func (this *Core) handleLogRequest(w http.ResponseWriter, r *http.Request) {
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
		if v == "" {
			v = fmt.Sprintf("not found: %v", k)
			k = "get error"
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
				tm := time.Now()
				k1 := fmt.Sprintf("%v", tm.Unix())
				ftm := tm.Format("2006-01-02 15:04:05")
				level := strings.ToUpper(k)
				v1 := fmt.Sprintf("%v - %v - %v", ftm, level, v)
				if err := this.Set(k1, v1); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				b, err := json.Marshal(map[string]string{k1: v1})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				io.WriteString(w, string(b))
			} else {
				ipport := fmt.Sprintf("%v", this.raft.Leader())
				ip := strings.Split(ipport, ":")[0]
				url := fmt.Sprintf("http://%v:%v/log", ip, httpPort)
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
		err := this.Delete(k)
		ret := make(map[string]string)
		if err != nil {
			ret["delete error"] = fmt.Sprintf("%v", err)
		} else {
			ret["delete success"] = fmt.Sprintf("%v", k)
		}
		b, err := json.Marshal(ret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(b))

	case "PUT":
		earliest := "2019-01-01 00:00:00"
		now := time.Now().Unix()
		q := query{}
		ret := resultxs{}
		if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var from, to int64
		fromStr := q.From
		if fromStr == "" {
			from = getTimeStamp(earliest)
		} else {
			from = getTimeStamp(fromStr)
		}
		toStr := q.To
		if toStr == "" {
			to = now
		} else {
			to = getTimeStamp(toStr)
		}
		//ret := query{}
		//ret.From = fmt.Sprintf("%v",from)
		//ret.To = fmt.Sprintf("%v",to)

		for k := from; k < to+1; k++ {
			v, err := this.Get(fmt.Sprintf("%v", k))
			if err == nil {
				if v != "" {
					r := resultx{}
					r.Key = fmt.Sprintf("%v", k)
					r.Result = v
					tm := time.Unix(k, 0)
					t := fmt.Sprintf("%v", tm.Format("2006-01-02 15:04:05"))
					r.Date = t
					ret.Results = append(ret.Results, r)
				}
			}
		}

		b, err := json.Marshal(ret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(b)

		io.WriteString(w, string(b))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func getTimeStamp(date string) int64 {
	loc, _ := time.LoadLocation("Local")
	reg, _ := regexp.Compile("[0-9]+")
	all := reg.FindAll([]byte(date), 6)
	l := len(all)
	if l < 1 {
		log.Println("check input time format")
		return -1
	}
	year := string(all[0])
	month := "01"
	day := "01"
	hour := "00"
	minute := "00"
	second := "00"
	if l > 1 {
		month = string(all[1])
	}
	if l > 2 {
		day = string(all[2])
	}
	if l > 3 {
		hour = string(all[3])
	}
	if l > 4 {
		minute = string(all[4])
	}
	if l > 5 {
		second = string(all[5])

	}
	//fmt.Printf("%v-%v-%v %v:%v:%v\n", year, month, day, hour, minute, second)
	str := fmt.Sprintf("%v-%v-%v %v:%v:%v", year, month, day, hour, minute, second)
	dt, _ := time.ParseInLocation("2006-01-02 15:04:05", str, loc)
	ts := dt.Unix()
	//fmt.Println(ts)
	return ts
}
