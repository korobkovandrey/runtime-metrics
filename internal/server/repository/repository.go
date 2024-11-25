package repository

type Type string

const (
	gaugeType   Type = `gauge`
	counterType Type = `counter`
)

type Repository interface {
	GetType() Type
	Update(name string, value string) error
	GetStorage() interface{}
}
