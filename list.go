package fp

import (
	"reflect"
	"sync/atomic"
)

/* core definitions */
type list struct {
	elem func() *atom
	next func() *list
}

type atom struct {
	typ reflect.Type
	val reflect.Value
}

/* core operations */
func car(c *list) *atom {
	if c == nil {
		return nil
	}
	return c.elem()
}

func cdr(c *list) *list {
	if c == nil {
		return nil
	}
	return c.next()
}

/* lazy cons */
// the atom func should not be nil, otherwise the list would be broken
func cons(atom func() *atom, listc func() *list) *list {
	return &list{
		elem: atom,
		next: listc,
	}
}

/* utils functions */
func isNil(l *list) bool {
	return l == nil
}

func createAtom(typ reflect.Type, val reflect.Value) *atom {
	return &atom{
		typ: typ,
		val: val,
	}
}

func emptyList() *list {
	return &list{
		elem: func() *atom { return nil },
		next: func() *list { return nil },
	}
}

func asSlice(elemTyp reflect.Type, l *list) reflect.Value {
	typ := reflect.SliceOf(elemTyp)
	slice := reflect.Zero(typ)
	processList(l, func(cell *atom) bool {
		slice = reflect.Append(slice, cell.val)
		return true
	})
	return slice
}

func processList(l *list, fn func(*atom) (continueIter bool)) {
	if isNil(l) {
		return
	}
	if fn == nil {
		fn = func(*atom) bool { return true }
	}
	for cell := car(l); cell != nil; {
		if !fn(cell) {
			break
		}
		l = cdr(l)
		cell = car(l)
	}
}

func carOnce(fn func() *atom) func() *atom {
	var flag int32
	var cache *atom
	return func() *atom {
		if atomic.CompareAndSwapInt32(&flag, 0, 1) {
			cache = fn()
		}
		return cache
	}
}

func cdrOnce(fn func() *list) func() *list {
	var flag int32
	var cache *list
	return func() *list {
		if atomic.CompareAndSwapInt32(&flag, 0, 1) {
			cache = fn()
		}
		return cache
	}
}

/* high order functions */
func mapcar(fn reflect.Value, list1 *list) *list {
	if isNil(list1) {
		return list1
	}
	return cons(
		func() *atom {
			elem := car(list1)
			if elem == nil {
				return nil
			}
			return createAtom(fn.Type().Out(0), fn.Call([]reflect.Value{elem.val})[0])
		},
		func() *list {
			return mapcar(fn, cdr(list1))
		},
	)
}

func batchcar(size int, list1 *list) *list {
	if isNil(list1) {
		return list1
	}

	carfn := carOnce(func() *atom {
		var firstPartition *atom
		for i := 0; i < size; i++ {
			elem := car(list1)
			if elem == nil {
				break
			}
			if firstPartition == nil {
				firstPartition = &atom{
					typ: reflect.SliceOf(elem.typ),
					val: reflect.Zero(reflect.SliceOf(elem.typ)),
				}
			}
			firstPartition.val = reflect.Append(firstPartition.val, elem.val)
			list1 = cdr(list1)
		}
		return firstPartition
	})
	cdrfn := cdrOnce(func() *list {
		carfn()
		return batchcar(size, list1)
	})
	return cons(carfn, cdrfn)
}

func takecar(size int, list1 *list) *list {
	if isNil(list1) {
		return list1
	}

	carfn := carOnce(func() *atom {
		if size <= 0 {
			return nil
		}
		return car(list1)
	})
	cdrfn := cdrOnce(func() *list {
		if size <= 0 {
			return nil
		}
		carfn()
		return takecar(size-1, cdr(list1))
	})
	return cons(carfn, cdrfn)
}

func takeWhile(fn reflect.Value, list1 *list) *list {
	if isNil(list1) {
		return list1
	}

	carfn := carOnce(func() *atom {
		elem := car(list1)
		if elem == nil {
			return nil
		}
		if !fn.Call([]reflect.Value{elem.val})[0].Bool() {
			return nil
		}
		return elem
	})
	cdrfn := cdrOnce(func() *list {
		if carfn() == nil {
			return nil
		}
		return takeWhile(fn, cdr(list1))
	})
	return cons(carfn, cdrfn)
}

func skipcar(size int, list1 *list) *list {
	if isNil(list1) {
		return list1
	}
	carfn := carOnce(func() *atom {
		if size <= 0 {
			return car(list1)
		} else {
			for i := 0; i < size; i++ {
				drop := car(list1)
				if drop == nil {
					break
				}
				list1 = cdr(list1)
			}
			return car(list1)
		}
	})
	cdrfn := cdrOnce(func() *list {
		carfn()
		return cdr(list1)
	})
	return cons(carfn, cdrfn)
}

func skipWhile(fn reflect.Value, list1 *list) *list {
	if isNil(list1) {
		return list1
	}

	var left *list
	carfn := carOnce(func() *atom {
		var elem *atom
		left = list1
		for elem = car(list1); elem != nil && fn.Call([]reflect.Value{elem.val})[0].Bool(); {
			left = cdr(left)
			elem = car(left)
		}
		return elem
	})
	cdrfn := cdrOnce(func() *list {
		carfn()
		return cdr(left)
	})
	return cons(carfn, cdrfn)
}

func selectcar(fn reflect.Value, list1 *list) *list {
	if isNil(list1) {
		return list1
	}
	return cons(
		func() *atom {
			var elem *atom
			for elem = car(list1); elem != nil && !fn.Call([]reflect.Value{elem.val})[0].Interface().(bool); {
				list1 = cdr(list1)
				elem = car(list1)
			}
			return elem
		},
		func() *list {
			return selectcar(fn, cdr(list1))
		},
	)
}

func foreachcar(fn reflect.Value, list1 *list) *list {
	if isNil(list1) {
		return list1
	}
	return cons(
		func() *atom {
			var elem *atom
			if elem = car(list1); elem != nil {
				fn.Call([]reflect.Value{elem.val})
			}
			return elem
		},
		func() *list {
			return foreachcar(fn, cdr(list1))
		},
	)
}

func concat(list1, list2 *list) *list {
	if isNil(list1) {
		return list2
	}
	if isNil(list2) {
		return list1
	}
	carfn := func() *atom {
		if list1 != nil {
			if elem := car(list1); elem != nil {
				return elem
			}
		}
		return car(list2)
	}
	cdrfn := func() *list {
		if list1 != nil {
			if v := car(list1); v != nil {
				return concat(cdr(list1), list2)
			}
		}
		return cdr(list2)
	}
	return cons(carfn, cdrfn)
}

func flatten(list1 *list) *list {
	return flattencar(nil, list1)
}

func flattencar(carlist *list, cdrlist *list) *list {
	makeListOnce := cdrOnce(func() *list {
		return _flattenCdrCar(cdrlist)
	})
	carfn := func() *atom {
		if carlist != nil {
			if elem := car(carlist); elem != nil {
				return elem
			}
		}
		if cdrlist != nil {
			return car(makeListOnce())
		}
		return nil
	}
	cdrfn := func() *list {
		if carlist != nil {
			if v := car(carlist); v != nil {
				return flattencar(cdr(carlist), cdrlist)
			}
		}
		if cdrlist != nil {
			return flattencar(cdr(makeListOnce()), cdr(cdrlist))
		}
		return nil
	}
	return cons(carfn, cdrfn)
}

func _flattenCdrCar(cdrlist *list) *list {
	elem := car(cdrlist)
	if elem == nil {
		return nil
	}
	subl := makeList(elem.typ, elem.val)
	/* oops, this is a empty list */
	if car(subl) == nil {
		return _flattenCdrCar(cdr(cdrlist))
	}
	return subl
}
