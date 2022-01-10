package fp

import (
	"reflect"

	"math"
)

// NewCounter create int source with elements 0, 1, 2...count
func NewCounter(count int) Source {
	return &counterSource{from: 0, to: count}
}

// Times create int stream with elements 0, 1, 2...count
func Times(count int) Stream {
	return StreamOf(NewCounter(count))
}

// NewCounterRange create int source with elements from...to
func NewCounterRange(from, to int) Source {
	return &counterSource{from: from, to: to + 1}
}

// RangeStream create int stream with elements from...to
func RangeStream(from, to int) Stream {
	return StreamOf(NewCounterRange(from, to))
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

// NaturalNumbers stream with uint64 elements 0, 1, 2...
func NaturalNumbers() Stream {
	return StreamOfSource(&naturalNumSource{})
}

// Index stream with int elements 0, 1, 2...
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
