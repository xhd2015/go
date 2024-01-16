package common

var ImportsKeepOrder bool

func Deduplicate(s []string) []string {
	set := make(map[string]bool)
	res := make([]string, 0, len(s))
	for _, k := range s {
		if set[k] {
			continue
		}
		set[k] = true
		res = append(res, k)
	}
	return res
}
