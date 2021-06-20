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
	// Map stream to another, fn should be func(element_type) another_type
	Map(fn interface{}) Stream
	// FlatMap stream to another, fn should be func(element_type) another_slice_type
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
	// Interact stream
	Interact(other Stream) Stream
	// Union append another stream
	Union(Stream) Stream
	// ToSet element as key, value is bool
	ToSet() KVStream
	// ToSet by func(element_type) any_type
	ToSetBy(fn interface{}) KVStream
	// GroupBy func(element_type) any_type, this is an aggregate op, so it would block stream
	GroupBy(fn interface{}) KVStream
	// Append element
	Append(element ...interface{}) Stream
	// Prepend element
	Prepend(element ...interface{}) Stream
	// Zip stream , fn should be func(self_element_type,other_element_type) another_type
	Zip(other Stream, fn interface{}) Stream

	// Run stream and drop value
	Run()
	// ToSlice ptr
	ToSlice(ptr interface{})
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
	return newStream(elemTyp, it)
}

func StreamOfSource(s Source) Stream {
	return newStream(s.ElemType(), s.Next)
}

type stream struct {
	expectElemTyp reflect.Type
	iter          iterator
	val           reflect.Value
	getValOnce    sync.Once
}

func newStream(expTyp reflect.Type, it iterator, mws ...middleware) *stream {
	if it == nil {
		it = func() (reflect.Value, bool) { return reflect.Value{}, false }
	}
	for i := range mws {
		it = mws[i](it)
	}
	return &stream{expectElemTyp: expTyp, iter: it}
}

func (q *stream) Map(fn interface{}) Stream {
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	return newStream(fnTyp.Out(0), q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if val, ok := next(); !ok {
				return reflect.Value{}, false
			} else {
				return fnVal.Call([]reflect.Value{val})[0], true
			}
		}
	})
}

func (q *stream) Zip(other Stream, fn interface{}) Stream {
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	onext := other.ToSource().Next
	return newStream(fnTyp.Out(0), q.iter, func(next iterator) iterator {
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

func (q *stream) FlatMap(fn interface{}) Stream {
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	if fnTyp.NumOut() == 2 && fnTyp.Out(1) == boolType {
		return q.flatMapBoolean(fnTyp, fnVal)
	} else if fnTyp.NumOut() == 2 && fnTyp.Out(1).ConvertibleTo(errType) {
		return q.flatMapError(fnTyp, fnVal)
	}
	return q.Map(fn).Flatten()
}

func (q *stream) Filter(fn interface{}) Stream {
	fnVal := reflect.ValueOf(fn)
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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
	if kind := q.expectElemTyp.Kind(); kind != reflect.Chan && kind != reflect.Slice && kind != reflect.Array {
		panic(q.expectElemTyp.String() + " can not be flatten")
	}
	return newStream(q.expectElemTyp.Elem(), q.iter, func(outernext iterator) iterator {
		var innernext iterator
		var inner reflect.Value
		return func() (item reflect.Value, ok bool) {
			for !ok {
				if !inner.IsValid() {
					inner, ok = outernext()
					if !ok {
						return
					}
					_, innernext = makeIter(inner)
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
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if val, ok := next(); ok && fnval.Call([]reflect.Value{val})[0].Bool() {
				return val, true
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) Skip(size int) Stream {
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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
	return newStream(q.expectElemTyp, func() (reflect.Value, bool) {
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

func (q *stream) Uniq() Stream {
	var iter iterator
	return newStream(q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			v := Value{
				typ: reflect.SliceOf(q.expectElemTyp),
				val: reflect.Zero(reflect.SliceOf(q.expectElemTyp)),
			}
			dup := make(map[interface{}]struct{})
			for {
				val, ok := q.iter()
				if !ok {
					break
				}
				key := val.Interface()
				if _, ok := dup[key]; !ok {
					v.val = reflect.Append(v.val, val)
				}
				dup[key] = struct{}{}
			}
			_, iter = makeIter(v.val)
		}
		return iter()
	})
}

func (q *stream) Union(other Stream) Stream {
	oNext := other.ToSource().Next
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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

func (q *stream) Interact(other Stream) Stream {
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

func (q *stream) ToSetBy(fn interface{}) KVStream {
	fntyp := reflect.TypeOf(fn)
	fnval := reflect.ValueOf(fn)

	iter := q.iter
	return newKvStream(fntyp.Out(0), q.expectElemTyp, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(fntyp.Out(0), q.expectElemTyp))
		for {
			val, ok := iter()
			if !ok {
				break
			}
			table.SetMapIndex(fnval.Call([]reflect.Value{val})[0], val)
		}
		return table
	})
}

func (q *stream) ToSet() KVStream {
	iter := q.iter
	return newKvStream(q.expectElemTyp, boolType, func() reflect.Value {
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
	return newStream(q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			v := Value{
				typ: reflect.SliceOf(q.expectElemTyp),
				val: reflect.Zero(reflect.SliceOf(q.expectElemTyp)),
			}
			getKey := reflect.ValueOf(fn)
			dup := make(map[interface{}]struct{})
			for {
				val, ok := q.iter()
				if !ok {
					break
				}
				key := getKey.Call([]reflect.Value{val})[0].Interface()
				if _, ok := dup[key]; !ok {
					v.val = reflect.Append(v.val, val)
				}
				dup[key] = struct{}{}
			}
			_, iter = makeIter(v.val)
		}
		return iter()
	})
}

func (q *stream) SortBy(fn interface{}) Stream {
	var iter iterator
	return newStream(q.expectElemTyp, func() (reflect.Value, bool) {
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
	return newStream(reflect.SliceOf(q.expectElemTyp), q.iter, func(next iterator) iterator {
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
	return newStream(reflect.SliceOf(q.expectElemTyp), q.iter, func(next iterator) iterator {
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

func (q *stream) ToSlice(dst interface{}) {
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(q.getValue(val.Elem()))
}

func (q *stream) GroupBy(fn interface{}) KVStream {
	keyTyp := reflect.TypeOf(fn).Out(0)
	valTyp := reflect.SliceOf(q.expectElemTyp)

	iter := q.iter
	return newKvStream(keyTyp, valTyp, func() reflect.Value {
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
		if !slice.IsValid() || slice.Len() > 0 {
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

func (q *stream) flatMapBoolean(fnTyp reflect.Type, fnVal reflect.Value) *stream {
	return newStream(fnTyp.Out(0), q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for {
				val, ok := next()
				if !ok {
					break
				}
				out := fnVal.Call([]reflect.Value{val})
				if out[1].Bool() {
					return out[0], true
				}
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) flatMapError(fnTyp reflect.Type, fnVal reflect.Value) *stream {
	return newStream(fnTyp.Out(0), q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for {
				val, ok := next()
				if !ok {
					break
				}
				out := fnVal.Call([]reflect.Value{val})
				if err := out[1].Interface(); err == nil || err.(error) == nil {
					return out[0], true
				}
			}
			return reflect.Value{}, false
		}
	})
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
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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
	return newStream(q.expectElemTyp, q.iter, func(next iterator) iterator {
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

func (rv Value) To(dst interface{}) {
	if !rv.val.IsValid() {
		return
	}
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(rv.val)
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
	val, ok := iter()
	if !ok {
		return iter
	}
	if f(val) {
		iter = repeatableIter(iter, f)
	}
	var flag int32
	return func() (reflect.Value, bool) {
		if atomic.CompareAndSwapInt32(&flag, 0, 1) {
			return val, true
		}
		return iter()
	}
}

func normalIter(iter iterator, f func(reflect.Value) bool) {
	for {
		val, ok := iter()
		if !ok {
			break
		}
		f(val)
	}
}
