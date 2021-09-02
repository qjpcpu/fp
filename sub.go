package fp

import (
	"reflect"
	"sync"
)

func (q *stream) Sub(other Stream) Stream {
	if isNilStream(other) {
		return q
	}
	var once sync.Once
	var set KVStream
	getSet := func() KVStream {
		once.Do(func() {
			set = other.ToSet()
		})
		return set
	}
	typ := reflect.FuncOf([]reflect.Type{q.expectElemTyp}, []reflect.Type{boolType}, false)
	fn := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(getSet().Contains(in[0].Interface()))}
	})
	return q.Reject(fn.Interface())
}

func (q *stream) SubBy(other Stream, keyfn interface{}) Stream {
	if isNilStream(other) {
		return q
	}
	var once sync.Once
	var set KVStream
	keyfnval := reflect.ValueOf(keyfn)
	getSet := func() KVStream {
		once.Do(func() {
			set = other.ToSetBy(keyfn)
		})
		return set
	}
	typ := reflect.FuncOf([]reflect.Type{q.expectElemTyp}, []reflect.Type{boolType}, false)
	fn := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		key := keyfnval.Call([]reflect.Value{in[0]})[0].Interface()
		return []reflect.Value{reflect.ValueOf(getSet().Contains(key))}
	})
	return q.Reject(fn.Interface())
}
