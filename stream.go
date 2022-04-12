package fp

import (
	"reflect"
	"sync"
)

type Stream interface {
	// Map stream to another, fn should be func(element_type) (another_type,&optional error/bool)
	Map(fn interface{}) Stream
	// FlatMap stream to another, fn should be func(element_type) (slice_type,&optional error)
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
	// Reduce0 stream, fn should be func(accumulative element_type,item element_type) element_type
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
	// Interact stream, keep element on left
	Interact(other Stream) Stream
	// InteractBy keyfn, keyfn is func(element_type) any_type, keep element on left
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
	// Error first error
	Error() error
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
	ctx := newCtx(nil)
	elemTyp, it := makeIter(ctx, reflect.ValueOf(arr))
	return newStream(ctx, elemTyp, it)
}

func Stream0Of(arr ...interface{}) Stream {
	length := len(arr)
	if length > 1 && arr[length-1] != nil && reflect.TypeOf(arr[length-1]).AssignableTo(errType) && arr[length-1].(error) != nil {
		ctx := newCtx(nil)
		/* arr[length-1] is error */
		ctx.SetErr(arr[length-1].(error))
		elemTyp, it := makeIter(ctx, reflect.ValueOf(arr[0]))
		return newStream(ctx, elemTyp, it)
	}
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
