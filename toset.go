package fp

import "reflect"

func (q *stream) ToSetBy(fn interface{}) KVStream {
	fntyp := reflect.TypeOf(fn)
	fnval := reflect.ValueOf(fn)

	keyTyp, valTyp := fntyp.Out(0), q.expectElemTyp
	getKV := func(elem reflect.Value) (reflect.Value, reflect.Value) {
		return fnval.Call([]reflect.Value{elem})[0], elem
	}
	if fntyp.NumOut() == 2 {
		keyTyp, valTyp = fntyp.Out(0), fntyp.Out(1)
		getKV = func(elem reflect.Value) (reflect.Value, reflect.Value) {
			res := fnval.Call([]reflect.Value{elem})
			return res[0], res[1]
		}
	}
	iter := q.iter
	return newKvStream(newCtx(q.ctx), keyTyp, valTyp, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(keyTyp, valTyp))
		for {
			val, ok := iter()
			if !ok {
				break
			}
			table.SetMapIndex(getKV(val))
		}
		return table
	})
}

func (q *stream) ToSet() KVStream {
	iter := q.iter
	return newKvStream(newCtx(q.ctx), q.expectElemTyp, boolType, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(q.expectElemTyp, boolType))
		_true := reflect.ValueOf(true)
		for {
			val, ok := iter()
			if !ok {
				break
			}
			table.SetMapIndex(val, _true)
		}
		return table
	})
}
