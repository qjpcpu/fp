package fp

import (
	"reflect"

	"math"
)

func NewCounter(count int) Source {
	return &counterSource{from: 0, to: count}
}

func NewCounterRange(from, to int) Source {
	return &counterSource{from: from, to: to + 1}
}

type counterSource struct {
	from, to int
}

func (cs *counterSource) ElemType() reflect.Type { return reflect.TypeOf(0) }
func (cs *counterSource) Next() (reflect.Value, bool) {
	if cs.from < cs.to {
		from := cs.from
		cs.from++
		return reflect.ValueOf(from), true
	}
	return reflect.Value{}, false
}

func NaturalNumbers() Stream {
	return StreamOfSource(&naturalNumSource{})
}

func Index() Stream {
	return StreamOfSource(&naturalNumSource{}).Map(func(i uint64) int { return int(i) })
}

type naturalNumSource struct {
	i uint64
}

func (cs *naturalNumSource) ElemType() reflect.Type { return reflect.TypeOf(uint64(0)) }
func (cs *naturalNumSource) Next() (reflect.Value, bool) {
	if cs.i < math.MaxUint64 {
		from := cs.i
		cs.i++
		return reflect.ValueOf(from), true
	}
	return reflect.Value{}, false
}
