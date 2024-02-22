package core

var (
	id           = make(chan []byte)
	dsmsync      = make(chan []byte)
	report       = make(chan []byte)
	newbie       = make(chan string)
	addraft      = make(chan string)
	maintainraft = make(chan string)
	installkube  = make(chan string)
	log2kubelog  = make(chan string)
)
