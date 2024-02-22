package core

type update struct {
	Action string // add, del
	Data   map[string]string
}
