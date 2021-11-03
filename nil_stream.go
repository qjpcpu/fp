package fp

import "reflect"

type nilStream struct{}

func newNilStream() Stream { return &nilStream{} }

func (ns *nilStream) Map(fn interface{}) Stream     { return ns }
func (ns *nilStream) FlatMap(fn interface{}) Stream { return ns }
func (ns *nilStream) Filter(fn interface{}) Stream  { return ns }
func (ns *nilStream) Reject(fn interface{}) Stream  { return ns }
func (ns *nilStream) Foreach(fn interface{}) Stream { return ns }
func (ns *nilStream) Flatten() Stream               { return ns }
func (ns *nilStream) Reduce(initval interface{}, fn interface{}) Value {
	return Value{
		typ: reflect.TypeOf(initval),
		val: reflect.ValueOf(initval),
	}
}
func (ns *nilStream) Reduce0(fn interface{}) Value {
	typ := reflect.TypeOf(fn).Out(0)
	return Value{
		typ: typ,
		val: reflect.Zero(typ),
	}
}
func (ns *nilStream) Partition(size int) Stream                               { return ns }
func (ns *nilStream) PartitionBy(fn interface{}, includeSplittor bool) Stream { return ns }
func (ns *nilStream) First() Value                                            { return Value{} }
func (ns *nilStream) IsEmpty() bool                                           { return true }
func (ns *nilStream) HasSomething() bool                                      { return false }
func (ns *nilStream) Exists() bool                                            { return false }
func (ns *nilStream) Take(n int) Stream                                       { return ns }
func (ns *nilStream) TakeWhile(fn interface{}) Stream                         { return ns }
func (ns *nilStream) Skip(size int) Stream                                    { return ns }
func (ns *nilStream) SkipWhile(fn interface{}) Stream                         { return ns }
func (ns *nilStream) Sort() Stream                                            { return ns }
func (ns *nilStream) SortBy(fn interface{}) Stream                            { return ns }
func (ns *nilStream) Uniq() Stream                                            { return ns }
func (ns *nilStream) UniqBy(fn interface{}) Stream                            { return ns }
func (ns *nilStream) Size() int                                               { return 0 }
func (ns *nilStream) Count() int                                              { return 0 }
func (ns *nilStream) Contains(interface{}) bool                               { return false }
func (ns *nilStream) ContainsBy(fn interface{}) bool                          { return false }
func (ns *nilStream) ToSource() Source                                        { return newNilSource() }
func (ns *nilStream) Sub(other Stream) Stream                                 { return ns }
func (ns *nilStream) SubBy(other Stream, keyfn interface{}) Stream            { return ns }
func (ns *nilStream) Interact(other Stream) Stream                            { return ns }
func (ns *nilStream) InteractBy(other Stream, keyfn interface{}) Stream       { return ns }
func (ns *nilStream) Union(o Stream) Stream                                   { return o }
func (ns *nilStream) ToSet() KVStream                                         { return newNilKVStream() }
func (ns *nilStream) ToSetBy(fn interface{}) KVStream                         { return newNilKVStream() }
func (ns *nilStream) GroupBy(fn interface{}) KVStream                         { return newNilKVStream() }
func (ns *nilStream) Reverse() Stream                                         { return ns }
func (ns *nilStream) Append(element ...interface{}) Stream {
	if len(element) == 0 {
		return ns
	}
	first := element[0]
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.ValueOf(first).Type()), len(element), len(element))
	for i := 0; i < len(element); i++ {
		slice.Index(i).Set(reflect.ValueOf(element[i]))
	}
	return StreamOf(slice.Interface())
}
func (ns *nilStream) Prepend(element ...interface{}) Stream        { return ns.Append(element...) }
func (ns *nilStream) Zip(other Stream, fn interface{}) Stream      { return ns }
func (ns *nilStream) ZipN(fn interface{}, others ...Stream) Stream { return ns }
func (ns *nilStream) Branch(processors ...StreamProcessor)         {}
func (ns *nilStream) Run()                                         {}
func (ns *nilStream) ToSlice(ptr interface{}) error {
	val := reflect.ValueOf(ptr)
	if elem := val.Elem(); elem.IsValid() && elem.Len() > 0 {
		elem.SetLen(0)
	}
	return nil
}

func (ns *nilStream) Strings() []string             { return nil }
func (ns *nilStream) StringsList() [][]string       { return nil }
func (ns *nilStream) Ints() []int                   { return nil }
func (ns *nilStream) Float64s() []float64           { return nil }
func (ns *nilStream) Bytes() []byte                 { return nil }
func (ns *nilStream) Int64s() []int64               { return nil }
func (ns *nilStream) Int32s() []int32               { return nil }
func (ns *nilStream) Uints() []uint                 { return nil }
func (ns *nilStream) Uint32s() []uint32             { return nil }
func (ns *nilStream) Uint64s() []uint64             { return nil }
func (ns *nilStream) JoinStrings(seq string) string { return "" }
func (ks *nilStream) Error() error                  { return nil }

type nilkvStream struct{}

func newNilKVStream() KVStream                            { return &nilkvStream{} }
func (ks *nilkvStream) Foreach(fn interface{}) KVStream   { return ks }
func (ks *nilkvStream) Map(fn interface{}) KVStream       { return ks }
func (ks *nilkvStream) MapToStream(fn interface{}) Stream { return newNilStream() }
func (ks *nilkvStream) Filter(fn interface{}) KVStream    { return ks }
func (ks *nilkvStream) Reject(fn interface{}) KVStream    { return ks }
func (ks *nilkvStream) Contains(key interface{}) bool     { return false }
func (ks *nilkvStream) Keys() Stream                      { return newNilStream() }
func (ks *nilkvStream) Values() Stream                    { return newNilStream() }
func (ks *nilkvStream) Size() int                         { return 0 }
func (ks *nilkvStream) Run()                              {}
func (ks *nilkvStream) To(dstPtr interface{}) error {
	val := reflect.ValueOf(dstPtr)
	if !val.Elem().IsValid() || val.Elem().IsNil() {
		val.Elem().Set(reflect.MakeMap(val.Elem().Type()))
	}
	return nil
}

type nilSource struct{}

func newNilSource() Source                        { return &nilSource{} }
func (ns *nilSource) ElemType() reflect.Type      { return reflect.TypeOf(nil) }
func (ns *nilSource) Next() (reflect.Value, bool) { return reflect.Value{}, false }

func isNilStream(s Stream) bool {
	if s == nil {
		return true
	}
	_, ok := s.(*nilStream)
	return ok
}
