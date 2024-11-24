package memstorage

const (
	typeGauge   = `gauge`
	typeCounter = `counter`
)

type MemStorage struct {
	float64Store
	int64Store
}

func NewMemStorage() MemStorage {
	return MemStorage{
		newFloat64Store(typeGauge),
		newInt64Store(typeCounter),
	}
}

func (m MemStorage) SetGauge(name string, value float64) {
	m.float64Store.set(typeGauge, name, value)
}

func (m MemStorage) GetGauge(name string) (value float64, ok bool) {
	return m.float64Store.get(typeGauge, name)
}

func (m MemStorage) IncrCounter(name string, value int64) {
	m.int64Store.incr(typeCounter, name, value)
}

func (m MemStorage) GetCounter(name string) (value int64, ok bool) {
	return m.int64Store.get(typeCounter, name)
}

func (m MemStorage) Get() MemStorage {
	return m
}
