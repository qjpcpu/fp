package fp

import "reflect"

var (
	boolType = reflect.TypeOf(true)
	errType  = reflect.TypeOf((*error)(nil)).Elem()
)

func NoError() func(error) bool {
	return func(e error) bool {
		return e == nil
	}
}

func EqualStr(s string) func(string) bool {
	return func(v string) bool {
		return s == v
	}
}

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
