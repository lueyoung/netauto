package core

import (
	"fmt"
	//"io/ioutil"
	"log"
	//"math/rand"
	//"os/exec"
	"strconv"
	"strings"
)

func removeSymbol(str string) string {
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	//str = strings.Replace(str, " ", "", -1)
	return str
}

func makeSystemInfo() *systeminfo {
	s := new(systeminfo)
	var tmp string
	//s. = execCommand("")
	// 0
	tmp = execCommand("uname -a")
	tmp = removeSymbol(tmp)
	s.SystemInfo = tmp
	// 1
	tmp = execCommand("cat /proc/cpuinfo | grep 'physical id' | sort | uniq | wc -l")
	tmp = removeSymbol(tmp)
	s.PhysicalCpuNum = tmp
	// 2
	tmp = execCommand("cat /proc/cpuinfo | grep 'processor' | wc -l")
	tmp = removeSymbol(tmp)
	s.ProcessorNum = tmp
	// 3
	tmp = execCommand("cat /proc/cpuinfo | grep 'cores' | uniq")
	tmp = removeSymbol(tmp)
	tmp = strings.Split(tmp, ":")[1]
	tmp = strings.Trim(tmp, " ")
	s.Cores = tmp
	// 4
	tmp = execCommand("cat /proc/meminfo | grep 'MemTotal:' | uniq")
	tmp = removeSymbol(tmp)
	tmp = strings.Split(tmp, ":")[1]
	tmp = strings.Trim(tmp, " ")
	s.MemTotal = tmp
	// 5
	tmp = execCommand("cat /proc/meminfo | grep 'MemFree:' | uniq")
	tmp = removeSymbol(tmp)
	tmp = strings.Split(tmp, ":")[1]
	tmp = strings.Trim(tmp, " ")
	s.MemFree = tmp

	return s
}

func (this *funcs) GetStatus(in args) *systeminfo {
	//func (this *funcs) GetStatus(in args) string {
	s := makeSystemInfo()
	//return fmt.Sprintf("%v", *s)
	return s
}

func (this *funcs) Add(input Args) int {
	return input.A + input.B

}

func score() float64 {
	s := makeSystemInfo()
	// cpu
	logicalCpuNum, err := strconv.ParseFloat(s.ProcessorNum, 64)
	if err != nil {
		log.Fatal(err)
	}
	core, err := strconv.ParseFloat(s.Cores, 64)
	if err != nil {
		log.Fatal(err)
	}
	cores := core * logicalCpuNum
	// mem
	memTotalStr := strings.Split(s.MemTotal, " ")[0]
	memTotal, err := strconv.ParseFloat(strings.Trim(memTotalStr, " "), 64)
	if err != nil {
		log.Fatal(err)
	}

	return 1.1*cores + 0.9*memTotal/1000/1000/2
	//return rand.Float64() + 1.1*cores + 0.9*memTotal/1000/1000/2
}

func (this *funcs) Score(input args) float64 {
	return score()
}

func (this *Core) getScores() map[string]float64 {
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
	return ret
}
