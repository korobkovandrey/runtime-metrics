package adapter

type RepositoryKey string

type Repository interface {
	AddType(t string)
	Get(t string, name string) (value any, ok bool)
	Set(t string, name string, value any)
	IncrInt64(t string, name string, value int64)
}
