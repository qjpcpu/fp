package fp

import "reflect"

func (q *stream) Reject(fn interface{}) Stream {
	typ := reflect.TypeOf(fn)
	val := reflect.ValueOf(fn)
	fnr := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		b := val.Call(in)[0].Bool()
		return []reflect.Value{reflect.ValueOf(!b)}
	})
	return q.Filter(fnr.Interface())
}
