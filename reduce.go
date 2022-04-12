package fp

import "reflect"

func (q *stream) Reduce(initval interface{}, fn interface{}) Value {
	typ := reflect.TypeOf(initval)
	memo := reflect.ValueOf(initval)
	fnval := reflect.ValueOf(fn)
	for {
		val, ok := q.iter()
		if !ok {
			break
		}
		memo = fnval.Call([]reflect.Value{memo, val})[0]
	}
	return Value{typ: typ, val: memo, err: q.Error()}
}

func (q *stream) Reduce0(fn interface{}) Value {
	initVal, ok := q.iter()
	if !ok {
		return Value{typ: q.expectElemTyp, val: reflect.Zero(q.expectElemTyp), err: q.ctx.Err()}
	}
	return q.Reduce(initVal.Interface(), fn)
}
