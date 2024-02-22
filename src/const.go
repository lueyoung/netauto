package core

import (
	//"html/template"
	"time"
)

const (
	raftPort             = 12345
	httpPort             = 18080
	enableSingle         = true
	retainSnapshotCount  = 2
	raftTimeout          = 10 * time.Second
	raftMaintainTotalTry = 5
	kubelog              = "/tmp/kube.log"
	kubeMasterPort       = 443
	realKubeMasterPort   = 6443
	magic                = "###{{.for.magic.usage}}"
	bin                  = "/usr/local/bin"
	cfsslVersion         = "R1.2"
	kubeVersion          = "v1.12.8"
	etcdVersion          = "v3.3.13"
	routerID             = "kube_api_007"
	virtualRouterID      = 57
	dockerVersion        = "18.03.1-ce"
	vagrantIP            = `10.0.2.15`
	projectName          = "dynas"
	pidpath              = "/var/run"
	kubeneeded           = 4
	clusterCidr          = "172.30.0.0/16"
	scanBatch            = 100
)

const templ = `
<html>
<body>
<h2>{{.TotalCount}} members</h2>
<table>
<tr style='text-align: left'>
  <th>Name</th>
  <th>Addr</th>
  <th>Role</th>
</tr>
{{range .Members}}
<tr style='text-align: left'>
  <th>{{.Name}}</th>
  <th>{{.Addr}}</th>
  <th>{{.Role}}</th>
</tr>
{{end}}
</table>
</body>
</html>
`

const templ0 = `<h1> hello world </h1>`

const sourceEnv = `
FILES=$(find /etc/kubernetes/env -name "*.env")
if [ -n "$FILES" ]
then
  for FILE in $FILES
  do
    [ -f $FILE ] && source $FILE
  done
fi
`
