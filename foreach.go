package fp

import "reflect"

func (q *stream) Foreach(fn interface{}) Stream {
	fnval := reflect.ValueOf(fn)
	withIndex := fnval.Type().NumIn() == 2
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var i int
		return func() (val reflect.Value, ok bool) {
			if val, ok = next(); ok {
				if withIndex {
					fnval.Call([]reflect.Value{val, reflect.ValueOf(i)})
					i++
				} else {
					fnval.Call([]reflect.Value{val})
				}
			}
			return
		}
	})
}
