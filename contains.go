package fp

import (
	"fmt"
	"reflect"
)

func (q *stream) Contains(e interface{}) (yes bool) {
	var eq func(reflect.Value) bool
	val := reflect.ValueOf(e)
	switch reflect.TypeOf(e).Kind() {
	case reflect.String:
		t := val.String()
		eq = func(v reflect.Value) bool { return v.String() == t }
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t := val.Int()
		eq = func(v reflect.Value) bool { return v.Int() == t }
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		t := val.Uint()
		eq = func(v reflect.Value) bool { return v.Uint() == t }
	case reflect.Bool:
		t := val.Bool()
		eq = func(v reflect.Value) bool { return v.Bool() == t }
	case reflect.Float32, reflect.Float64:
		t := val.Float()
		eq = func(v reflect.Value) bool { return v.Float() == t }
	default:
		eq = func(v reflect.Value) bool { return reflect.DeepEqual(v.Interface(), e) }
	}

	q.iter = repeatableIter(q.iter, func(v reflect.Value) bool {
		yes = eq(v)
		return !yes
	})
	return
}

func (q *stream) ContainsBy(eqfn interface{}) (yes bool) {
	fnval := reflect.ValueOf(eqfn)
	q.iter = repeatableIter(q.iter, func(v reflect.Value) bool {
		yes = fnval.Call([]reflect.Value{v})[0].Bool()
		return !yes
	})
	return
}

func repeatableIter(iter iterator, f func(reflect.Value) bool) iterator {
	if iter == nil {
		return nil
	}
	var vals []reflect.Value
	for {
		val, ok := iter()
		if !ok {
			break
		}
		vals = append(vals, val)
		if !f(val) {
			break
		}
	}
	var i int
	return func() (reflect.Value, bool) {
		for i < len(vals) {
			i++
			return vals[i-1], true
		}
		return iter()
	}
}

func (q *stream) compare(a, b reflect.Value) int {
	switch q.expectElemTyp.Kind() {
	case reflect.String:
		if a.String() < b.String() {
			return -1
		} else if a.String() > b.String() {
			return 1
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() < b.Int() {
			return -1
		} else if a.Int() > b.Int() {
			return 1
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if a.Uint() < b.Uint() {
			return -1
		} else if a.Uint() > b.Uint() {
			return 1
		}
	case reflect.Bool:
		if !a.Bool() && b.Bool() {
			return -1
		} else if a.Bool() && !b.Bool() {
			return 1
		}
	default:
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			s1, s2 := fmt.Sprint(a.Interface()), fmt.Sprint(b.Interface())
			if s1 < s2 {
				return -1
			} else if s1 > s2 {
				return 1
			}
		}
	}
	return 0
}
