package fp

import (
	"reflect"
	"sync/atomic"
)

type KVStream interface {
	// Foreach element of object
	// fn should be func(key_type,element_type)
	Foreach(fn interface{}) KVStream
	// Map k-v pair
	// fn should be func(key_type,element_type) (any_type,any_type)
	Map(fn interface{}) KVStream
	// FlatMap map to array, fn should be func(key_type,element_type) any_type
	FlatMap(fn interface{}) Stream
	// Filter kv pair
	Filter(fn interface{}) KVStream
	// Reject kv pair
	Reject(fn interface{}) KVStream
	// Contains key
	Contains(key interface{}) bool
	// Keys of map
	Keys() Stream
	// Values of map
	Values() Stream
	// Size of map
	Size() int
	// Result of map
	Result() interface{}
}

type kvStream struct {
	getMap           func() reflect.Value
	keyType, valType reflect.Type
}

func KVStreamOf(m interface{}) KVStream {
	if s, ok := m.(KVSource); ok {
		return KVStreamOfSource(s)
	}
	if reflect.TypeOf(m).Kind() != reflect.Map {
		panic("argument should be map")
	}
	tp := reflect.TypeOf(m)
	return newKvStream(tp.Key(), tp.Elem(), func() reflect.Value {
		return reflect.ValueOf(m)
	})
}

func KVStreamOfSource(s KVSource) KVStream {
	keyType, valType := s.ElemType()
	return newKvStream(keyType, valType, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(keyType, valType))
		for {
			if k, v, ok := s.Next(); ok {
				table.SetMapIndex(k, v)
			} else {
				break
			}
		}
		return table
	})
}

func (obj *kvStream) Foreach(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	getMap := obj.getMap
	return newKvStream(obj.keyType, obj.valType, func() reflect.Value {
		mp := getMap()
		iter := mp.MapRange()
		for iter.Next() {
			fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
		}
		return mp
	})
}

func (obj *kvStream) Map(fn interface{}) KVStream {
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	getMap := obj.getMap
	return newKvStream(fnTyp.Out(0), fnTyp.Out(1), func() reflect.Value {
		iter := getMap().MapRange()
		table := reflect.MakeMap(reflect.MapOf(fnTyp.Out(0), fnTyp.Out(1)))
		for iter.Next() {
			out := fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
			table.SetMapIndex(out[0], out[1])
		}
		return table
	})
}

func (obj *kvStream) FlatMap(fn interface{}) Stream {
	fnVal := reflect.ValueOf(fn)
	var iter *reflect.MapIter
	var done bool
	return newStream(fnVal.Type().Out(0), func() (reflect.Value, bool) {
		if iter == nil {
			iter = obj.getMap().MapRange()
		}
		if !done && iter.Next() {
			out := fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
			return out[0], true
		}
		done = true
		return reflect.Value{}, false
	})
}

// Filter kv pair
func (obj *kvStream) Filter(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	return newKvStream(obj.keyType, obj.valType, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(obj.keyType, obj.valType))
		iter := obj.getMap().MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			if ok := fnVal.Call([]reflect.Value{k, v})[0].Bool(); ok {
				table.SetMapIndex(k, v)
			}
		}
		return table
	})
}

// Reject kv pair
func (obj *kvStream) Reject(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	return newKvStream(obj.keyType, obj.valType, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(obj.keyType, obj.valType))
		iter := obj.getMap().MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			if ok := fnVal.Call([]reflect.Value{k, v})[0].Bool(); !ok {
				table.SetMapIndex(k, v)
			}
		}
		return table
	})
}

// Contains key
func (obj *kvStream) Contains(key interface{}) bool {
	kval := reflect.ValueOf(key)
	if kval.Type() != obj.keyType && kval.Type().ConvertibleTo(obj.keyType) {
		kval = kval.Convert(obj.keyType)
	}
	if ele := obj.getMap().MapIndex(kval); !ele.IsValid() {
		return false
	}
	return true
}

// Keys of object
func (obj *kvStream) Keys() Stream {
	var iter *reflect.MapIter
	var done bool
	return newStream(obj.keyType, func() (reflect.Value, bool) {
		if iter == nil {
			iter = obj.getMap().MapRange()
		}
		if !done && iter.Next() {
			return iter.Key(), true
		}
		done = true
		return reflect.Value{}, false
	})
}

// Values of object
func (obj *kvStream) Values() Stream {
	var iter *reflect.MapIter
	var done bool
	return newStream(obj.valType, func() (reflect.Value, bool) {
		if iter == nil {
			iter = obj.getMap().MapRange()
		}
		if !done && iter.Next() {
			return iter.Value(), true
		}
		done = true
		return reflect.Value{}, false
	})
}

func (l *kvStream) Result() interface{} {
	return Value{
		typ: reflect.MapOf(l.keyType, l.valType),
		val: l.getMap(),
	}.Result()
}

// Size of map
func (obj *kvStream) Size() int {
	return obj.getMap().Len()
}

func newKvStream(k, v reflect.Type, getmp func() reflect.Value) *kvStream {
	return &kvStream{keyType: k, valType: v, getMap: getMapOnce(getmp)}
}

func getMapOnce(f func() reflect.Value) func() reflect.Value {
	var flag int32
	var v reflect.Value
	return func() reflect.Value {
		if atomic.CompareAndSwapInt32(&flag, 0, 1) {
			v = f()
		}
		return v
	}
}
