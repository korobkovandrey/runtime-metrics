package adapter

type RepositoryKey string

type Repository interface {
	AddType(t string)
	Set(t string, name string, value any)
	IncrInt64(t string, name string, value int64)
	GetStorageData() interface{}
}
