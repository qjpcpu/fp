package fp

import "reflect"

func (q *stream) Size() int {
	return q.getValue(reflect.Value{}).Len()
}
