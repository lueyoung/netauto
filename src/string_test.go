package core

import (
	"testing"
)

func Test_MakefileInc(t *testing.T) {
	str := string(calicoMakefileInc)
	str += "ETCD_KEY_PEM=`cat ${SSL}/etcd-key.pem | base64 | tr -d '\\n'`\n"
	str += "ETCD_PEM=`cat ${SSL}/etcd.pem | base64 | tr -d '\\n'`\n"
	str += "CA_PEM=`cat ${SSL}/ca.pem | base64 | tr -d '\\n'`\n"
	t.Log(str)
}
