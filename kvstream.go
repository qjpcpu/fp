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
	// fn should be func(key_type,element_type) (any_type,any_type,&optional error)
	Map(fn interface{}) KVStream
	// ZipMap map to array, fn should be func(key_type,element_type) (any_type,&optional error)
	ZipMap(fn interface{}) Stream
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
	// Run stream
	Run()
	// To dst ptr
	To(dstPtr interface{}) error
}

type kvStream struct {
	getMap           func() reflect.Value
	keyType, valType reflect.Type
	ctx              context
}

func KVStreamOf(m interface{}) KVStream {
	if s, ok := m.(KVSource); ok {
		return KVStreamOfSource(s)
	}
	if reflect.TypeOf(m).Kind() != reflect.Map {
		panic("argument should be map")
	}
	tp := reflect.TypeOf(m)
	return newKvStream(nil, tp.Key(), tp.Elem(), func() reflect.Value {
		return reflect.ValueOf(m)
	})
}

func KVStreamOfSource(s KVSource) KVStream {
	keyType, valType := s.ElemType()
	return newKvStream(nil, keyType, valType, func() reflect.Value {
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
	return newKvStream(newCtx(obj.ctx), obj.keyType, obj.valType, func() reflect.Value {
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
	hasErr := fnVal.Type().NumOut() == 3 && fnVal.Type().Out(2).ConvertibleTo(errType)
	ctx := newCtx(obj.ctx)
	return newKvStream(ctx, fnTyp.Out(0), fnTyp.Out(1), func() reflect.Value {
		iter := getMap().MapRange()
		table := reflect.MakeMap(reflect.MapOf(fnTyp.Out(0), fnTyp.Out(1)))
		for iter.Next() {
			out := fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
			if !hasErr {
			} else if err := obj.asErr(out[2].Interface()); err != nil {
				ctx.SetErr(err)
				break
			}
			table.SetMapIndex(out[0], out[1])
		}
		return table
	})
}

func (obj *kvStream) ZipMap(fn interface{}) Stream {
	fnVal := reflect.ValueOf(fn)
	var iter *reflect.MapIter
	var done bool
	ctx := newCtx(obj.ctx)
	hasErr := fnVal.Type().NumOut() == 2 && fnVal.Type().Out(1).ConvertibleTo(errType)
	return newStream(ctx, fnVal.Type().Out(0), func() (reflect.Value, bool) {
		if iter == nil {
			iter = obj.getMap().MapRange()
		}
		if !done && iter.Next() {
			out := fnVal.Call([]reflect.Value{iter.Key(), iter.Value()})
			if !hasErr {
			} else if err := obj.asErr(out[1].Interface()); err != nil {
				ctx.SetErr(err)
				done = true
				return reflect.Value{}, false
			}
			return out[0], true
		}
		done = true
		return reflect.Value{}, false
	})
}

// Filter kv pair
func (obj *kvStream) Filter(fn interface{}) KVStream {
	fnVal := reflect.ValueOf(fn)
	return newKvStream(newCtx(obj.ctx), obj.keyType, obj.valType, func() reflect.Value {
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
	return newKvStream(newCtx(obj.ctx), obj.keyType, obj.valType, func() reflect.Value {
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
	return newStream(newCtx(obj.ctx), obj.keyType, func() (reflect.Value, bool) {
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
	return newStream(newCtx(obj.ctx), obj.valType, func() (reflect.Value, bool) {
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
	return l.getRelut().Result()
}

func (l *kvStream) Run() {
	_ = l.Result()
}

func (l *kvStream) To(ptr interface{}) error {
	l.getRelut().To(ptr)
	return l.ctx.Err()
}

func (l *kvStream) getRelut() Value {
	val := Value{
		typ: reflect.MapOf(l.keyType, l.valType),
		val: l.getMap(),
	}
	if !val.val.IsValid() || val.val.IsNil() {
		val.val = reflect.MakeMap(val.typ)
	}
	return val
}

func (obj *kvStream) asErr(out interface{}) error {
	if out == nil {
		return nil
	}
	return out.(error)
}

// Size of map
func (obj *kvStream) Size() int {
	return obj.getMap().Len()
}

func newKvStream(ctx context, k, v reflect.Type, getmp func() reflect.Value) *kvStream {
	if ctx == nil {
		ctx = newCtx(nil)
	}
	if ctx.Err() != nil {
		getmp = func() reflect.Value {
			return reflect.MakeMap(reflect.MapOf(k, v))
		}
	}
	return &kvStream{ctx: ctx, keyType: k, valType: v, getMap: getMapOnce(getmp)}
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
