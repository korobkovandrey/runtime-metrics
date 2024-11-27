package adapter

type Repository interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (value float64, ok bool)
	IncrCounter(name string, value int64)
	GetCounter(name string) (value int64, ok bool)
	GetStorageData() interface{}
}
