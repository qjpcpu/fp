package fp

import "reflect"

func (q *stream) Branch(processors ...StreamProcessor) {
	switch len(processors) {
	case 0:
		return
	case 1:
		processors[0](q)
		return
	}

	slice := reflect.MakeSlice(reflect.SliceOf(q.expectElemTyp), 0, 0)
	movingStream := newStream(newCtx(q.ctx), q.expectElemTyp, q.iter, func(next iterator) iterator {
		return func() (val reflect.Value, ok bool) {
			if val, ok = next(); ok {
				slice = reflect.Append(slice, val)
			}
			return
		}
	})

	for _, processor := range processors {
		processor(StreamOf(slice.Interface()).Union(movingStream))
	}
}
