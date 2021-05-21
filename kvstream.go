package fp

import (
	"reflect"
)

type KVStream interface {
	// Foreach element of object
	// fn should be func(key_type,element_type)
	Foreach(fn interface{}) KVStream
	// Map k-v pair
	// fn should be func(key_type,element_type) (any_type,any_type)
	Map(fn interface{}) KVStream
	// MapValue fn should be func(optional[key_type],element_type) (any_type)
	MapValue(fn interface{}) KVStream
	// MapKey fn should be func(key_type,optional[element_type]) (any_type)
	MapKey(fn interface{}) KVStream
	// Filter kv pair
	Filter(fn interface{}) KVStream
	// Reject kv pair
	Reject(fn interface{}) KVStream
	// Contains key
	Contains(key interface{}) bool
	Keys() Stream
	Values() Stream
	Size() int
	Result() interface{}
}

type kvStream struct {
	mapVal           reflect.Value
	keyType, valType reflect.Type
}

func KVStreamOf(m interface{}) KVStream {
	if reflect.TypeOf(m).Kind() != reflect.Map {
		panic("argument should be map")
	}
	obj := newKvList()
	tp := reflect.TypeOf(m)
	obj.mapVal = reflect.ValueOf(m)
	obj.keyType = tp.Key()
	obj.valType = tp.Elem()
	return obj
}

func (obj *kvStream) Foreach(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	iter := obj.mapVal.MapRange()
	for iter.Next() {
		fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
	}
	return obj
}

func (obj *kvStream) Map(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	iter := obj.mapVal.MapRange()
	keyTp, valTp := obj.parseMapFunction(fn)
	table := reflect.MakeMap(reflect.MapOf(keyTp, valTp))
	for iter.Next() {
		out := fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
		table.SetMapIndex(out[0], out[1])
	}
	return KVStreamOf(table.Interface())
}

func (obj *kvStream) MapValue(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	iter := obj.mapVal.MapRange()
	keyTp, valTp, fnVal := obj.parseMapValueFunction(fn)
	table := reflect.MakeMap(reflect.MapOf(keyTp, valTp))
	for iter.Next() {
		out := fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
		table.SetMapIndex(iter.Key(), out[0])
	}
	return KVStreamOf(table.Interface())
}

func (obj *kvStream) MapKey(fn interface{}) KVStream {
	iter := obj.mapVal.MapRange()
	keyTp, fnVal := obj.parseMapKeyFunction(fn)
	table := reflect.MakeMap(reflect.MapOf(keyTp, obj.valType))
	for iter.Next() {
		out := fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
		table.SetMapIndex(out[0], iter.Value())
	}
	return KVStreamOf(table.Interface())
}

// Filter kv pair
func (obj *kvStream) Filter(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	obj.parseFilterFunction(fn)

	table := reflect.MakeMap(obj.mapVal.Type())
	iter := obj.mapVal.MapRange()
	for iter.Next() {
		k, v := iter.Key(), iter.Value()
		if ok := fnVal.Call([]reflect.Value{k, v})[0].Bool(); ok {
			table.SetMapIndex(k, v)
		}
	}
	return KVStreamOf(table.Interface())
}

// Reject kv pair
func (obj *kvStream) Reject(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	obj.parseFilterFunction(fn)

	table := reflect.MakeMap(obj.mapVal.Type())
	iter := obj.mapVal.MapRange()
	for iter.Next() {
		k, v := iter.Key(), iter.Value()
		if ok := fnVal.Call([]reflect.Value{k, v})[0].Bool(); !ok {
			table.SetMapIndex(k, v)
		}
	}
	return KVStreamOf(table.Interface())
}

// Contains key
func (obj *kvStream) Contains(key interface{}) bool {
	kval := reflect.ValueOf(key)
	if kval.Type() != obj.keyType && kval.Type().ConvertibleTo(obj.keyType) {
		kval = kval.Convert(obj.keyType)
	}
	if ele := obj.mapVal.MapIndex(kval); !ele.IsValid() {
		return false
	}
	return true
}

// Keys of object
func (obj *kvStream) Keys() Stream {
	keys := obj.mapVal.MapKeys()
	slice := reflect.MakeSlice(reflect.SliceOf(obj.keyType), len(keys), len(keys))
	for i := 0; i < len(keys); i++ {
		slice.Index(i).Set(keys[i])
	}
	return StreamOf(slice.Interface())
}

// Values of object
func (obj *kvStream) Values() Stream {
	keys := obj.mapVal.MapKeys()
	slice := reflect.MakeSlice(reflect.SliceOf(obj.valType), len(keys), len(keys))
	for i := 0; i < len(keys); i++ {
		slice.Index(i).Set(obj.mapVal.MapIndex(keys[i]))
	}
	return StreamOf(slice.Interface())
}

func (l *kvStream) Result() interface{} {
	return Value{
		typ: l.mapVal.Type(),
		val: l.mapVal,
	}.Result()
}

// Size of map
func (obj *kvStream) Size() int {
	return obj.mapVal.Len()
}

func newKvList() *kvStream {
	return &kvStream{}
}

func (obj *kvStream) parseMapFunction(fn interface{}) (keytyp reflect.Type, valTyp reflect.Type) {
	tp := reflect.TypeOf(fn)
	if tp.Kind() != reflect.Func {
		panic("should be function")
	}
	if tp.NumIn() != 2 || tp.NumOut() != 2 {
		panic("map function should be 2 intput 2 output")
	}
	if tp.In(0) != obj.keyType || tp.In(1) != obj.valType {
		panic("map function input/output shoule match")
	}
	return tp.Out(0), tp.Out(1)
}

func (obj *kvStream) parseMapValueFunction(fn interface{}) (keytyp reflect.Type, valTyp reflect.Type, fnVal reflect.Value) {
	tp := reflect.TypeOf(fn)
	if tp.Kind() != reflect.Func {
		panic("should be function")
	}
	if tp.NumOut() != 1 {
		panic("map function should be 1 output")
	}
	if tp.NumIn() == 0 || tp.NumIn() > 2 {
		panic("map function should be at least 1 input")
	}
	switch tp.NumIn() {
	case 1:
		if tp.In(0) != obj.valType {
			panic("map function bad type")
		}
		ft := reflect.FuncOf([]reflect.Type{obj.keyType, obj.valType}, []reflect.Type{tp.Out(0)}, false)
		return obj.keyType, tp.Out(0), reflect.MakeFunc(ft, func(in []reflect.Value) []reflect.Value {
			return reflect.ValueOf(fn).Call(in[1:])
		})
	case 2:
		if tp.In(0) != obj.keyType || tp.In(1) != obj.valType {
			panic("map function bad type")
		}
		return tp.In(0), tp.Out(0), reflect.ValueOf(fn)
	default:
		panic("map function should be at least 1 input")
	}
}

func (obj *kvStream) parseMapKeyFunction(fn interface{}) (reflect.Type, reflect.Value) {
	tp := reflect.TypeOf(fn)
	if tp.Kind() != reflect.Func {
		panic("should be function")
	}
	if tp.NumOut() != 1 || tp.NumIn() == 0 || tp.In(0) != obj.keyType {
		panic("map function should be 1 output")
	}
	if tp.NumIn() == 2 && tp.In(1) != obj.valType {
		panic("map function bad signature")
	}
	fnval := reflect.ValueOf(fn)
	if tp.NumIn() == 1 {
		ft := reflect.FuncOf([]reflect.Type{obj.keyType, obj.valType}, []reflect.Type{tp.Out(0)}, false)
		fnval = reflect.MakeFunc(ft, func(in []reflect.Value) []reflect.Value {
			return reflect.ValueOf(fn).Call(in[:1])
		})
	}
	return tp.Out(0), fnval
}

func (obj *kvStream) parseFilterFunction(fn interface{}) {
	tp := reflect.TypeOf(fn)
	if tp.Kind() != reflect.Func {
		panic("should be function")
	}
	if tp.NumIn() != 2 || tp.NumOut() != 1 {
		panic("filter function should be 2 intput 2 output")
	}
	if tp.In(0) != obj.keyType || tp.In(1) != obj.valType {
		panic("filter function input/output shoule match")
	}
	if tp.Out(0).Kind() != reflect.Bool {
		panic("filter function output shoule be boolean")
	}
}
