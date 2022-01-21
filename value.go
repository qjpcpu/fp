package fp

import "reflect"

/* value related */
type Value struct {
	typ reflect.Type
	val reflect.Value
}

func (rv Value) To(dst interface{}) bool {
	if !rv.val.IsValid() {
		return false
	}
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(rv.val)
	return true
}

func (rv Value) Result() interface{} {
	if !rv.val.IsValid() {
		return nil
	}
	return rv.val.Interface()
}

func (rv Value) Err() error {
	if !rv.val.IsValid() {
		return nil
	}
	res := rv.val.Interface()
	if res == nil {
		return nil
	}
	return res.(error)
}

func (rv Value) Strings() (s []string) {
	rv.To(&s)
	return
}

func (rv Value) Bytes() (s []byte) {
	rv.To(&s)
	return
}

func (rv Value) Ints() (s []int) {
	rv.To(&s)
	return
}

func (rv Value) Int64s() (s []int64) {
	rv.To(&s)
	return
}

func (rv Value) Int32s() (s []int32) {
	rv.To(&s)
	return
}

func (rv Value) Uints() (s []uint) {
	rv.To(&s)
	return
}

func (rv Value) Uint32s() (s []uint32) {
	rv.To(&s)
	return
}

func (rv Value) Uint64s() (s []uint64) {
	rv.To(&s)
	return
}

func (rv Value) StringsList() (s [][]string) {
	rv.To(&s)
	return
}

func (rv Value) String() (s string) {
	rv.To(&s)
	return
}

func (rv Value) Int() (s int) {
	rv.To(&s)
	return
}

func (rv Value) Int64() (s int64) {
	rv.To(&s)
	return
}

func (rv Value) Int32() (s int32) {
	rv.To(&s)
	return
}

func (rv Value) Uint() (s uint) {
	rv.To(&s)
	return
}

func (rv Value) Uint32() (s uint32) {
	rv.To(&s)
	return
}

func (rv Value) Uint64() (s uint64) {
	rv.To(&s)
	return
}

func (rv Value) Float64() (s float64) {
	rv.To(&s)
	return
}
