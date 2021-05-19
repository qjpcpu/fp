package fp

import (
	"bufio"
	"io"
	"reflect"
)

func StreamOf(arr interface{}) Stream {
	elemTyp, list := makeListWithElemType(reflect.TypeOf(arr), reflect.ValueOf(arr))
	return newStream(elemTyp, list)
}

func StreamOfSource(s Source) Stream {
	if s == nil {
		return newNilStream()
	}
	return newStream(s.ElemType(), makeListBySource(s.ElemType(), s))
}

type Source interface {
	// ElemType element type
	ElemType() reflect.Type
	// Next element
	Next() (reflect.Value, bool)
}

func makeList(typ reflect.Type, val reflect.Value) *list {
	_, l := makeListWithElemType(typ, val)
	return l
}

func makeListWithElemType(typ reflect.Type, val reflect.Value) (reflect.Type, *list) {
	if source, ok := val.Interface().(Source); ok && source != nil {
		return source.ElemType(), makeListBySource(source.ElemType(), source)
	}
	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		source := newSliceSource(typ.Elem(), val)
		return source.ElemType(), makeListBySource(source.ElemType(), source)
	case reflect.Chan:
		source := newChannelSource(typ.Elem(), val)
		return source.ElemType(), makeListBySource(source.ElemType(), source)
	}
	panic("not support " + typ.String())
}

func makeListBySource(elemTyp reflect.Type, source Source) *list {
	if source != nil {
		el := emptyList()
		carfn := carOnce(func() *atom {
			if c, ok := source.Next(); ok {
				return createAtom(c)
			}
			return nil
		})
		el.elem = carfn

		el.next = cdrOnce(func() *list {
			_ = carfn()
			return makeListBySource(elemTyp, source)
		})
		return el
	}
	return nil
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
