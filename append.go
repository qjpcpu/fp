package fp

import (
	"reflect"
	"sync/atomic"
)

func (q *stream) Append(v ...interface{}) Stream {
	nq := q
	for _, elem := range v {
		nq = nq.appendOne(elem)
	}
	return nq
}

func (q *stream) appendOne(v interface{}) *stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var flag int32
		return func() (reflect.Value, bool) {
			if flag == 0 {
				if val, ok := next(); ok {
					return val, ok
				}
			}
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				return reflect.ValueOf(v), true
			}
			return reflect.Value{}, false
		}
	})
}
