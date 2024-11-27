package memstorage

type float64Store struct {
	*store
}

func newFloat64Store(types ...string) *float64Store {
	return &float64Store{
		newStore(types...),
	}
}
