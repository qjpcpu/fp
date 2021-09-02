package fp

import (
	"reflect"
	"sync/atomic"
)

func (q *stream) Skip(size int) Stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for ; size > 0; size-- {
				if _, ok := next(); !ok {
					return reflect.Value{}, false
				}
			}
			return next()
		}
	})
}

func (q *stream) SkipWhile(fn interface{}) Stream {
	fnval := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var flag int32
		return func() (reflect.Value, bool) {
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				for {
					val, ok := next()
					if !ok {
						return reflect.Value{}, false
					}
					if !fnval.Call([]reflect.Value{val})[0].Bool() {
						return val, true
					}
				}
			}
			return next()
		}
	})
}
