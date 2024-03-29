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

type KVSource interface {
	ElemType() (reflect.Type, reflect.Type)
	Next() (reflect.Value, reflect.Value, bool)
}

func makeIter(ctx context, val reflect.Value) (reflect.Type, iterator) {
	typ := val.Type()
	if source, ok := val.Interface().(Source); ok && source != nil {
		return source.ElemType(), source.Next
	}
	if isIterFunction(val) {
		return val.Type().Out(0), func() (reflect.Value, bool) {
			out := val.Call(nil)
			return out[0], out[1].Bool()
		}
	} else if isIterFunction2(val) {
		return val.Type().Out(0), func() (reflect.Value, bool) {
			out := val.Call(nil)
			if _err := out[2].Interface(); _err != nil && _err.(error) != nil {
				ctx.SetErr(_err.(error))
				return reflect.Value{}, false
			}
			return out[0], out[1].Bool()
		}
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

func isIterFunction(fn reflect.Value) bool {
	typ := fn.Type()
	return typ.Kind() == reflect.Func && typ.NumIn() == 0 && typ.NumOut() == 2 && typ.Out(1) == boolType
}

func isIterFunction2(fn reflect.Value) bool {
	typ := fn.Type()
	return typ.Kind() == reflect.Func && typ.NumIn() == 0 && typ.NumOut() == 3 && typ.Out(1) == boolType && typ.Out(2) == errType
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
