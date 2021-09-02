package fp

import (
	"reflect"
	"sync/atomic"
)

func (q *stream) Prepend(v ...interface{}) Stream {
	nq := q
	for i := len(v) - 1; i >= 0; i-- {
		nq = nq.prependOne(v[i])
	}
	return nq
}

func (q *stream) prependOne(v interface{}) *stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var flag int32
		return func() (reflect.Value, bool) {
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				return reflect.ValueOf(v), true
			}
			return next()
		}
	})
}
