package contracts

type Container interface {
	Instances() []interface{}
	Singleton(resolver interface{})
	Transient(resolver interface{})
	Make(receiver interface{}) []interface{}
}
