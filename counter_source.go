package fp

import (
	"reflect"
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
