package core

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// config
type Config struct {
	Network   string
	Join      string
	Force     bool
	Maxmem    int
	Vagrant   bool
	Nosystemd bool
	Container bool
}

// command
type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type status struct {
	TotalCount int
	Members    *[]staff
}

type staff struct {
	Name string
	Addr string
	Role string
}

type query struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type resultxs struct {
	Results []resultx `json:"ret"`
}

type resultx struct {
	Result string `json:"log"`
	Date   string `json:"created"`
	Key    string `json:"timestamp"`
}

type dsm struct {
	A string `json:"a"`
	B string `json:"b"`
	C string `json:"c"`
}

// rpc
type args struct {
	K string `json:"k"`
	V string `json:"v"`
}

type xxx struct {
	A, B int
	Name string
}
type Args struct {
	A, B int
}

func (x *xxx) Add(input Args) int {
	return input.A + input.B
}

type yyy struct {
	val int
}

func (y *yyy) Inc(val int) int {
	y.val = y.val + val
	return y.val
}
func (y *yyy) Set(val int) string {
	y.val = val
	return ""
}

func (m *dsm) Set(input args) error {
	k := input.K
	v := input.V
	r := reflect.ValueOf(m).Elem()
	r.FieldByName(strings.ToUpper(k)).Set(reflect.ValueOf(v))
	//r0 := reflect.ValueOf(this.mem).Elem()
	//v0 := r0.FieldByName(strings.ToUpper(k)).String()
	//r0.FieldByName(strings.ToUpper(k)).Set(reflect.ValueOf(v))
	//r1 := reflect.ValueOf(this.mem).Elem()
	//v1 := r1.FieldByName(strings.ToUpper(k)).String()
	//fmt.Printf("previoous: %v, current: %v\n", v0, v1)
	return nil
}

// for rpc
type funcs struct{}

// state
type systeminfo struct {
	// uname -a
	SystemInfo string `json:"systemInfo"`
	// cat /proc/cpuinfo | grep "physical id" | sort | uniq | wc –l
	PhysicalCpuNum string `json:"physicalCpuNum"`
	// cat /proc/cpuinfo | grep "processor" | wc –l
	ProcessorNum string `json:"processorNum"`
	// cat /proc/cpuinfo | grep "cores" | uniq
	Cores string `json:"cores"`
	// cat /proc/meminfo | grep "MemTotal:" | uniq
	MemTotal string `json:"memTotal"`
	// cat /proc/meminfo | grep "MemFree:" | uniq
	MemFree string `json:"memFree"`
}

// for kubenetes
type kube struct {
	role      string
	Install   bool `json:"install"`
	installed bool
	OK        bool   `json:"ok"`
	Error     string `json:"error"`
	Typed     string `json:"typed"`
}

func (k *kube) SetRole(input args) error {
	k.role = input.V
	return nil
}

func (k *kube) GetTyped(a args) string {
	return k.Typed
}

func (k *kube) GetRole(input args) string {
	return k.role
}

type file struct {
	Content string `json:"content"`
	Path    string `json:"path"`
	hash    string
}

func (this *file) set(path, content string) {
	this.Path = path
	this.Content = content
}

func (this *file) get() (string, string) {
	return this.Path, this.Content
}

func append2file(path, content string) error {
	fn, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		msg := fmt.Sprintf("Failed to open the file: %v, error: %v", path, err.Error())
		log.Fatal(msg)
	}
	defer fn.Close()
	if _, err := fn.Write([]byte("\n")); err != nil {
		log.Fatal(err)
	}
	if _, err := fn.Write([]byte(content)); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (f *file) append() error {
	path, content := f.get()
	return append2file(path, content)
}

func containCheck(b []byte, sub string) bool {
	//return strings.Contains(string(b), sub)
	pattern := fmt.Sprintf("%v", sub)
	res, err := regexp.Match(pattern, b)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func (f *file) ifWritten() bool {
	path, content := f.get()
	// 0 make hash
	f.hash = makeHash(content)
	// 1 read
	fn, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fn.Close()
	b, err := ioutil.ReadAll(fn)
	if err != nil {
		log.Fatal(err)
	}
	return containCheck(b, fmt.Sprintf("%v-%v", string(magic), f.hash))
}

func makeHash(data string) string {
	h := sha1.New()
	io.WriteString(h, data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (f *file) magic() error {
	path, _ := f.get()
	//return append2file(path, string(magic))
	return append2file(path, fmt.Sprintf("%v-%v", string(magic), f.hash))
}

func (f *file) try2append() error {
	if f.ifWritten() {
		return nil
	}
	err := f.magic()
	if err != nil {
		log.Fatal(err)
	}
	err = f.append()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (f *file) new() error {
	f.deleteIfExisted()
	path, content := f.get()
	fn, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		msg := fmt.Sprintf("Failed to open the file: %v, error: %v", path, err.Error())
		log.Fatal(msg)
	}
	defer fn.Close()
	if _, err := fn.Write([]byte(content)); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (f *file) newFile() error {
	return f.new()
}

func (f *file) newExec() error {
	f.newFile()
	path, _ := f.get()
	cmd := fmt.Sprintf("chmod a+x %v", path)
	_, err := execCmd(cmd)
	return err
}

func (f *file) deleteIfExisted() error {
	path, _ := f.get()
	if checkFileExist(path) {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *funcs) Copy(f file) error {
	path, content := f.get()
	fn, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		msg := fmt.Sprintf("Failed to open the file: %v, error: %v", path, err.Error())
		log.Fatal(msg)
	}
	defer fn.Close()
	if _, err := fn.Write([]byte(content)); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (this *funcs) Append(f file) error {
	return f.append()
}

func (this *funcs) IfExist(f file) bool {
	path, _ := f.get()
	return checkFileExist(path)
}

func (this *funcs) Try2Append(f file) error {
	return f.try2append()
}

func (this *funcs) New(f file) error {
	return f.new()
}

func (this *funcs) NewFile(f file) error {
	return f.newFile()
}

func (this *funcs) NewExec(f file) error {
	return f.newExec()
}

type linuxcmd struct {
	Cmd  string `json:"cmd"`
	Args string `json:"args"`
}

func (this *linuxcmd) get() (string, string) {
	return this.Cmd, this.Args
}

func (this *linuxcmd) set(cmd, arg string) error {
	this.Cmd = cmd
	this.Args = arg
	return nil
}

func (this *linuxcmd) run() (string, error) {
	cmd := this.getCmd()
	return execCmd(cmd)
}
func (this *linuxcmd) getCmd() string {
	cmd, arg := this.get()
	if arg != "" {
		return fmt.Sprintf("%v %v", cmd, arg)
	}
	return cmd
}
func (this *linuxcmd) runLogged() (string, error) {
	cmd := this.getCmd()
	return execCmdLogged(cmd)
}

func (this *funcs) Run(cmd linuxcmd) (string, error) {
	return cmd.run()
}
func (this *funcs) RunLogged(cmd linuxcmd) (string, error) {
	return cmd.runLogged()
}

func (this *funcs) ConfigNtpClient(f file) error {
	path, server := f.get()
	conf := ntpClientConf
	conf = strings.Replace(conf, "{{.ntp.server}}", server, -1)
	f.set(path, conf)
	return f.newFile()
}

type kubeclusterinfo struct {
	Num    int      `json:"num"`
	Master []string `json:"master"`
	Node   []string `json:"node"`
	VIP    string   `json:"vip"`
}

func (this *kubeclusterinfo) Copy(k kubeclusterinfo) error {
	this.Master = k.Master
	this.Node = k.Node
	this.Num = k.Num
	this.VIP = k.VIP
	return nil
}

func (this *kubeclusterinfo) getMaster() []string {
	return sortString(this.Master)
}

func (this *kubeclusterinfo) getNode() []string {
	return sortString(this.Node)
}

func (this *kubeclusterinfo) getVIP() string {
	return this.VIP
}
