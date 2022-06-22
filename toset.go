package fp

import (
	"fmt"
	"reflect"
)

func (q *stream) ToSetBy(fn interface{}) KVStream {
	fntyp := reflect.TypeOf(fn)
	fnval := reflect.ValueOf(fn)

	var keyTyp, valTyp reflect.Type
	var getKV func(reflect.Value) (reflect.Value, reflect.Value, error)
	switch fntyp.NumOut() {
	case 1:
		keyTyp, valTyp = fntyp.Out(0), q.expectElemTyp
		getKV = func(elem reflect.Value) (reflect.Value, reflect.Value, error) {
			return fnval.Call([]reflect.Value{elem})[0], elem, nil
		}
	case 2:
		keyTyp, valTyp = fntyp.Out(0), fntyp.Out(1)
		getKV = func(elem reflect.Value) (reflect.Value, reflect.Value, error) {
			res := fnval.Call([]reflect.Value{elem})
			return res[0], res[1], nil
		}
	case 3:
		keyTyp, valTyp = fntyp.Out(0), fntyp.Out(1)
		getKV = func(elem reflect.Value) (reflect.Value, reflect.Value, error) {
			var err error
			res := fnval.Call([]reflect.Value{elem})
			if er := res[2].Interface(); er != nil {
				err = er.(error)
			}
			return res[0], res[1], err
		}
	default:
		panic(fmt.Errorf(`bad ToSetBy function %v`, fntyp))
	}

	iter := q.iter
	ctx := newCtx(q.ctx)
	return newKvStream(ctx, keyTyp, valTyp, func() reflect.Value {
		table := reflect.MakeMap(reflect.MapOf(keyTyp, valTyp))
		for {
			val, ok := iter()
			if !ok {
				break
			}
			k, v, err := getKV(val)
			if err != nil {
				ctx.SetErr(err)
				break
			} else {
				table.SetMapIndex(k, v)
			}
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
