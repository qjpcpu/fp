package fp

import "reflect"

func (q *stream) Union(other Stream) Stream {
	if isNilStream(other) {
		return q
	}
	oNext := other.ToSource().Next
	return newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		var otherDone bool
		return func() (reflect.Value, bool) {
			if !otherDone {
				val, ok := next()
				if ok {
					return val, true
				}
				otherDone = true
			}
			return oNext()
		}
	})
}
