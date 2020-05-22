package store

type DB interface {
	Get(key string) (value interface{}, err error)
	Set(key string, value interface{}) (err error)

	GetAll() <-chan interface{}
}
