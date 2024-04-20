// package bc provides the data that is passed to calls in the bytecoder.
//
// Data is the actual type containing data. Pass is what the calls receive,
// ensuring with the type system that unwrapping of the data happens. This is
// important because the settings applicable to only one level of recursion can
// be removed here.
package bc

// Data is the data that is passed to the bytecoder.
type Data struct {
	Discard             bool // Discard determines whether computation result can be discarded. Non-transitive
	ForbidTemp          bool // ForbidTemp determines whether tmp register can be used.
	AcceptTemp          bool // AcceptTemp determines whether the result in tmp register is acceptable. Non-transitive
	OpDepth             int  // OpDepth is the depth of arithemtics, logic and relational.
	InFor               bool // InFor determines whether the current node is in a for loop. Transitive
	InFunc              bool // InFunc determines whether the current node is in a function. Transitive
	CtxID, CtxLo, CtxHi int  // CtxID is the current context. CtxLo is the lower bound of the allocated contexts. CtxHi is the upper bound.
}

// Pass is the value a bytecoder receives as argument.
type Pass struct {
	data Data
}

// Data retrieves passed data from a Pass.
func (p Pass) Data() Data {
	return p.data
}

// Pass transforms data for the next level of call, with options applied on the data.
func (d Data) Pass(options ...Option) Pass {
	d.Discard = false
	d.AcceptTemp = false

	for _, o := range options {
		o(&d)
	}

	return Pass{d}
}

// bytecoder options.
type Option func(d *Data)

// WithDiscard sets discard flag on the data.
func WithDiscard(discard bool) Option {
	return func(d *Data) {
		d.Discard = discard
	}
}

// WithForbidTemp sets forbidTemp flag on the data.
func WithForbidTemp(forbidTemp bool) Option {
	return func(d *Data) {
		d.ForbidTemp = forbidTemp
	}
}

// WithDiscard sets discard flag on the data.
func WithAcceptTemp(acceptTemp bool) Option {
	return func(d *Data) {
		d.AcceptTemp = acceptTemp
	}
}

// WithOpDepth sets opDepth on the data.
func WithOpDepth(opDepth int) Option {
	return func(d *Data) {
		d.OpDepth = opDepth
	}
}

// WithInFor sets inFor flag on the data.
func WithInFor(inFor bool) Option {
	return func(d *Data) {
		d.InFor = inFor
	}
}

// WithInFunc sets inFor flag on the data.
func WithInFunc(inFunc bool) Option {
	return func(d *Data) {
		d.InFunc = inFunc
	}
}

// WithCtxID sets ctxID on the data.
func WithCtxID(ctxID int) Option {
	return func(d *Data) {
		d.CtxID = ctxID
	}
}

// WithCtxLo sets ctxLo on the data.
func WithCtxLo(ctxLo int) Option {
	return func(d *Data) {
		d.CtxLo = ctxLo
	}
}

// WithCtxHi sets ctxHi on the data.
func WithCtxHi(ctxHi int) Option {
	return func(d *Data) {
		d.CtxHi = ctxHi
	}
}
