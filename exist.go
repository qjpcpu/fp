package fp

import (
	"reflect"
	"sync/atomic"
)

func (q *stream) IsEmpty() bool {
	old := q.iter
	v, ok := q.iter()
	if ok {
		var flag int32
		q.iter = func() (reflect.Value, bool) {
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				return v, true
			}
			return old()
		}
	}
	return !ok
}

func (q *stream) HasSomething() bool {
	return !q.IsEmpty()
}

func (q *stream) Exists() bool {
	return !q.IsEmpty()
}
