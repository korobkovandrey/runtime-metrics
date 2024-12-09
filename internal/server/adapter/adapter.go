package adapter

type Repository interface {
	AddType(t string)
	Keys(t string) []string
	Get(t string, name string) (value any, ok bool)
	Set(t string, name string, value any)
	IncrInt64(t string, name string, value int64)
}
