package core

import (
	"testing"
)

func Test_0(t *testing.T) {
	var str []string
	t.Log(sortString(str))
}

func Test_1(t *testing.T) {
	str := []string{"192.168.100.81"}
	t.Log(sortString(str))
}

func Test_2(t *testing.T) {
	str := []string{"192.168.100.82", "192.168.100.81"}
	t.Log(sortString(str))
}

func Test_3(t *testing.T) {
	str := []string{"192.168.100.82", "192.168.100.84", "192.168.100.81"}
	t.Log(sortString(str))
}

func Test_4(t *testing.T) {
	str := []string{"192.168.100.82", "192.168.100.84", "192.168.100.81", "192.168.100.83"}
	t.Log(sortString(str))
}
