package memstorage

const (
	typeGauge   = `gauge`
	typeCounter = `counter`
)

type MemStorage struct {
	*float64Store
	*int64Store
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		newFloat64Store(typeGauge),
		newInt64Store(typeCounter),
	}
}

func (m MemStorage) SetGauge(name string, value float64) {
	m.float64Store.set(typeGauge, name, value)
}

func (m MemStorage) GetGauge(name string) (value float64, ok bool) {
	v, ok := m.float64Store.get(typeCounter, name)
	value, _ = v.(float64)
	return
}

func (m MemStorage) IncrCounter(name string, value int64) {
	m.int64Store.incrInt64(typeCounter, name, value)
}

func (m MemStorage) GetCounter(name string) (value int64, ok bool) {
	v, ok := m.int64Store.get(typeCounter, name)
	value, _ = v.(int64)
	return
}

func (m MemStorage) GetStorageData() interface{} {
	return map[string]interface{}{
		`float64Store`: m.float64Store.getData(),
		`int64Store`:   m.int64Store.getData(),
	}
}
