package fp

type nilMonad struct{ err error }

func newNilMonad(err error) nilMonad        { return nilMonad{err: err} }
func (m nilMonad) Map(fn interface{}) Monad { return m }

func (m nilMonad) ExpectPass(fn interface{}) Monad { return m }

func (m nilMonad) ExpectNoError(fn interface{}) Monad { return m }

func (m nilMonad) StreamOf(fn interface{}) Stream { return newNilStream() }

func (m nilMonad) Zip(interface{}, ...Monad) Monad { return m }

func (m nilMonad) Once() Monad { return m }

func (m nilMonad) Val() Value { return Value{err: m.err} }

func (m nilMonad) fnContainer() func() (interface{}, bool, error) {
	return func() (interface{}, bool, error) { return nil, false, m.err }
}
