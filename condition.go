package fp

import "reflect"

/* when function should be func(type) bool
 * then/else function should be func(type) (any1,any2...)
 */
type IWhen interface {
	When(interface{}) IThen
}

type IThen interface {
	Then(interface{}) INext
}

type INext interface {
	IWhen
	Else(interface{}) interface{}
}

type ifBuilder struct {
	ifList   []interface{}
	thenList []interface{}
}

func When(fn interface{}) IThen {
	return ifBuilder{
		ifList: []interface{}{fn},
	}
}

func (b ifBuilder) Then(fn interface{}) INext {
	return ifBuilder{
		ifList:   b.ifList,
		thenList: append(b.thenList, fn),
	}
}

func (b ifBuilder) When(fn interface{}) IThen {
	return ifBuilder{
		ifList:   append(b.ifList, fn),
		thenList: b.thenList,
	}
}

func (b ifBuilder) Else(fn interface{}) interface{} {
	conditionList := resolveConditionFuncList(b.ifList...)
	return reflect.MakeFunc(reflect.TypeOf(fn), func(in []reflect.Value) []reflect.Value {
		for i, fn := range conditionList() {
			if reflect.ValueOf(fn).Call(in)[0].Bool() {
				return reflect.ValueOf(b.thenList[i]).Call(in)
			}
		}
		return reflect.ValueOf(fn).Call(in)
	}).Interface()
}

type Condition interface {
	And(fns ...interface{}) Condition
	Or(fns ...interface{}) Condition
	Out() interface{}
	To(ptr interface{})
}

func And(fn1, fn2 interface{}, other ...interface{}) Condition {
	c := condition{fn: reflect.ValueOf(resolveConditionFunc(fn1))}
	list := append([]interface{}{fn2}, other...)
	return c.And(list...)
}

func Or(fn1, fn2 interface{}, other ...interface{}) Condition {
	c := condition{fn: reflect.ValueOf(resolveConditionFunc(fn1))}
	list := append([]interface{}{fn2}, other...)
	return c.Or(list...)
}

func Not(fn interface{}) Condition {
	fn = resolveConditionFunc(fn)
	return condition{
		fn: reflect.MakeFunc(reflect.TypeOf(fn), func(in []reflect.Value) []reflect.Value {
			return []reflect.Value{reflect.ValueOf(!reflect.ValueOf(fn).Call(in)[0].Bool())}
		}),
	}
}

func resolveConditionFunc(f interface{}) interface{} {
	for {
		if inner, ok := f.(Condition); ok {
			f = inner.Out()
		} else {
			return f
		}
	}
}

func resolveConditionFuncList(list ...interface{}) func() []interface{} {
	var realList []interface{}
	var resolved bool
	return func() []interface{} {
		if resolved {
			return realList
		}
		StreamOf(list).Map(resolveConditionFunc).ToSlice(&realList)
		resolved = true
		return realList
	}
}

type condition struct {
	fn reflect.Value
}

func (c condition) Out() interface{}   { return c.fn.Interface() }
func (c condition) To(ptr interface{}) { reflect.ValueOf(ptr).Elem().Set(c.fn) }

func (c condition) And(fns ...interface{}) Condition {
	list := append([]interface{}{c.fn.Interface()}, fns...)
	getList := resolveConditionFuncList(list...)
	return condition{
		fn: reflect.MakeFunc(c.fn.Type(), func(in []reflect.Value) []reflect.Value {
			for _, fn := range getList() {
				if out := reflect.ValueOf(fn).Call(in); !out[0].Bool() {
					return out
				}
			}
			return []reflect.Value{reflect.ValueOf(true)}
		}),
	}
}

func (c condition) Or(fns ...interface{}) Condition {
	list := append([]interface{}{c.fn.Interface()}, fns...)
	getList := resolveConditionFuncList(list...)
	return condition{
		fn: reflect.MakeFunc(c.fn.Type(), func(in []reflect.Value) []reflect.Value {
			for _, fn := range getList() {
				if out := reflect.ValueOf(fn).Call(in); out[0].Bool() {
					return out
				}
			}
			return []reflect.Value{reflect.ValueOf(false)}
		}),
	}
}
