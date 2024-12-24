package adapter

type Repository interface {
	AddType(t string)
	Keys(t string) []string
	SetFloat64(t string, name string, value float64)
	SetInt64(t string, name string, value int64)
	IncrInt64(t string, name string, value int64)
	Get(t string, name string) (value any, ok bool)
	GetFloat64(t string, name string) (value float64, ok bool)
	GetInt64(t string, name string) (value int64, ok bool)
}
