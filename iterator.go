package fp

import "reflect"

type iterator func() (reflect.Value, bool)

type middleware func(iterator) iterator
