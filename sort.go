package fp

import (
	"reflect"
	"sort"
)

func (q *stream) Sort() Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			arr := q.getResult().Result()
			v := reflect.ValueOf(arr)
			sort.SliceStable(arr, func(i, j int) bool {
				return q.compare(v.Index(i), v.Index(j)) < 0
			})
			val := Value{
				typ: reflect.TypeOf(arr),
				val: reflect.ValueOf(arr),
			}
			_, iter = makeIter(val.val)
		}
		return iter()
	})
}

func (q *stream) SortBy(fn interface{}) Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			arr := q.getResult().Result()
			v := reflect.ValueOf(arr)
			fnval := reflect.ValueOf(fn)
			sort.SliceStable(arr, func(i, j int) bool {
				return fnval.Call([]reflect.Value{v.Index(i), v.Index(j)})[0].Bool()
			})
			val := Value{
				typ: reflect.TypeOf(arr),
				val: reflect.ValueOf(arr),
			}
			_, iter = makeIter(val.val)
		}
		return iter()
	})
}
