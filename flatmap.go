package fp

func (q *stream) FlatMap(fn interface{}) Stream {
	return q.Map(fn).Flatten()
}
