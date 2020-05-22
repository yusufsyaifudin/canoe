package repo

type Service interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
}
