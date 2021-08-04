package fp

import (
	"reflect"
	"strings"
)

var (
	boolType   = reflect.TypeOf(true)
	errType    = reflect.TypeOf((*error)(nil)).Elem()
	streamType = reflect.TypeOf((*Stream)(nil)).Elem()
)

func NoError() func(error) bool {
	return func(e error) bool {
		return e == nil
	}
}

/* partial functions for filter */

// Equal return a function like func(same_type_of_v) bool, equal by reflect.DeepEqual
func Equal(v interface{}) interface{} {
	typ := reflect.TypeOf(v)
	return reflect.MakeFunc(reflect.FuncOf([]reflect.Type{typ}, []reflect.Type{boolType}, false), func(in []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(reflect.DeepEqual(in[0].Interface(), v))}
	}).Interface()
}

// EqualIgnoreCase compare string ignore case
func EqualIgnoreCase(v string) func(string) bool {
	return func(in string) bool {
		return strings.EqualFold(v, in)
	}
}

func EmptyString() func(string) bool {
	return func(s string) bool { return strings.TrimSpace(s) == "" }
}

/* functions for reduce */

/* stupid golang */
func ShorterString(a, b string) string {
	if len(a) > len(b) {
		return b
	}
	return a
}

func LongerString(a, b string) string {
	if len(a) > len(b) {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func MaxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
func MaxUint16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}
func MaxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}
func MaxInt8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}
func MaxInt16(a, b int16) int16 {
	if a > b {
		return a
	}
	return b
}
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MinInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func MinUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MinUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
func MinUint16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
func MinUint8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}
func MinInt8(a, b int8) int8 {
	if a < b {
		return a
	}
	return b
}
func MinInt16(a, b int16) int16 {
	if a < b {
		return a
	}
	return b
}
