package fp

import "reflect"

var (
	boolType = reflect.TypeOf(true)
	errType  = reflect.TypeOf((*error)(nil)).Elem()
)

func funcInputs(typ reflect.Type) (list []reflect.Type) {
	list = make([]reflect.Type, 0, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		list = append(list, typ.In(i))
	}
	return
}

func funcOutputs(typ reflect.Type) (list []reflect.Type) {
	list = make([]reflect.Type, 0, typ.NumOut())
	for i := 0; i < typ.NumOut(); i++ {
		list = append(list, typ.Out(i))
	}
	return
}
