package fp

import "reflect"

func (q *stream) Reverse() Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			arr := q.getResult().Result()
			v := reflect.ValueOf(arr)
			idx := v.Len() - 1
			iter = func() (reflect.Value, bool) {
				if idx >= 0 {
					idx--
					return v.Index(idx + 1), true
				}
				return reflect.Value{}, false
			}
		}
		return iter()
	})
}
