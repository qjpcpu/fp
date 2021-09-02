package fp

import "reflect"

func (q *stream) GroupBy(fn interface{}) KVStream {
	keyTyp := reflect.TypeOf(fn).Out(0)
	valTyp := reflect.SliceOf(q.expectElemTyp)

	iter := q.iter
	return newKvStream(newCtx(q.ctx), keyTyp, valTyp, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(keyTyp, valTyp))
		fnVal := reflect.ValueOf(fn)
		for {
			val, ok := iter()
			if !ok {
				break
			}
			key := fnVal.Call([]reflect.Value{val})[0]
			slice := table.MapIndex(key)
			if !slice.IsValid() {
				slice = reflect.Zero(valTyp)
			}
			slice = reflect.Append(slice, val)
			table.SetMapIndex(key, slice)
		}
		return table
	})
}
