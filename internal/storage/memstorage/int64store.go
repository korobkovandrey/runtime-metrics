package memstorage

type int64Store map[string]map[string]int64

func newInt64Store(types ...string) int64Store {
	s := int64Store{}
	for _, t := range types {
		s[t] = map[string]int64{}
	}
	return s
}

func (s int64Store) set(t string, name string, value int64) {
	s[t][name] = value
}
func (s int64Store) incr(t string, name string, value int64) {
	s[t][name] += value
}
func (s int64Store) get(t string, name string) (value int64, ok bool) {
	value, ok = s[t][name]
	return
}
