package storage

type Metric interface {
	Handler(name string, value string) error
}

type Storage interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (value float64, ok bool)
	IncrCounter(name string, value int64)
	GetCounter(name string) (value int64, ok bool)
}
