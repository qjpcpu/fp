package fp

import (
	"reflect"
)

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

func (q *stream) LPartitionBy(fn interface{}, includeSplittor bool) Stream {
	if !includeSplittor {
		return q.PartitionBy(fn, includeSplittor)
	}
	fnval := reflect.ValueOf(fn)
	return newStream(newCtx(q.ctx), reflect.SliceOf(q.expectElemTyp), q.iter, func(next iterator) iterator {
		typ := reflect.SliceOf(q.expectElemTyp)
		var lastHead reflect.Value
		return func() (reflect.Value, bool) {
			slice := reflect.Zero(typ)
			if lastHead.IsValid() {
				slice = reflect.Append(slice, lastHead)
				lastHead = reflect.Value{}
			}
			for {
				if val, ok := next(); !ok {
					break
				} else {
					if fnval.Call([]reflect.Value{val})[0].Bool() {
						if slice.Len() == 0 {
							slice = reflect.Append(slice, val)
							continue
						}
						lastHead = val
						break
					}
					slice = reflect.Append(slice, val)
				}
			}
			return slice, slice.IsValid() && slice.Len() > 0
		}
	})
}
