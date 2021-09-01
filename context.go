package fp

type context interface {
	SetErr(err error)
	Err() error
}
type _context struct {
	parent context
	err    error
}

func (ctx *_context) SetErr(err error) {
	ctx.err = err
}

func (ctx *_context) Err() error {
	if err := ctx.err; err != nil {
		return err
	}
	if ctx.parent != nil {
		return ctx.parent.Err()
	}
	return nil
}

func newCtx(parent context) context {
	if parent == nil {
		parent = &_context{}
	}
	return &_context{parent: parent}
}
