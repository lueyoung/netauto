package core

import (
	"encoding/json"
	"fmt"
)

type delegate struct{}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}
	//fmt.Printf("%v %v\n", b[0], 'i')
	//fmt.Println(b)

	//str := string(b)
	switch b[0] {
	case 'i': //introduce
		fmt.Println(b)
		d.handleIntroduce(b)
	case 'd': //dsm sync
		fmt.Println(b)
		d.handleDsmSync(b)
	case 'r': //report
		fmt.Println(b)
		d.handleReport(b)
	}
}

func (d *delegate) handleIntroduce(b []byte) {
	c := b[2:]
	id <- c
}

func (d *delegate) handleDsmSync(b []byte) {
	c := b[2:]
	dsmsync <- c
}

func (d *delegate) handleReport(b []byte) {
	c := b[2:]
	report <- c
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	mtx.RLock()
	m := items
	mtx.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	mtx.Lock()
	for k, v := range m {
		items[k] = v
	}
	mtx.Unlock()
}
