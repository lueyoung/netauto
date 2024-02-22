package core

func trueInMap(m map[string]bool) bool {
	res := false
	for _, i := range m {
		if res = res || i; res {
			return res
		}
	}
	return res
}
