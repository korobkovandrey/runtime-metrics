package memstorage

type int64Store struct {
	*store
}

func newInt64Store(types ...string) *int64Store {
	return &int64Store{
		newStore(types...),
	}
}
