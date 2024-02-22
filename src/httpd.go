package core

import (
	"net/http"
	"strings"
)

func (this *Core) Close() {
	this.ln.Close()
	return
}

func (this *Core) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 底层数据库rest接口
	if strings.HasPrefix(r.URL.Path, "/key") {
		this.handleKeyRequest(w, r)
		// 分布式日志系统rest接口
	} else if strings.HasPrefix(r.URL.Path, "/log") {
		this.handleLogRequest(w, r)
		// todo
	} else if strings.HasPrefix(r.URL.Path, "/dsm") {
		this.handleDsmRequest(w, r)
		// 分布式共享内存rest接口
	} else if strings.HasPrefix(r.URL.Path, "/mem") {
		this.handleDmapRequest(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/countmem") {
		this.handleCountDmapRequest(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/allmem") {
		this.handleAllDmapRequest(w, r)
		// rpc接口
	} else if r.URL.Path == "/v1/api" {
		this.rpc.ServeHTTP(w, r)
		// 返回集群所有IP
	} else if r.URL.Path == "/ips" {
		this.handleIPRequest(w, r)
		// 返回Raft网络的Leader
	} else if r.URL.Path == "/leader" {
		this.handleLeader(w, r)
		// 返回网络的简报
	} else if r.URL.Path == "/status" {
		this.handleStatus(w, r)
	} else if r.URL.Path == "/memmax" {
		this.handleDsmMax(w, r)
	} else if r.URL.Path == "/memsize" {
		this.handleDsmSize(w, r)
	} else if r.URL.Path == "/test" {
		this.handleTest(w, r)
	} else if r.URL.Path == "/test0" {
		this.handleTest0(w, r)
	} else if r.URL.Path == "/test00" {
		this.handleTest00(w, r)
		// 返回系统信息
	} else if r.URL.Path == "/systeminfo" {
		this.handleSystemInfo(w, r)
		// 返回节点的评分
	} else if r.URL.Path == "/scores" {
		this.handleScores(w, r)
	} else if r.URL.Path == "/kube" {
		this.handleKubeRequest(w, r)
		// 返回Kubernetes集群状态
	} else if r.URL.Path == "/kube-status" {
		this.handleKubeStatusRequest(w, r)
	} else if r.URL.Path == "/kube-fix" {
		this.handleKubeFixRequest(w, r)
	} else if r.URL.Path == "/kube-exist" {
		this.handleKubeExistRequest(w, r)
		// 返回节点类型
	} else if r.URL.Path == "/typed" {
		this.handleTypedRequest(w, r)
		// 退出程序
	} else if r.URL.Path == "/exit" {
		handleExitRequest(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
