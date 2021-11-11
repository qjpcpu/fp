package fp

import (
	"reflect"
)

type Cursor interface {
	Next() bool
	Scan(...interface{}) error
}

// StreamByCursor create stream by cursor
// mapfn should looks like func(type1,type2...) (typex,&optional bool/error)
func StreamByCursor(c Cursor, mapfn interface{}) Stream {
	argTypes := inTypes(reflect.TypeOf(mapfn))
	makeArgs := _makeCursorMapArgsWithErr(argTypes)

	failedRes := []reflect.Value{
		reflect.ValueOf(makeArgs(nil)),
		reflect.ValueOf(false),
	}
	/* ftyp is func()([]interface{},bool) */
	ftyp := reflect.FuncOf(nil, []reflect.Type{reflect.TypeOf([]interface{}{}), boolType}, false)
	fn := reflect.MakeFunc(ftyp, func([]reflect.Value) []reflect.Value {
		if c.Next() {
			return []reflect.Value{
				reflect.ValueOf(makeArgs(c.Scan)),
				reflect.ValueOf(true),
			}
		}
		return failedRes
	})
	if nmap, bmap, ok := convertBooleanMap(mapfn); ok {
		return StreamOf(fn.Interface()).
			Map(_wrapCursorMap(nmap)).Map(bmap)
	}
	return StreamOf(fn.Interface()).
		Map(_wrapCursorMap(mapfn))
}

func _makeCursorMapArgsWithErr(argTypes []reflect.Type) func(func(...interface{}) error) []interface{} {
	return func(scan func(...interface{}) error) []interface{} {
		args := make([]interface{}, len(argTypes))
		for i, t := range argTypes {
			if t.Kind() == reflect.Ptr {
				args[i] = reflect.New(t.Elem()).Interface()
			} else {
				args[i] = reflect.New(t).Interface()
			}
		}
		var err error
		if scan != nil {
			err = scan(args...)
		}
		for i, t := range argTypes {
			if t.Kind() != reflect.Ptr {
				args[i] = reflect.ValueOf(args[i]).Elem().Interface()
			}
		}
		return append(args, err)
	}
}

/* convert cursor map function as func([]interface{}) (origtype,error) */
func _wrapCursorMap(mapfn interface{}) interface{} {
	var withExtraRetErr bool
	retTypes := outTypes(reflect.TypeOf(mapfn))
	var zeroRetVals []reflect.Value
	if !retTypes[len(retTypes)-1].AssignableTo(errType) {
		StreamOf(retTypes).Map(reflect.Zero).ToSlice(&zeroRetVals)
		retTypes = append(retTypes, errType)
		withExtraRetErr = true
	} else {
		StreamOf(retTypes[:len(retTypes)-1]).Map(reflect.Zero).ToSlice(&zeroRetVals)
	}
	mapfnVal := reflect.ValueOf(mapfn)
	realMapFnTyp := reflect.FuncOf([]reflect.Type{reflect.TypeOf([]interface{}{})}, retTypes, false)

	return reflect.MakeFunc(realMapFnTyp, func(in []reflect.Value) []reflect.Value {
		realIn := in[0].Interface().([]interface{})
		realInVals := make([]reflect.Value, len(realIn))
		for i, v := range realIn {
			realInVals[i] = reflect.ValueOf(v)
		}
		if err := realIn[len(realIn)-1]; err == nil || err.(error) == nil {
			out := mapfnVal.Call(realInVals[:len(realInVals)-1])
			if withExtraRetErr {
				out = append(out, reflect.Zero(errType))
			}
			return out
		} else {
			return append(zeroRetVals, realInVals[len(realInVals)-1])
		}
	}).Interface()
}

func convertBooleanMap(mapfn interface{}) (newMapFn interface{}, filterMapFn interface{}, ok bool) {
	mapVal := reflect.ValueOf(mapfn)
	retTypes := outTypes(reflect.TypeOf(mapfn))
	if len(retTypes) != 2 || retTypes[1] != boolType {
		return nil, nil, false
	}

	composedType := reflect.StructOf([]reflect.StructField{
		{Name: "Val", Type: retTypes[0]},
		{Name: "OK", Type: boolType},
	})
	newMapFnType := reflect.FuncOf(inTypes(reflect.TypeOf(mapfn)), []reflect.Type{composedType}, false)
	newMapFn = reflect.MakeFunc(newMapFnType, func(in []reflect.Value) []reflect.Value {
		out := mapVal.Call(in)
		val := reflect.New(composedType)
		val.Elem().FieldByName("Val").Set(out[0])
		val.Elem().FieldByName("OK").Set(out[1])
		return []reflect.Value{val.Elem()}
	}).Interface()

	filterFnType := reflect.FuncOf([]reflect.Type{composedType}, []reflect.Type{retTypes[0], retTypes[1]}, false)
	filterMapFn = reflect.MakeFunc(filterFnType, func(in []reflect.Value) []reflect.Value {
		return []reflect.Value{in[0].FieldByName("Val"), in[0].FieldByName("OK")}
	}).Interface()

	ok = true
	return
}
