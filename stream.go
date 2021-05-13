package fp

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type Stream interface {
	// Map stream to another, fn should be func(element_type) another_type
	Map(fn interface{}) Stream
	// Filter stream, fn should be func(element_type) bool
	Filter(fn interface{}) Stream
	// Reject stream, fn should be func(element_type) bool
	Reject(fn interface{}) Stream
	// Foreach stream element, fn should be func(element_type)
	Foreach(fn interface{}) Stream
	// Flatten stream, element should be flatten-able
	Flatten() Stream
	// Reduce stream, fn should be func(initval_type, element_type) initval_type
	Reduce(initval interface{}, fn interface{}) Value
	// Partition stream, split stream into small batch
	Partition(size int) Stream
	// PartitionBy func(elem_type) bool
	PartitionBy(fn interface{}, includeSplittor bool) Stream
	// First value of stream
	First() Value
	// IsEmpty stream
	IsEmpty() bool
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
	// Uniq stream, this is an aggregate op, so it would block stream
	Uniq() Stream
	// UniqBy stream, fn should be func(element_type) any_type, this is an aggregate op, so it would block stream
	UniqBy(fn interface{}) Stream
	// Size of stream, this is an aggregate op, so it would block stream
	Size() int
	// Contains element
	Contains(interface{}) bool
	// ToSource convert stream to source
	ToSource() Source
	// Join after another stream
	Join(Stream) Stream
	// GroupBy func(element_type) any_type, this is an aggregate op, so it would block stream
	GroupBy(fn interface{}) KVStream
	// Append element
	Append(element interface{}) Stream
	// Prepend element
	Prepend(element interface{}) Stream

	// Run stream and drop value
	Run()
	// ToSlice ptr
	ToSlice(ptr interface{})
	// Result of stream
	Result() Value
}

var streamType = reflect.TypeOf((*Stream)(nil)).Elem()

type stream struct {
	expectElemTyp reflect.Type
	list          *list
	val           reflect.Value
	getValOnce    sync.Once
}

func newStream(expTyp reflect.Type, list *list) *stream {
	return &stream{expectElemTyp: expTyp, list: list}
}

func (q *stream) Map(fn interface{}) Stream {
	fnVal := reflect.ValueOf(fn)
	return newStream(fnVal.Type().Out(0), mapcar(fnVal, q.list))
}

func (q *stream) Filter(fn interface{}) Stream {
	fnVal := reflect.ValueOf(fn)
	return newStream(q.expectElemTyp, selectcar(fnVal, q.list))
}

func (q *stream) Flatten() Stream {
	if q.expectElemTyp == streamType {
		/* sadly, i can't know the detail element type cause of our lazy evaluation */
		/* i have to car for the first element to get elem type */
		/* well, it's not lazy enough here */
		if l := flatten(q.list); l == nil {
		} else if e := car(l); e != nil {
			return newStream(e.typ, l)
		}
		return &nilStream{}
	}
	if kind := q.expectElemTyp.Kind(); kind != reflect.Chan && kind != reflect.Slice && kind != reflect.Array {
		panic(q.expectElemTyp.String() + " can not be flatten")
	}
	return newStream(q.expectElemTyp.Elem(), flatten(q.list))
}

func (q *stream) Foreach(fn interface{}) Stream {
	return newStream(q.expectElemTyp, foreachcar(reflect.ValueOf(fn), q.list))
}

func (q *stream) Prepend(v interface{}) Stream {
	typ, val := reflect.TypeOf(v), reflect.ValueOf(v)
	if typ != q.expectElemTyp {
		panic("bad element type")
	}
	old := q.list
	l := cons(
		func() *atom {
			return createAtom(typ, val)
		},
		func() *list {
			return old
		})
	return newStream(q.expectElemTyp, l)
}

func (q *stream) Append(v interface{}) Stream {
	typ, val := reflect.TypeOf(v), reflect.ValueOf(v)
	if typ != q.expectElemTyp {
		panic("bad element type")
	}
	slice := reflect.MakeSlice(reflect.SliceOf(typ), 1, 1)
	slice.Index(0).Set(val)
	return newStream(q.expectElemTyp, concat(q.list, makeListBySource(typ, newSliceSource(typ, slice))))
}

func (q *stream) IsEmpty() bool {
	return car(q.list) == nil
}

func (q *stream) Take(size int) Stream {
	return newStream(q.expectElemTyp, takecar(size, q.list))
}

func (q *stream) TakeWhile(fn interface{}) Stream {
	return newStream(q.expectElemTyp, takeWhile(reflect.ValueOf(fn), q.list))
}

func (q *stream) Skip(size int) Stream {
	return newStream(q.expectElemTyp, skipcar(size, q.list))
}

func (q *stream) SkipWhile(fn interface{}) Stream {
	return newStream(q.expectElemTyp, skipWhile(reflect.ValueOf(fn), q.list))
}

func (q *stream) First() Value {
	if isNil(q.list) {
		return Value{
			typ: q.expectElemTyp,
			val: reflect.Zero(q.expectElemTyp),
		}
	}
	return valueOfCell(car(q.list))
}

func (q *stream) Sort() Stream {
	if isNil(q.list) {
		val := Value{
			typ: q.expectElemTyp,
			val: reflect.Zero(q.expectElemTyp),
		}
		elemTyp, list := makeListWithElemType(val.typ, val.val)
		return newStream(elemTyp, list)
	}
	arr := q.Result().Interface()
	v := reflect.ValueOf(arr)
	sort.SliceStable(arr, func(i, j int) bool {
		return q.compare(v.Index(i), v.Index(j)) < 0
	})
	val := Value{
		typ: reflect.TypeOf(arr),
		val: reflect.ValueOf(arr),
	}
	elemTyp, list := makeListWithElemType(val.typ, val.val)
	return newStream(elemTyp, list)
}

func (q *stream) Uniq() Stream {
	v := Value{
		typ: reflect.SliceOf(q.expectElemTyp),
		val: reflect.Zero(reflect.SliceOf(q.expectElemTyp)),
	}
	dup := make(map[interface{}]struct{})
	processList(q.list, func(e *atom) bool {
		key := e.val.Interface()
		if _, ok := dup[key]; !ok {
			v.val = reflect.Append(v.val, e.val)
		}
		dup[key] = struct{}{}
		return true
	})
	elemTyp, list := makeListWithElemType(v.typ, v.val)
	return newStream(elemTyp, list)
}

func (q *stream) Join(other Stream) Stream {
	var sourcelist *list
	if s, ok := other.(*stream); ok {
		if s.expectElemTyp != q.expectElemTyp {
			panic("different stream type")
		}
		sourcelist = s.list
	} else if _, ok := other.(*nilStream); ok {
		return q
	} else {
		source := other.ToSource()
		if source.ElemType() != q.expectElemTyp {
			panic("different stream type")
		}
		sourcelist = makeListBySource(source.ElemType(), source)
	}
	return newStream(q.expectElemTyp, concat(sourcelist, q.list))
}

func (q *stream) UniqBy(fn interface{}) Stream {
	v := Value{
		typ: reflect.SliceOf(q.expectElemTyp),
		val: reflect.Zero(reflect.SliceOf(q.expectElemTyp)),
	}
	getKey := reflect.ValueOf(fn)
	dup := make(map[interface{}]struct{})
	processList(q.list, func(e *atom) bool {
		key := getKey.Call([]reflect.Value{e.val})[0].Interface()
		if _, ok := dup[key]; !ok {
			v.val = reflect.Append(v.val, e.val)
		}
		dup[key] = struct{}{}
		return true
	})
	elemTyp, list := makeListWithElemType(v.typ, v.val)
	return newStream(elemTyp, list)
}

func (q *stream) SortBy(fn interface{}) Stream {
	if isNil(q.list) {
		val := Value{
			typ: q.expectElemTyp,
			val: reflect.Zero(q.expectElemTyp),
		}
		elemTyp, list := makeListWithElemType(val.typ, val.val)
		return newStream(elemTyp, list)
	}
	arr := q.Result().Interface()
	v := reflect.ValueOf(arr)
	fnval := reflect.ValueOf(fn)
	sort.SliceStable(arr, func(i, j int) bool {
		return fnval.Call([]reflect.Value{v.Index(i), v.Index(j)})[0].Bool()
	})
	val := Value{
		typ: reflect.TypeOf(arr),
		val: reflect.ValueOf(arr),
	}
	elemTyp, list := makeListWithElemType(val.typ, val.val)
	return newStream(elemTyp, list)
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
	processList(q.list, func(cell *atom) bool {
		memo = fnval.Call([]reflect.Value{memo, cell.val})[0]
		return true
	})
	return Value{typ: typ, val: memo}
}

func (q *stream) Partition(size int) Stream {
	if size < 1 {
		panic("batch size should be greater than 0")
	}
	return newStream(reflect.SliceOf(q.expectElemTyp), partitoncar(size, q.list))
}

func (q *stream) PartitionBy(fn interface{}, includeSplittor bool) Stream {
	return newStream(reflect.SliceOf(q.expectElemTyp), partitoncarby(reflect.ValueOf(fn), includeSplittor, q.list))
}

func (q *stream) ToSlice(dst interface{}) {
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(q.getValue())
}

func (q *stream) GroupBy(fn interface{}) KVStream {
	keyTyp := reflect.TypeOf(fn).Out(0)
	valTyp := reflect.SliceOf(q.expectElemTyp)
	table := reflect.MakeMap(reflect.MapOf(keyTyp, valTyp))

	fnVal := reflect.ValueOf(fn)
	processList(q.list, func(cell *atom) bool {
		key := fnVal.Call([]reflect.Value{cell.val})[0]
		slice := table.MapIndex(key)
		if !slice.IsValid() {
			slice = reflect.Zero(valTyp)
		}
		slice = reflect.Append(slice, cell.val)
		table.SetMapIndex(key, slice)
		return true
	})
	return KVStreamOf(table.Interface())
}

func (q *stream) Result() Value {
	return Value{
		typ: reflect.SliceOf(q.expectElemTyp),
		val: q.getValue(),
	}
}

func (q *stream) Size() int {
	return q.getValue().Len()
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
	processList(q.list, func(cell *atom) bool {
		yes = eq(cell.val)
		return !yes
	})
	return
}

func (q *stream) Run() {
	q.getValOnce.Do(func() {
		/* let gc work */
		tmp := q.list
		q.list = nil
		processList(tmp, nil)
	})
}

/* stream to source */
func (q *stream) ToSource() Source {
	return q
}

func (q *stream) ElemType() reflect.Type {
	return q.expectElemTyp
}

func (q *stream) CAR() reflect.Value {
	if elem := car(q.list); elem == nil {
		return reflect.Value{}
	} else {
		return elem.val
	}
}

func (q *stream) CDR() Source {
	return newStream(q.expectElemTyp, cdr(q.list))
}

func (q *stream) getValue() reflect.Value {
	q.getValOnce.Do(func() {
		/* let gc work */
		tmp := q.list
		q.list = nil
		q.val = asSlice(q.expectElemTyp, tmp)
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

/* value related */
type Value struct {
	typ reflect.Type
	val reflect.Value
}

func valueOfCell(e *atom) Value {
	return Value{
		typ: e.typ,
		val: e.val,
	}
}

func (rv Value) Result(dst interface{}) {
	if !rv.val.IsValid() {
		return
	}
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(rv.val)
}

func (rv Value) Interface() interface{} {
	if !rv.val.IsValid() {
		return nil
	}
	return rv.val.Interface()
}

func (rv Value) String() (s string) {
	rv.Result(&s)
	return
}

func (rv Value) Strings() (s []string) {
	rv.Result(&s)
	return
}

func (rv Value) Bytes() (s []byte) {
	rv.Result(&s)
	return
}

func (rv Value) Ints() (s []int) {
	rv.Result(&s)
	return
}

func (rv Value) Int64s() (s []int64) {
	rv.Result(&s)
	return
}

func (rv Value) Int32s() (s []int32) {
	rv.Result(&s)
	return
}

func (rv Value) Uints() (s []uint) {
	rv.Result(&s)
	return
}

func (rv Value) Uint32s() (s []uint32) {
	rv.Result(&s)
	return
}

func (rv Value) Uint64s() (s []uint64) {
	rv.Result(&s)
	return
}

func (rv Value) StringsList() (s [][]string) {
	rv.Result(&s)
	return
}
