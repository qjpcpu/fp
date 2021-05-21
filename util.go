package fp

import "reflect"

var (
	boolType = reflect.TypeOf(true)
	errType  = reflect.TypeOf((*error)(nil)).Elem()
)
