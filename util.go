package fp

import "reflect"

var (
	boolType = reflect.TypeOf(true)
	errType  = reflect.TypeOf((*error)(nil)).Elem()
)

func NoError() func(error) bool {
	return func(e error) bool {
		return e == nil
	}
}
