package fp

import "reflect"

func (q *stream) Map(fn interface{}) Stream {
	fnTyp := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)
	ctx := newCtx(q.ctx)
	mapFn := func(in reflect.Value) (reflect.Value, bool) {
		return fnVal.Call([]reflect.Value{in})[0], true
	}
	if fnTyp.NumOut() == 2 && fnTyp.Out(1) == boolType {
		mapFn = func(in reflect.Value) (reflect.Value, bool) {
			out := fnVal.Call([]reflect.Value{in})
			return out[0], out[1].Bool()
		}
	} else if fnTyp.NumOut() == 2 && fnTyp.Out(1).ConvertibleTo(errType) {
		mapFn = func(in reflect.Value) (reflect.Value, bool) {
			out := fnVal.Call([]reflect.Value{in})
			err := out[1].Interface()
			if err != nil && err.(error) != nil {
				ctx.SetErr(err.(error))
			}
			return out[0], err == nil || err.(error) == nil
		}
	} else if fnTyp.NumOut() == 1 {
	} else {
		panic("Map function must be func(element_type) another_type or func(element_type) (another_type,error/bool), now " + fnTyp.String())
	}

	return newStream(ctx, fnTyp.Out(0), q.iter, func(next iterator) iterator {
		return func() (reflect.Value, bool) {
			for {
				if val, ok := next(); !ok {
					return reflect.Value{}, false
				} else if val, ok = mapFn(val); ok {
					return val, true
				} else if ctx.Err() != nil {
					return val, false
				}
			}
		}
	})
}
