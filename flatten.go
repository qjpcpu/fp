package fp

import (
	"reflect"
	"sync/atomic"
)

func (q *stream) Flatten() Stream {
	if kind := q.expectElemTyp.Kind(); kind != reflect.Chan && kind != reflect.Slice && kind != reflect.Array && q.expectElemTyp != streamType {
		panic(q.expectElemTyp.String() + " can not be flatten")
	}

	var elemType reflect.Type
	_makeIter := makeIter
	if q.expectElemTyp == streamType {
		iter := q.iter
		for {
			v, ok := iter()
			if !ok {
				/* no inner stream found and cause we couldn't guess inner type */
				/* just return nil stream */
				return newNilStream()
			}
			/* we should find first non-NilStream element */
			if isNilStream(v.Interface().(Stream)) {
				continue
			}
			/* gotcha */
			elemType = v.Interface().(Stream).ToSource().ElemType()
			/* recover q's iter */
			var flag int32
			q.iter = func() (reflect.Value, bool) {
				if atomic.CompareAndSwapInt32(&flag, 0, 1) {
					return v, true
				}
				return iter()
			}
			break
		}
		_makeIter = func(_ context, v reflect.Value) (reflect.Type, iterator) {
			return elemType, v.Interface().(Stream).ToSource().Next
		}
	} else {
		elemType = q.expectElemTyp.Elem()
	}
	ctx2 := newCtx(q.ctx)
	return newStream(ctx2, elemType, q.iter, func(outernext iterator) iterator {
		var innernext iterator
		var inner reflect.Value
		return func() (item reflect.Value, ok bool) {
			for !ok {
				if !inner.IsValid() {
					inner, ok = outernext()
					if !ok {
						return
					}
					/* we should jump over nil stream element */
					if innerS, isStream := inner.Interface().(Stream); isStream && isNilStream(innerS) {
						/* disable inner value */
						inner = reflect.Value{}
						ok = false
						continue
					}
					_, innernext = _makeIter(ctx2, inner)
				}
				item, ok = innernext()
				if !ok {
					inner = reflect.Value{}
				}
			}
			return
		}
	})
}
