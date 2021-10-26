package fp

import (
	"reflect"
	"strings"
)

func (q *stream) ToSlice(dst interface{}) error {
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		panic(`fp: dst must be pointer`)
	}
	val.Elem().Set(q.getValue(val.Elem()))
	return q.ctx.Err()
}

func (q *stream) JoinStrings(seq string) string {
	return strings.Join(q.Strings(), seq)
}

func (q *stream) getResult() Value {
	return Value{
		typ: reflect.SliceOf(q.expectElemTyp),
		val: q.getValue(reflect.Value{}),
	}
}

func (q *stream) Run() {
	q.getValOnce.Do(func() {
		for {
			if _, ok := q.iter(); !ok {
				break
			}
		}
	})
}

func (q *stream) Error() error {
	if q.expectElemTyp.AssignableTo(errType) {
		if err := q.SkipWhile(NoError()).First().Err(); err != nil {
			return err
		}
		return q.ctx.Err()
	}
	q.Run()
	return q.ctx.Err()
}

func (q *stream) getValue(slice reflect.Value) reflect.Value {
	q.getValOnce.Do(func() {
		if !slice.IsValid() {
			slice = reflect.Zero(reflect.SliceOf(q.expectElemTyp))
		} else if slice.Len() > 0 {
			slice = reflect.MakeSlice(reflect.SliceOf(q.expectElemTyp), 0, slice.Len())
		}
		for {
			if val, ok := q.iter(); ok {
				slice = reflect.Append(slice, val)
			} else {
				break
			}
		}
		q.val = slice
	})
	return q.val
}

func (q *stream) Strings() (s []string)       { return q.getResult().Strings() }
func (q *stream) Ints() (s []int)             { return q.getResult().Ints() }
func (q *stream) Int64s() (s []int64)         { return q.getResult().Int64s() }
func (q *stream) Int32s() (s []int32)         { return q.getResult().Int32s() }
func (q *stream) Uints() (s []uint)           { return q.getResult().Uints() }
func (q *stream) Uint32s() (s []uint32)       { return q.getResult().Uint32s() }
func (q *stream) Uint64s() (s []uint64)       { return q.getResult().Uint64s() }
func (q *stream) Bytes() (s []byte)           { return q.getResult().Bytes() }
func (q *stream) Float64s() (s []float64)     { q.getResult().To(&s); return }
func (q *stream) StringsList() (s [][]string) { return q.getResult().StringsList() }
