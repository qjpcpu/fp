package fp

import (
	"bufio"
	"io"
	"reflect"
)

type Source interface {
	// ElemType element type
	ElemType() reflect.Type
	// Next element
	Next() (reflect.Value, bool)
}

func makeList(val reflect.Value) (reflect.Type, iterator) {
	typ := val.Type()
	if source, ok := val.Interface().(Source); ok && source != nil {
		return source.ElemType(), source.Next
	}
	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		source := newSliceSource(typ.Elem(), val)
		return source.ElemType(), source.Next
	case reflect.Chan:
		source := newChannelSource(typ.Elem(), val)
		return source.ElemType(), source.Next
	}
	panic("not support " + typ.String())
}

/* slice stream */
type sliceSource struct {
	elemType reflect.Type
	arr      reflect.Value
	offset   int
}

func newSliceSource(elemTyp reflect.Type, arr reflect.Value) Source {
	return &sliceSource{
		elemType: elemTyp,
		arr:      arr,
	}
}

func (ss *sliceSource) ElemType() reflect.Type { return ss.elemType }

func (ss *sliceSource) Next() (reflect.Value, bool) {
	if ss.offset >= ss.arr.Len() {
		return reflect.Value{}, false
	}
	offset := ss.offset
	ss.offset++
	return ss.arr.Index(offset), true
}

/* channel stream */
type channelSource struct {
	elemType reflect.Type
	ch       reflect.Value
}

func newChannelSource(elemTyp reflect.Type, ch reflect.Value) Source {
	return &channelSource{
		elemType: elemTyp,
		ch:       ch,
	}
}

func (cs *channelSource) ElemType() reflect.Type {
	return cs.elemType
}

func (cs *channelSource) Next() (reflect.Value, bool) {
	if _, recv, ok := reflect.Select([]reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: cs.ch,
		},
	}); ok {
		return recv, true
	}
	return reflect.Value{}, false
}

type lineSource struct {
	s       *bufio.Scanner
	elemTyp reflect.Type
}

func NewLineSource(r io.Reader) Source {
	return &lineSource{
		s:       bufio.NewScanner(r),
		elemTyp: reflect.TypeOf(""),
	}
}

func (ls *lineSource) ElemType() reflect.Type {
	return ls.elemTyp
}

func (ls *lineSource) Next() (reflect.Value, bool) {
	if ls.s.Scan() {
		return reflect.ValueOf(ls.s.Text()), true
	}
	return reflect.Value{}, false
}
