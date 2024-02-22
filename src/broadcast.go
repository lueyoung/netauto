package core

import (
	"memberlist"
)

type broadcast struct {
	msg    []byte
	notify chan struct{}
}

func (this *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (this *broadcast) Message() []byte {
	return this.msg
}

func (this *broadcast) Finished() {
	/*
		select {
		case this.notify <- struct{}{}:
		default:
		}
	*/
	if this.notify != nil {
		close(this.notify)
	}
}
