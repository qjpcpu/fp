package fp

import "reflect"

func (q *stream) Partition(size int) Stream {
	if size < 1 {
		panic("batch size should be greater than 0")
	}
	return newStream(newCtx(q.ctx), reflect.SliceOf(q.expectElemTyp), q.iter, func(next iterator) iterator {
		typ := reflect.SliceOf(q.expectElemTyp)
		return func() (reflect.Value, bool) {
			var slice reflect.Value
			for i := 0; i < size; i++ {
				if val, ok := next(); !ok {
					break
				} else {
					if !slice.IsValid() {
						slice = reflect.Zero(typ)
					}
					slice = reflect.Append(slice, val)
				}
			}
			return slice, slice.IsValid() && slice.Len() > 0
		}
	})
}

func (q *stream) PartitionBy(fn interface{}, includeSplittor bool) Stream {
	fnval := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), reflect.SliceOf(q.expectElemTyp), q.iter, func(next iterator) iterator {
		typ := reflect.SliceOf(q.expectElemTyp)
		return func() (reflect.Value, bool) {
			var slice reflect.Value
			for {
				if val, ok := next(); !ok {
					break
				} else {
					if !slice.IsValid() {
						slice = reflect.Zero(typ)
					}
					if fnval.Call([]reflect.Value{val})[0].Bool() {
						if includeSplittor {
							slice = reflect.Append(slice, val)
						}
						break
					}
					slice = reflect.Append(slice, val)
				}
			}
			return slice, slice.IsValid() && slice.Len() > 0
		}
	})
}
