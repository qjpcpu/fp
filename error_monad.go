package fp

import (
	"reflect"
)

type Monad interface {
	Map(fn interface{}) Monad
	FlatMap(fn interface{}) Stream
	To(ptr interface{}) error
}

func M(v ...interface{}) Monad {
	typ := reflect.FuncOf(nil, []reflect.Type{reflect.TypeOf(v[0]), boolType, errType}, false)
	if len(v) > 1 {
		if _, ok := v[len(v)-1].(error); ok {
			return newErrorMonad(reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
				return []reflect.Value{reflect.ValueOf(v[0]), reflect.ValueOf(true), reflect.ValueOf(v[len(v)-1])}
			}))
		}
		if _, ok := v[len(v)-1].(bool); ok {
			return newErrorMonad(reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
				return []reflect.Value{reflect.ValueOf(v[0]), reflect.ValueOf(v[len(v)-1]), reflect.Zero(errType)}
			}))
		}
	}
	return newErrorMonad(reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(v[0]), reflect.ValueOf(true), reflect.Zero(errType)}
	}))
}

type errorMonad struct {
	fn reflect.Value
}

func newErrorMonad(fn reflect.Value) errorMonad {
	return errorMonad{fn: fn}
}

func (em errorMonad) Map(fn interface{}) Monad {
	fnVal := toErrMonadFunc(fn)
	outTyp := reflect.FuncOf(nil, outTypes(fnVal.Type()), false)
	return newErrorMonad(reflect.MakeFunc(outTyp, func(in []reflect.Value) []reflect.Value {
		out := em.fn.Call(in)
		if e := out[2].Interface(); e != nil && e.(error) != nil {
			return []reflect.Value{reflect.Zero(fnVal.Type().Out(0)), reflect.ValueOf(false), out[2]}
		}
		if !out[1].Bool() {
			return []reflect.Value{reflect.Zero(fnVal.Type().Out(0)), reflect.ValueOf(false), reflect.Zero(errType)}
		}
		return fnVal.Call(out[:1])
	}))
}

func (em errorMonad) FlatMap(fn interface{}) Stream {
	fnVal := toErrMonadFunc(fn)
	ctx := newCtx(nil)
	elemTyp := fnVal.Type().Out(0).Elem()
	var iter iterator
	return newStream(ctx, elemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			out := em.fn.Call(nil)
			if e := out[2].Interface(); e != nil && e.(error) != nil {
				ctx.SetErr(e.(error))
				return reflect.Value{}, false
			}
			if !out[1].Bool() {
				return reflect.Value{}, false
			}
			out = fnVal.Call(out[:1])
			if e := out[2].Interface(); e != nil && e.(error) != nil {
				ctx.SetErr(e.(error))
				return reflect.Value{}, false
			}
			if !out[1].Bool() {
				return reflect.Value{}, false
			}
			_, iter = makeIter(out[0])
		}
		return iter()
	})
}

func (em errorMonad) To(ptr interface{}) error {
	out := em.fn.Call(nil)
	if e := out[2].Interface(); e != nil && e.(error) != nil {
		return e.(error)
	}
	if out[1].Bool() {
		reflect.ValueOf(ptr).Elem().Set(out[0])
	}
	return nil
}

func toErrMonadFunc(fn interface{}) reflect.Value {
	typ := reflect.TypeOf(fn)
	ntyp := reflect.FuncOf(inTypes(typ), []reflect.Type{typ.Out(0), boolType, errType}, false)
	if typ.NumOut() == 2 && typ.Out(1).AssignableTo(errType) {
		return reflect.MakeFunc(ntyp, func(in []reflect.Value) []reflect.Value {
			out := reflect.ValueOf(fn).Call(in)
			return []reflect.Value{out[0], reflect.ValueOf(true), out[1]}
		})
	} else if typ.NumOut() == 2 && typ.Out(1).AssignableTo(boolType) {
		return reflect.MakeFunc(ntyp, func(in []reflect.Value) []reflect.Value {
			out := reflect.ValueOf(fn).Call(in)
			return []reflect.Value{out[0], out[1], reflect.Zero(errType)}
		})
	}
	return reflect.MakeFunc(ntyp, func(in []reflect.Value) []reflect.Value {
		out := reflect.ValueOf(fn).Call(in)
		return []reflect.Value{out[0], reflect.ValueOf(true), reflect.Zero(errType)}
	})
}
