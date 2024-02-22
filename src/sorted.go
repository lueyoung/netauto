package core

import (
	"sort"
)

// A data structure to hold a key/value pair.
type Pair struct {
	Key   string
	Value float64
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int      { return len(p) }

//func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function to turn a map into a PairList, then sort and return it.
// from bigger to smaller
func sortMapByValue(m map[string]float64) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	return p
}

func sortString(strs []string) []string {
	if len(strs) <= 1 {
		return strs
	}
	var sorted sort.StringSlice
	for _, str := range strs {
		sorted = append(sorted, str)
	}
	sort.Sort(sorted)
	return sorted
}
