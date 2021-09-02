package fp

import "reflect"

func (q *stream) First() (f Value) {
	f.typ = q.expectElemTyp
	q.iter = repeatableIter(q.iter, func(v reflect.Value) bool {
		f.val = v
		return false
	})
	return f
}
