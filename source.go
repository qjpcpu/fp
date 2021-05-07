package fp

import (
	"bufio"
	"io"
	"reflect"
	"sync"
)

func StreamOf(arr interface{}) Stream {
	elemTyp, list := makeListWithElemType(reflect.TypeOf(arr), reflect.ValueOf(arr))
	return newStream(elemTyp, list)
}

func StreamOfSource(s Source) Stream {
	return newStream(s.ElemType(), makeListBySource(s.ElemType(), s))
}

type Source interface {
	ElemType() reflect.Type
	CAR() reflect.Value
	CDR() Source
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
		el.elem = carOnce(func() *cell {
			if c := source.CAR(); c.IsValid() {
				return createCell(elemTyp, c)
			}
			return nil
		})

		el.next = cdrOnce(func() *list {
			return makeListBySource(elemTyp, source.CDR())
		})
		return el
	}
	return nil
}

/* slice stream */
type sliceSource struct {
	elemType reflect.Type
	arr      reflect.Value
}

func newSliceSource(elemTyp reflect.Type, arr reflect.Value) Source {
	return &sliceSource{
		elemType: elemTyp,
		arr:      arr,
	}
}

func (ss *sliceSource) ElemType() reflect.Type { return ss.elemType }

func (ss *sliceSource) CAR() reflect.Value {
	if ss.arr.Len() == 0 {
		return reflect.Value{}
	}
	return ss.arr.Index(0)
}

func (ss *sliceSource) CDR() Source {
	if ss.arr.Len() <= 1 {
		return nil
	}
	return &sliceSource{
		elemType: ss.elemType,
		arr:      ss.arr.Slice(1, ss.arr.Len()),
	}
}

/* channel stream */
type channelSource struct {
	elemType reflect.Type
	once     *sync.Once
	ch       reflect.Value
	first    reflect.Value
}

func newChannelSource(elemTyp reflect.Type, ch reflect.Value) Source {
	return &channelSource{
		elemType: elemTyp,
		once:     new(sync.Once),
		ch:       ch,
		first:    reflect.Value{},
	}
}

func (cs *channelSource) ElemType() reflect.Type {
	return cs.elemType
}

func (cs *channelSource) CAR() reflect.Value {
	cs.once.Do(func() {
		if _, recv, ok := reflect.Select([]reflect.SelectCase{
			{
				Dir:  reflect.SelectRecv,
				Chan: cs.ch,
			},
		}); ok {
			cs.first = recv
		}
	})
	return cs.first
}

func (cs *channelSource) CDR() Source {
	if v := cs.CAR(); !v.IsValid() {
		return nil
	}
	return &channelSource{
		elemType: cs.elemType,
		once:     new(sync.Once),
		ch:       cs.ch,
		first:    reflect.Value{},
	}
}

type lineSource struct {
	s       *bufio.Scanner
	elemTyp reflect.Type
	line    reflect.Value
	read    sync.Once
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

func (ls *lineSource) CAR() reflect.Value {
	ls.read.Do(func() {
		if ls.s.Scan() {
			ls.line = reflect.ValueOf(ls.s.Text())
		}
	})
	return ls.line
}

func (ls *lineSource) CDR() Source {
	if v := ls.CAR(); !v.IsValid() {
		return nil
	}
	return &lineSource{
		s:       ls.s,
		elemTyp: ls.elemTyp,
	}
}
