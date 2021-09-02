package fp

import "reflect"

func (q *stream) Uniq() Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			dup := make(map[interface{}]struct{})
			iter = func() (reflect.Value, bool) {
				for {
					val, ok := q.iter()
					if !ok {
						return val, false
					}
					key := val.Interface()
					if _, ok := dup[key]; !ok {
						dup[key] = struct{}{}
						return val, true
					}
				}
			}
		}
		return iter()
	})
}

func (q *stream) UniqBy(fn interface{}) Stream {
	var iter iterator
	return newStream(newCtx(q.ctx), q.expectElemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			getKey := reflect.ValueOf(fn)
			dup := make(map[interface{}]struct{})
			iter = func() (reflect.Value, bool) {
				for {
					val, ok := q.iter()
					if !ok {
						return val, false
					}
					key := getKey.Call([]reflect.Value{val})[0].Interface()
					if _, ok := dup[key]; !ok {
						dup[key] = struct{}{}
						return val, true
					}
				}
			}
		}
		return iter()
	})
}
