package fp

import "reflect"

func (q *stream) Take(size int) Stream {
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if size > 0 {
				if val, ok := next(); ok {
					size--
					return val, true
				}
			}
			return reflect.Value{}, false
		}
	})
}

func (q *stream) TakeWhile(fn interface{}) Stream {
	fnval := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			if val, ok := next(); ok && fnval.Call([]reflect.Value{val})[0].Bool() {
				return val, true
			}
			return reflect.Value{}, false
		}
	})
}
