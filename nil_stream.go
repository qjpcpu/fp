package fp

import "reflect"

type nilStream struct{}

func (ns *nilStream) Map(fn interface{}) Stream     { return ns }
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
func (ns *nilStream) Partition(size int) Stream       { return ns }
func (ns *nilStream) First() Value                    { return Value{} }
func (ns *nilStream) IsEmpty() bool                   { return true }
func (ns *nilStream) Take(n int) Stream               { return ns }
func (ns *nilStream) Skip(size int) Stream            { return ns }
func (ns *nilStream) Sort() Stream                    { return ns }
func (ns *nilStream) SortBy(fn interface{}) Stream    { return ns }
func (ns *nilStream) Uniq() Stream                    { return ns }
func (ns *nilStream) UniqBy(fn interface{}) Stream    { return ns }
func (ns *nilStream) Size() int                       { return 0 }
func (ns *nilStream) Contains(interface{}) bool       { return false }
func (ns *nilStream) ToSource() Source                { return nil }
func (ns *nilStream) Join(s Stream) Stream            { return s }
func (ns *nilStream) GroupBy(fn interface{}) KVStream { return nil }
func (ns *nilStream) Append(v interface{}) Stream {
	typ, val := reflect.TypeOf(v), reflect.ValueOf(v)
	slice := reflect.MakeSlice(reflect.SliceOf(typ), 1, 1)
	slice.Index(0).Set(val)
	return newStream(typ, makeListBySource(typ, newSliceSource(typ, slice)))
}
func (ns *nilStream) Prepend(element interface{}) Stream { return ns.Append(element) }
func (ns *nilStream) Run()                               {}
func (ns *nilStream) ToSlice(ptr interface{})            {}
func (ns *nilStream) Result() Value                      { return Value{} }
