package fp

import (
	"fmt"
	"reflect"
)

func (q *stream) Zip(other Stream, fn interface{}) Stream {
	if isNilStream(other) {
		return other
	}
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	onext := other.ToSource().Next
	return newStream(newCtx(q.ctx), fnTyp.Out(0), q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if val1, ok1 := next(); ok1 {
				if val2, ok2 := onext(); ok2 {
					return fnVal.Call([]reflect.Value{val1, val2})[0], true
				}
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) ZipN(fn interface{}, others ...Stream) Stream {
	for _, s := range others {
		if isNilStream(s) {
			return s
		}
	}
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	if fnTyp.NumIn() != len(others)+1 {
		panic(fmt.Sprintf("zip function must have %v input param", len(others)+1))
	}

	return newStream(newCtx(q.ctx), fnTyp.Out(0), q.iter, func(next iterator) iterator {
		/* build iterator list */
		var iteratorList []iterator
		StreamOf(others).Map(func(s Stream) iterator {
			return s.ToSource().Next
		}).Prepend(next).
			ToSlice(&iteratorList)

		var done bool
		return func() (reflect.Value, bool) {
			if !done {
				input := make([]reflect.Value, len(others)+1)
				for i := range iteratorList {
					if val, ok := iteratorList[i](); ok {
						input[i] = val
					} else {
						done = true
						return reflect.Value{}, false
					}
				}
				return fnVal.Call(input)[0], true
			}
			return reflect.Value{}, false
		}
	})
}
