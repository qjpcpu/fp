package fp

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Stream interface {
	// Map stream to another, fn should be func(element_type) (another_type,&optional error/bool)
	Map(fn interface{}) Stream
	// FlatMap stream to another, fn should be func(element_type) slice_type
	FlatMap(fn interface{}) Stream
	// Filter stream, fn should be func(element_type) bool
	Filter(fn interface{}) Stream
	// Reject stream, fn should be func(element_type) bool
	Reject(fn interface{}) Stream
	// Foreach stream element, fn should be func(element_type,optional[int])
	Foreach(fn interface{}) Stream
	// Flatten stream, element should be flatten-able
	Flatten() Stream
	// Reduce stream, fn should be func(initval_type, element_type) initval_type
	Reduce(initval interface{}, fn interface{}) Value
	// Reduce0 stream, fn should be func(element_type,element_type) element_type
	Reduce0(fn interface{}) Value
	// Partition stream, split stream into small batch
	Partition(size int) Stream
	// PartitionBy func(elem_type) bool
	PartitionBy(fn interface{}, includeSplittor bool) Stream
	// First value of stream
	First() Value
	// IsEmpty stream
	IsEmpty() bool
	// HasSomething in stream
	HasSomething() bool
	// Exists at least one element in stream
	Exists() bool
	// Take first n elements
	Take(n int) Stream
	// TakeWhile fn return false
	TakeWhile(fn interface{}) Stream
	// Skip first n elements
	Skip(size int) Stream
	// SkipWhile fn return false
	SkipWhile(fn interface{}) Stream
	// Sort stream, this is an aggregate op, so it would block stream
	Sort() Stream
	// SortBy fn stream, fn should be func(element_type,element_type) bool, this is an aggregate op, so it would block stream
	SortBy(fn interface{}) Stream
	// Uniq stream, keep first when duplicated, this is an aggregate op, so it would block stream
	Uniq() Stream
	// UniqBy stream, keep first when duplicated, fn should be func(element_type) any_type, this is an aggregate op, so it would block stream
	UniqBy(fn interface{}) Stream
	// Size of stream, this is an aggregate op, so it would block stream
	Size() int
	// Count alias to Size
	Count() int
	// Contains element
	Contains(interface{}) bool
	// ContainsBy func(element_type) bool
	ContainsBy(fn interface{}) bool
	// ToSource convert stream to source
	ToSource() Source
	// Sub stream
	Sub(other Stream) Stream
	// SubBy keyfn, keyfn is func(element_type) any_type
	SubBy(other Stream, keyfn interface{}) Stream
	// Interact stream
	Interact(other Stream) Stream
	// InteractBy keyfn, keyfn is func(element_type) any_type
	InteractBy(other Stream, keyfn interface{}) Stream
	// Union append another stream
	Union(Stream) Stream
	// ToSet element as key, value is bool
	ToSet() KVStream
	// ToSet by func(element_type) any_type or func(element_type) (key_type,val_type)
	ToSetBy(fn interface{}) KVStream
	// GroupBy func(element_type) any_type, result is a kv set (any_type: [element_type]), this is an aggregate op, so it would block stream
	GroupBy(fn interface{}) KVStream
	// Append element
	Append(element ...interface{}) Stream
	// Prepend element
	Prepend(element ...interface{}) Stream
	// Zip stream , fn should be func(self_element_type,other_element_type) another_type
	Zip(other Stream, fn interface{}) Stream
	// ZipN multiple stream , fn should be func(self_element_type,other_element_type1,other_element_type2,...) another_type
	ZipN(fn interface{}, others ...Stream) Stream
	// Branch to multiple output stream, should not used for unlimited stream
	Branch(processors ...StreamProcessor)
	// Reverse a stream
	Reverse() Stream

	// Run stream and drop value
	Run()
	// ToSlice ptr
	ToSlice(ptr interface{}) error
	// Result of stream
	Result() interface{}
	// shortcuts
	Strings() []string
	StringsList() [][]string
	Ints() []int
	Float64s() []float64
	Bytes() []byte
	Int64s() []int64
	Int32s() []int32
	Uints() []uint
	Uint32s() []uint32
	Uint64s() []uint64
	// JoinStrings shortcut for strings.Join(stream.Strings(),seq)
	JoinStrings(seq string) string
}

func StreamOf(arr interface{}) Stream {
	elemTyp, it := makeIter(reflect.ValueOf(arr))
	return newStream(nil, elemTyp, it)
}

func Stream0Of(arr ...interface{}) Stream {
	return StreamOf(arr[0])
}

func StreamOfSource(s Source) Stream {
	return newStream(nil, s.ElemType(), s.Next)
}

type StreamProcessor func(Stream)

type stream struct {
	expectElemTyp reflect.Type
	iter          iterator
	val           reflect.Value
	getValOnce    sync.Once
	ctx           context
}

func newStream(ctx context, expTyp reflect.Type, it iterator, mws ...middleware) *stream {
	if it == nil {
		it = func() (reflect.Value, bool) { return reflect.Value{}, false }
	}
	if ctx == nil {
		ctx = newCtx(ctx)
	}
	if ctx.Err() != nil {
		it = func() (reflect.Value, bool) { return reflect.Value{}, false }
	} else {
		for i := range mws {
			it = mws[i](it)
		}
	}
	return &stream{expectElemTyp: expTyp, iter: it, ctx: ctx}
}

func (q *stream) Map(fn interface{}) Stream {
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	ctx := newCtx(q.ctx)
	mapFn := func(in reflect.Value) (reflect.Value, bool) {
		return fnVal.Call([]reflect.Value{in})[0], true
	}
	if fnTyp.NumOut() == 2 && fnTyp.Out(1) == boolType {
		mapFn = func(in reflect.Value) (reflect.Value, bool) {
			out := fnVal.Call([]reflect.Value{in})
			return out[0], out[1].Bool()
		}
	} else if fnTyp.NumOut() == 2 && fnTyp.Out(1).ConvertibleTo(errType) {
		mapFn = func(in reflect.Value) (reflect.Value, bool) {
			out := fnVal.Call([]reflect.Value{in})
			err := out[1].Interface()
			if err != nil && err.(error) != nil {
				ctx.SetErr(err.(error))
			}
			return out[0], err == nil || err.(error) == nil
		}
	} else if fnTyp.NumOut() == 1 {
	} else {
		panic("Map function must be func(element_type) another_type or func(element_type) (another_type,error/bool), now " + fnTyp.String())
	}

	return newStream(ctx, fnTyp.Out(0), q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for {
				if val, ok := next(); !ok {
					return reflect.Value{}, false
				} else if val, ok = mapFn(val); ok {
					return val, true
				} else if ctx.Err() != nil {
					return val, false
				}
			}
		}
	})
}

func (q *stream) Zip(other Stream, fn interface{}) Stream {
	if isNilStream(other) {
		return other
	}
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	onext := other.ToSource().Next
	return newStream(newCtx(q.ctx), fnTyp.Out(0), q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if val1, ok1 := next(); ok1 {
				if val2, ok2 := onext(); ok2 {
					return fnVal.Call([]reflect.Value{val1, val2})[0], true
				}
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) ZipN(fn interface{}, others ...Stream) Stream {
	for _, s := range others {
		if isNilStream(s) {
			return s
		}
	}
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	if fnTyp.NumIn() != len(others)+1 {
		panic(fmt.Sprintf("zip function must have %v input param", len(others)+1))
	}

	return newStream(newCtx(q.ctx), fnTyp.Out(0), q.iter, func(next iterator) iterator {
		/* build iterator list */
		var iteratorList []iterator
		StreamOf(others).Map(func(s Stream) iterator {
			return s.ToSource().Next
		}).Prepend(next).
			ToSlice(&iteratorList)

		var done bool
		return func() (reflect.Value, bool) {
			if !done {
				input := make([]reflect.Value, len(others)+1)
				for i := range iteratorList {
					if val, ok := iteratorList[i](); ok {
						input[i] = val
					} else {
						done = true
						return reflect.Value{}, false
					}
				}
				return fnVal.Call(input)[0], true
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) FlatMap(fn interface{}) Stream {
	return q.Map(fn).Flatten()
}

func (q *stream) Filter(fn interface{}) Stream {
	fnVal := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for {
				if val, ok := next(); !ok {
					break
				} else if fnVal.Call([]reflect.Value{val})[0].Bool() {
					return val, true
				}
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) Flatten() Stream {
	if kind := q.expectElemTyp.Kind(); kind != reflect.Chan && kind != reflect.Slice && kind != reflect.Array && q.expectElemTyp != streamType {
		panic(q.expectElemTyp.String() + " can not be flatten")
	}

	var elemType reflect.Type
	_makeIter := makeIter
	if q.expectElemTyp == streamType {
		iter := q.iter
		for {
			v, ok := iter()
			if !ok {
				/* no inner stream found and cause we couldn't guess inner type */
				/* just return nil stream */
				return newNilStream()
			}
			/* we should find first non-NilStream element */
			if isNilStream(v.Interface().(Stream)) {
				continue
			}
			/* gotcha */
			elemType = v.Interface().(Stream).ToSource().ElemType()
			/* recover q's iter */
			var flag int32
			q.iter = func() (reflect.Value, bool) {
				if atomic.CompareAndSwapInt32(&flag, 0, 1) {
					return v, true
				}
				return iter()
			}
			break
		}
		_makeIter = func(v reflect.Value) (reflect.Type, iterator) {
			return elemType, v.Interface().(Stream).ToSource().Next
		}
	} else {
		elemType = q.expectElemTyp.Elem()
	}
	return newStream(newCtx(q.ctx), elemType, q.iter, func(outernext iterator) iterator {
		var innernext iterator
		var inner reflect.Value
		return func() (item reflect.Value, ok bool) {
			for !ok {
				if !inner.IsValid() {
					inner, ok = outernext()
					if !ok {
						return
					}
					/* we should jump over nil stream element */
					if innerS, isStream := inner.Interface().(Stream); isStream && isNilStream(innerS) {
						/* disable inner value */
						inner = reflect.Value{}
						ok = false
						continue
					}
					_, innernext = _makeIter(inner)
				}
				item, ok = innernext()
				if !ok {
					inner = reflect.Value{}
				}
			}
			return
		}
	})
}

func (q *stream) Foreach(fn interface{}) Stream {
	fnval := reflect.ValueOf(fn)
	withIndex := fnval.Type().NumIn() == 2
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var i int
		return func() (val reflect.Value, ok bool) {
			if val, ok = next(); ok {
				if withIndex {
					fnval.Call([]reflect.Value{val, reflect.ValueOf(i)})
					i++
				} else {
					fnval.Call([]reflect.Value{val})
				}
			}
			return
		}
	})
}

func (q *stream) Prepend(v ...interface{}) Stream {
	nq := q
	for i := len(v) - 1; i >= 0; i-- {
		nq = nq.prependOne(v[i])
	}
	return nq
}

func (q *stream) Append(v ...interface{}) Stream {
	nq := q
	for _, elem := range v {
		nq = nq.appendOne(elem)
	}
	return nq
}

func (q *stream) IsEmpty() bool {
	old := q.iter
	v, ok := q.iter()
	if ok {
		var flag int32
		q.iter = func() (reflect.Value, bool) {
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				return v, true
			}
			return old()
		}
	}
	return !ok
}

func (q *stream) HasSomething() bool {
	return !q.IsEmpty()
}

func (q *stream) Exists() bool {
	return !q.IsEmpty()
}

func (q *stream) Take(size int) Stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if size > 0 {
				if val, ok := next(); ok {
					size--
					return val, true
				}
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) TakeWhile(fn interface{}) Stream {
	fnval := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if val, ok := next(); ok && fnval.Call([]reflect.Value{val})[0].Bool() {
				return val, true
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) Skip(size int) Stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for ; size > 0; size-- {
				if _, ok := next(); !ok {
					return reflect.Value{}, false
				}
			}
			return next()
		}
	})
}

func (q *stream) SkipWhile(fn interface{}) Stream {
	fnval := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var flag int32
		return func() (reflect.Value, bool) {
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				for {
					val, ok := next()
					if !ok {
						return reflect.Value{}, false
					}
					if !fnval.Call([]reflect.Value{val})[0].Bool() {
						return val, true
					}
				}
			}
			return next()
		}
	})
}

func (q *stream) First() (f Value) {
	f.typ = q.expectElemTyp
	q.iter = repeatableIter(q.iter, func(v reflect.Value) bool {
		f.val = v
		return false
	})
	return f
}

func (q *stream) Sort() Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			arr := q.getResult().Result()
			v := reflect.ValueOf(arr)
			sort.SliceStable(arr, func(i, j int) bool {
				return q.compare(v.Index(i), v.Index(j)) < 0
			})
			val := Value{
				typ: reflect.TypeOf(arr),
				val: reflect.ValueOf(arr),
			}
			_, iter = makeIter(val.val)
		}
		return iter()
	})
}

func (q *stream) Reverse() Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			arr := q.getResult().Result()
			v := reflect.ValueOf(arr)
			idx := v.Len() - 1
			iter = func() (reflect.Value, bool) {
				if idx >= 0 {
					idx--
					return v.Index(idx + 1), true
				}
				return reflect.Value{}, false
			}
		}
		return iter()
	})
}

func (q *stream) Uniq() Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			dup := make(map[interface{}]struct{})
			iter = func() (reflect.Value, bool) {
				for {
					val, ok := q.iter()
					if !ok {
						return val, false
					}
					key := val.Interface()
					if _, ok := dup[key]; !ok {
						dup[key] = struct{}{}
						return val, true
					}
				}
			}
		}
		return iter()
	})
}

func (q *stream) Union(other Stream) Stream {
	if isNilStream(other) {
		return q
	}
	oNext := other.ToSource().Next
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var otherDone bool
		return func() (reflect.Value, bool) {
			if !otherDone {
				val, ok := next()
				if ok {
					return val, true
				}
				otherDone = true
			}
			return oNext()
		}
	})
}

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

func (q *stream) Interact(other Stream) Stream {
	if isNilStream(other) {
		return other
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
	return q.Filter(fn.Interface())
}

func (q *stream) InteractBy(other Stream, keyfn interface{}) Stream {
	if isNilStream(other) {
		return other
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
	return q.Filter(fn.Interface())
}

func (q *stream) ToSetBy(fn interface{}) KVStream {
	fntyp := reflect.TypeOf(fn)
	fnval := reflect.ValueOf(fn)

	keyTyp, valTyp := fntyp.Out(0), q.expectElemTyp
	getKV := func(elem reflect.Value) (reflect.Value, reflect.Value) {
		return fnval.Call([]reflect.Value{elem})[0], elem
	}
	if fntyp.NumOut() == 2 {
		keyTyp, valTyp = fntyp.Out(0), fntyp.Out(1)
		getKV = func(elem reflect.Value) (reflect.Value, reflect.Value) {
			res := fnval.Call([]reflect.Value{elem})
			return res[0], res[1]
		}
	}
	iter := q.iter
	return newKvStream(newCtx(q.ctx), keyTyp, valTyp, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(keyTyp, valTyp))
		for {
			val, ok := iter()
			if !ok {
				break
			}
			table.SetMapIndex(getKV(val))
		}
		return table
	})
}

func (q *stream) ToSet() KVStream {
	iter := q.iter
	return newKvStream(newCtx(q.ctx), q.expectElemTyp, boolType, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(q.expectElemTyp, boolType))
		_true := reflect.ValueOf(true)
		for {
			val, ok := iter()
			if !ok {
				break
			}
			table.SetMapIndex(val, _true)
		}
		return table
	})
}

func (q *stream) UniqBy(fn interface{}) Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			getKey := reflect.ValueOf(fn)
			dup := make(map[interface{}]struct{})
			iter = func() (reflect.Value, bool) {
				for {
					val, ok := q.iter()
					if !ok {
						return val, false
					}
					key := getKey.Call([]reflect.Value{val})[0].Interface()
					if _, ok := dup[key]; !ok {
						dup[key] = struct{}{}
						return val, true
					}
				}
			}
		}
		return iter()
	})
}

func (q *stream) SortBy(fn interface{}) Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			arr := q.getResult().Result()
			v := reflect.ValueOf(arr)
			fnval := reflect.ValueOf(fn)
			sort.SliceStable(arr, func(i, j int) bool {
				return fnval.Call([]reflect.Value{v.Index(i), v.Index(j)})[0].Bool()
			})
			val := Value{
				typ: reflect.TypeOf(arr),
				val: reflect.ValueOf(arr),
			}
			_, iter = makeIter(val.val)
		}
		return iter()
	})
}

func (q *stream) Reject(fn interface{}) Stream {
	typ := reflect.TypeOf(fn)
	val := reflect.ValueOf(fn)
	fnr := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		b := val.Call(in)[0].Bool()
		return []reflect.Value{reflect.ValueOf(!b)}
	})
	return q.Filter(fnr.Interface())
}

func (q *stream) Reduce(initval interface{}, fn interface{}) Value {
	typ := reflect.TypeOf(initval)
	memo := reflect.ValueOf(initval)
	fnval := reflect.ValueOf(fn)
	for {
		val, ok := q.iter()
		if !ok {
			break
		}
		memo = fnval.Call([]reflect.Value{memo, val})[0]
	}
	return Value{typ: typ, val: memo}
}

func (q *stream) Reduce0(fn interface{}) Value {
	initVal, ok := q.iter()
	if !ok {
		return Value{typ: q.expectElemTyp, val: reflect.Zero(q.expectElemTyp)}
	}
	return q.Reduce(initVal.Interface(), fn)
}

func (q *stream) Partition(size int) Stream {
	if size < 1 {
		panic("batch size should be greater than 0")
	}
	return newStream(newCtx(q.ctx), reflect.SliceOf(q.expectElemTyp), q.iter, func(next iterator) iterator {
		typ := reflect.SliceOf(q.expectElemTyp)
		return func() (reflect.Value, bool) {
			var slice reflect.Value
			for i := 0; i < size; i++ {
				if val, ok := next(); !ok {
					break
				} else {
					if !slice.IsValid() {
						slice = reflect.Zero(typ)
					}
					slice = reflect.Append(slice, val)
				}
			}
			return slice, slice.IsValid() && slice.Len() > 0
		}
	})
}

func (q *stream) PartitionBy(fn interface{}, includeSplittor bool) Stream {
	fnval := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), reflect.SliceOf(q.expectElemTyp), q.iter, func(next iterator) iterator {
		typ := reflect.SliceOf(q.expectElemTyp)
		return func() (reflect.Value, bool) {
			var slice reflect.Value
			for {
				if val, ok := next(); !ok {
					break
				} else {
					if !slice.IsValid() {
						slice = reflect.Zero(typ)
					}
					if fnval.Call([]reflect.Value{val})[0].Bool() {
						if includeSplittor {
							slice = reflect.Append(slice, val)
						}
						break
					}
					slice = reflect.Append(slice, val)
				}
			}
			return slice, slice.IsValid() && slice.Len() > 0
		}
	})
}

func (q *stream) ToSlice(dst interface{}) error {
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(q.getValue(val.Elem()))
	return q.ctx.Err()
}

func (q *stream) GroupBy(fn interface{}) KVStream {
	keyTyp := reflect.TypeOf(fn).Out(0)
	valTyp := reflect.SliceOf(q.expectElemTyp)

	iter := q.iter
	return newKvStream(newCtx(q.ctx), keyTyp, valTyp, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(keyTyp, valTyp))
		fnVal := reflect.ValueOf(fn)
		for {
			val, ok := iter()
			if !ok {
				break
			}
			key := fnVal.Call([]reflect.Value{val})[0]
			slice := table.MapIndex(key)
			if !slice.IsValid() {
				slice = reflect.Zero(valTyp)
			}
			slice = reflect.Append(slice, val)
			table.SetMapIndex(key, slice)
		}
		return table
	})
}

func (q *stream) JoinStrings(seq string) string {
	return strings.Join(q.Strings(), seq)
}

func (q *stream) getResult() Value {
	return Value{
		typ: reflect.SliceOf(q.expectElemTyp),
		val: q.getValue(reflect.Value{}),
	}
}

func (q *stream) Size() int {
	return q.getValue(reflect.Value{}).Len()
}

func (q *stream) Count() int {
	return q.Size()
}

func (q *stream) Contains(e interface{}) (yes bool) {
	var eq func(reflect.Value) bool
	val := reflect.ValueOf(e)
	switch reflect.TypeOf(e).Kind() {
	case reflect.String:
		t := val.String()
		eq = func(v reflect.Value) bool { return v.String() == t }
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t := val.Int()
		eq = func(v reflect.Value) bool { return v.Int() == t }
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		t := val.Uint()
		eq = func(v reflect.Value) bool { return v.Uint() == t }
	case reflect.Bool:
		t := val.Bool()
		eq = func(v reflect.Value) bool { return v.Bool() == t }
	case reflect.Float32, reflect.Float64:
		t := val.Float()
		eq = func(v reflect.Value) bool { return v.Float() == t }
	default:
		eq = func(v reflect.Value) bool { return reflect.DeepEqual(v.Interface(), e) }
	}

	q.iter = repeatableIter(q.iter, func(v reflect.Value) bool {
		yes = eq(v)
		return !yes
	})
	return
}

func (q *stream) ContainsBy(eqfn interface{}) (yes bool) {
	fnval := reflect.ValueOf(eqfn)
	q.iter = repeatableIter(q.iter, func(v reflect.Value) bool {
		yes = fnval.Call([]reflect.Value{v})[0].Bool()
		return !yes
	})
	return
}

func (q *stream) Run() {
	q.getValOnce.Do(func() {
		for {
			if _, ok := q.iter(); !ok {
				break
			}
		}
	})
}

/* stream to source */
func (q *stream) ToSource() Source {
	return q
}

func (q *stream) ElemType() reflect.Type {
	return q.expectElemTyp
}

func (q *stream) Next() (reflect.Value, bool) {
	return q.iter()
}

func (q *stream) getValue(slice reflect.Value) reflect.Value {
	q.getValOnce.Do(func() {
		if !slice.IsValid() {
			slice = reflect.Zero(reflect.SliceOf(q.expectElemTyp))
		} else if slice.Len() > 0 {
			slice = reflect.MakeSlice(reflect.SliceOf(q.expectElemTyp), 0, slice.Len())
		}
		for {
			if val, ok := q.iter(); ok {
				slice = reflect.Append(slice, val)
			} else {
				break
			}
		}
		q.val = slice
	})
	return q.val
}

func (q *stream) compare(a, b reflect.Value) int {
	switch q.expectElemTyp.Kind() {
	case reflect.String:
		if a.String() < b.String() {
			return -1
		} else if a.String() > b.String() {
			return 1
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() < b.Int() {
			return -1
		} else if a.Int() > b.Int() {
			return 1
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if a.Uint() < b.Uint() {
			return -1
		} else if a.Uint() > b.Uint() {
			return 1
		}
	case reflect.Bool:
		if !a.Bool() && b.Bool() {
			return -1
		} else if a.Bool() && !b.Bool() {
			return 1
		}
	default:
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			s1, s2 := fmt.Sprint(a.Interface()), fmt.Sprint(b.Interface())
			if s1 < s2 {
				return -1
			} else if s1 > s2 {
				return 1
			}
		}
	}
	return 0
}

func (q *stream) prependOne(v interface{}) *stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var flag int32
		return func() (reflect.Value, bool) {
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				return reflect.ValueOf(v), true
			}
			return next()
		}
	})
}

func (q *stream) appendOne(v interface{}) *stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var flag int32
		return func() (reflect.Value, bool) {
			if flag == 0 {
				if val, ok := next(); ok {
					return val, ok
				}
			}
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				return reflect.ValueOf(v), true
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) Branch(processors ...StreamProcessor) {
	switch len(processors) {
	case 0:
		return
	case 1:
		processors[0](q)
		return
	}

	slice := reflect.MakeSlice(reflect.SliceOf(q.expectElemTyp), 0, 0)
	movingStream := newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (val reflect.Value, ok bool) {
			if val, ok = next(); ok {
				slice = reflect.Append(slice, val)
			}
			return
		}
	})

	for _, processor := range processors {
		processor(StreamOf(slice.Interface()).Union(movingStream))
	}
}

func (q *stream) Result() interface{}         { return q.getResult().Result() }
func (q *stream) Strings() (s []string)       { return q.getResult().Strings() }
func (q *stream) Ints() (s []int)             { return q.getResult().Ints() }
func (q *stream) Int64s() (s []int64)         { return q.getResult().Int64s() }
func (q *stream) Int32s() (s []int32)         { return q.getResult().Int32s() }
func (q *stream) Uints() (s []uint)           { return q.getResult().Uints() }
func (q *stream) Uint32s() (s []uint32)       { return q.getResult().Uint32s() }
func (q *stream) Uint64s() (s []uint64)       { return q.getResult().Uint64s() }
func (q *stream) Bytes() (s []byte)           { return q.getResult().Bytes() }
func (q *stream) Float64s() (s []float64)     { q.getResult().To(&s); return }
func (q *stream) StringsList() (s [][]string) { return q.getResult().StringsList() }

/* value related */
type Value struct {
	typ reflect.Type
	val reflect.Value
}

func (rv Value) To(dst interface{}) bool {
	if !rv.val.IsValid() {
		return false
	}
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(rv.val)
	return true
}

func (rv Value) Result() interface{} {
	if !rv.val.IsValid() {
		return nil
	}
	return rv.val.Interface()
}

func (rv Value) Err() error {
	if !rv.val.IsValid() {
		return nil
	}
	res := rv.val.Interface()
	if res == nil {
		return nil
	}
	return res.(error)
}

func (rv Value) Strings() (s []string) {
	rv.To(&s)
	return
}

func (rv Value) Bytes() (s []byte) {
	rv.To(&s)
	return
}

func (rv Value) Ints() (s []int) {
	rv.To(&s)
	return
}

func (rv Value) Int64s() (s []int64) {
	rv.To(&s)
	return
}

func (rv Value) Int32s() (s []int32) {
	rv.To(&s)
	return
}

func (rv Value) Uints() (s []uint) {
	rv.To(&s)
	return
}

func (rv Value) Uint32s() (s []uint32) {
	rv.To(&s)
	return
}

func (rv Value) Uint64s() (s []uint64) {
	rv.To(&s)
	return
}

func (rv Value) StringsList() (s [][]string) {
	rv.To(&s)
	return
}

func (rv Value) String() (s string) {
	rv.To(&s)
	return
}

func (rv Value) Int() (s int) {
	rv.To(&s)
	return
}

func (rv Value) Int64() (s int64) {
	rv.To(&s)
	return
}

func (rv Value) Int32() (s int32) {
	rv.To(&s)
	return
}

func (rv Value) Uint32() (s uint32) {
	rv.To(&s)
	return
}

func (rv Value) Uint64() (s uint64) {
	rv.To(&s)
	return
}

func (rv Value) Float64() (s float64) {
	rv.To(&s)
	return
}

func repeatableIter(iter iterator, f func(reflect.Value) bool) iterator {
	if iter == nil {
		return nil
	}
	var vals []reflect.Value
	for {
		val, ok := iter()
		if !ok {
			break
		}
		vals = append(vals, val)
		if !f(val) {
			break
		}
	}
	var i int
	return func() (reflect.Value, bool) {
		for i < len(vals) {
			i++
			return vals[i-1], true
		}
		return iter()
	}
}
