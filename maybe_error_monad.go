package fp

import (
	"reflect"
	"sync"
)

type Monad interface {
	// Map func(type1) (type2,&optional error/bool)
	Map(fn interface{}) Monad
	// ExpectPass func(type1) (bool)
	ExpectPass(fn interface{}) Monad
	// ExpectNoError func(type1) (error)
	ExpectNoError(fn interface{}) Monad
	// StreamOf func(type1) (type2,&optional error/bool)
	StreamOf(fn interface{}) Stream
	// Zip monads
	Zip(interface{}, ...Monad) Monad
	// Once monad
	Once() Monad
	// Value of monad
	Val() Value
	// fnContainer return error_boolean_monad real container
	fnContainer() func() (interface{}, bool, error)
}

func M(v ...interface{}) Monad {
	if isNilObject(v[0]) {
		if len(v) > 1 {
			if _, ok := v[len(v)-1].(error); ok {
				return newNilMonad(v[len(v)-1].(error))
			}
		}
		return newNilMonad(nil)
	}
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
	/* fn is kind of func() (any,bool,error) */
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

func (em errorMonad) fnContainer() func() (interface{}, bool, error) {
	return func() (interface{}, bool, error) {
		out := em.fn.Call(nil)
		if e := out[2].Interface(); e != nil && e.(error) != nil {
			return nil, false, e.(error)
		}
		if !out[1].Bool() {
			return nil, false, nil
		}
		return out[0].Interface(), true, nil
	}
}

func (em errorMonad) Zip(fn interface{}, others ...Monad) Monad {
	fnVal := toErrMonadFunc(fn)
	outTyp := reflect.FuncOf(nil, outTypes(fnVal.Type()), false)
	return newErrorMonad(reflect.MakeFunc(outTyp, func(in []reflect.Value) []reflect.Value {
		monads := append([]Monad{em}, others...)
		var input []reflect.Value
		for _, m := range monads {
			out, ok, err := m.fnContainer()()
			if err != nil {
				return []reflect.Value{reflect.Zero(fnVal.Type().Out(0)), reflect.ValueOf(false), reflect.ValueOf(err)}
			}
			if !ok {
				return []reflect.Value{reflect.Zero(fnVal.Type().Out(0)), reflect.ValueOf(false), reflect.Zero(errType)}
			}
			input = append(input, reflect.ValueOf(out))
		}
		return fnVal.Call(input)
	}))
}

func (em errorMonad) Once() Monad {
	var out []reflect.Value
	var once sync.Once
	return newErrorMonad(reflect.MakeFunc(em.fn.Type(), func(in []reflect.Value) []reflect.Value {
		once.Do(func() {
			out = em.fn.Call(nil)
		})
		return out
	}))
}

func (em errorMonad) ExpectPass(fn interface{}) Monad {
	return em.Expect(fn)
}

func (em errorMonad) ExpectNoError(fn interface{}) Monad {
	return em.Expect(fn)
}

func (em errorMonad) Expect(fn interface{}) Monad {
	typ := reflect.TypeOf(fn)
	wrapTyp := reflect.FuncOf(nil, outTypes(em.fn.Type()), false)
	if typ.NumOut() == 1 && typ.Out(0).AssignableTo(errType) {
		return newErrorMonad(reflect.MakeFunc(wrapTyp, func(in []reflect.Value) []reflect.Value {
			out := em.fn.Call(in)
			if e := out[2].Interface(); e != nil && e.(error) != nil {
				return []reflect.Value{reflect.Zero(em.fn.Type().Out(0)), reflect.ValueOf(false), out[2]}
			}
			if !out[1].Bool() {
				return []reflect.Value{reflect.Zero(em.fn.Type().Out(0)), reflect.ValueOf(false), reflect.Zero(errType)}
			}
			out1 := reflect.ValueOf(fn).Call(out[:1])
			return []reflect.Value{out[0], reflect.ValueOf(true), out1[0]}
		}))
	} else if typ.NumOut() == 1 && typ.Out(0).AssignableTo(boolType) {
		return newErrorMonad(reflect.MakeFunc(wrapTyp, func(in []reflect.Value) []reflect.Value {
			out := em.fn.Call(in)
			if e := out[2].Interface(); e != nil && e.(error) != nil {
				return []reflect.Value{reflect.Zero(em.fn.Type().Out(0)), reflect.ValueOf(false), out[2]}
			}
			if !out[1].Bool() {
				return []reflect.Value{reflect.Zero(em.fn.Type().Out(0)), reflect.ValueOf(false), reflect.Zero(errType)}
			}
			out1 := reflect.ValueOf(fn).Call(out[:1])
			return []reflect.Value{out[0], out1[0], reflect.Zero(errType)}
		}))
	} else {
		panic("bad expect function " + typ.String())
	}
}

func (em errorMonad) StreamOf(fn interface{}) Stream {
	fnVal := toErrMonadFunc(fn)
	ctx := newCtx(nil)
	evalM := func() (reflect.Value, bool, error) {
		out := em.fn.Call(nil)
		if e := out[2].Interface(); e != nil && e.(error) != nil {
			return reflect.Value{}, false, e.(error)
		}
		if !out[1].Bool() {
			return reflect.Value{}, false, nil
		}
		out = fnVal.Call(out[:1])
		if e := out[2].Interface(); e != nil && e.(error) != nil {
			return reflect.Value{}, false, e.(error)
		}
		if !out[1].Bool() {
			return reflect.Value{}, false, nil
		}
		return out[0], true, nil
	}

	if typ := fnVal.Type().Out(0); typ != streamType {
	} else if out, ok, err := evalM(); err != nil {
		sc := newNilSource()
		ctx.SetErr(err)
		return newStream(ctx, sc.ElemType(), sc.Next)
	} else if !ok {
		return newNilStream()
	} else {
		return out.Interface().(Stream)
	}
	elemTyp := fnVal.Type().Out(0).Elem()
	var iter iterator
	return newStream(ctx, elemTyp, func() (reflect.Value, bool) {
		if iter == nil {
			out, ok, err := evalM()
			if err != nil {
				ctx.SetErr(err)
				return reflect.Value{}, false
			} else if !ok {
				return reflect.Value{}, false
			}
			_, iter = makeIter(ctx, out)
		}
		return iter()
	})
}

func (em errorMonad) Val() Value {
	out := em.fn.Call(nil)
	if e := out[2].Interface(); e != nil && e.(error) != nil {
		return Value{err: e.(error)}
	}
	if out[1].Bool() {
		return Value{typ: out[0].Type(), val: out[0]}
	}
	/* out[0] would always valid */
	return Value{typ: out[0].Type()}
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

func isNilObject(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	}
	return false
}
