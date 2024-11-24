package memstorage

type float64Store map[string]map[string]float64

func newFloat64Store(types ...string) float64Store {
	s := float64Store{}
	for _, t := range types {
		s[t] = map[string]float64{}
	}
	return s
}

func (s float64Store) set(t string, name string, value float64) {
	s[t][name] = value
}
func (s float64Store) incr(t string, name string, value float64) {
	s[t][name] += value
}
func (s float64Store) get(t string, name string) (value float64, ok bool) {
	value, ok = s[t][name]
	return
}
