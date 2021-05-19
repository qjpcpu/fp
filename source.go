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

type IndexSource interface {
	Source
	Index(int) (reflect.Value, bool)
}

func makeList(typ reflect.Type, val reflect.Value) *list {
	_, l := makeListWithElemType(typ, val)
	return l
}

func makeListWithElemType(typ reflect.Type, val reflect.Value) (reflect.Type, *list) {
	if source, ok := val.Interface().(Source); ok && source != nil {
		return source.ElemType(), makeListBySource(source)
	}
	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		source := newSliceSource(typ.Elem(), val)
		return source.ElemType(), makeListBySource(source)
	case reflect.Chan:
		source := newChannelSource(typ.Elem(), val)
		return source.ElemType(), makeListBySource(source)
	}
	panic("not support " + typ.String())
}

func makeListBySource(source Source) *list {
	if is, ok := source.(IndexSource); ok {
		return makeListByIndexSource(is, 0)
	}
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
			return makeListBySource(source)
		})
		return el
	}
	return nil
}

func makeListByIndexSource(source IndexSource, index int) *list {
	if source != nil {
		el := emptyList()
		el.elem = func() *atom {
			if c, ok := source.Index(index); ok {
				return createAtom(c)
			}
			return nil
		}

		el.next = func() *list {
			return makeListByIndexSource(source, index+1)
		}
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

func (ss *sliceSource) Index(i int) (reflect.Value, bool) {
	if i >= ss.arr.Len() {
		return reflect.Value{}, false
	}
	return ss.arr.Index(i), true
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
