package fp

import "reflect"

func (q *stream) Filter(fn interface{}) Stream {
	fnVal := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for {
				if val, ok := next(); !ok {
					break
				} else if fnVal.Call([]reflect.Value{val})[0].Bool() {
					return val, true
				}
			}
			return reflect.Value{}, false
		}
	})
}
