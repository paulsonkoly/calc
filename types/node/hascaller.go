package node

// HasCaller determines whether a node contains a call or not.
type HasCaller interface {
	HasCall() bool
}

func (c Call) HasCall() bool     { return true }
func (f Function) HasCall() bool { return false }
func (i Int) HasCall() bool      { return false }
func (f Float) HasCall() bool    { return false }
func (s String) HasCall() bool   { return false }
func (b Bool) HasCall() bool     { return false }
func (l List) HasCall() bool {
	for _, t := range l.Elems {
		if t.HasCall() {
			return true
		}
	}
	return false
}
func (b BinOp) HasCall() bool       { return b.Left.HasCall() || b.Right.HasCall() }
func (a Assign) HasCall() bool      { return a.Value.HasCall() }
func (u UnOp) HasCall() bool        { return u.Target.HasCall() }
func (u IndexAt) HasCall() bool     { return u.Ary.HasCall() || u.At.HasCall() }
func (u IndexFromTo) HasCall() bool { return u.Ary.HasCall() || u.From.HasCall() || u.To.HasCall() }
func (i If) HasCall() bool          { return i.Condition.HasCall() || i.TrueCase.HasCall() }
func (i IfElse) HasCall() bool {
	return i.Condition.HasCall() || i.TrueCase.HasCall() || i.FalseCase.HasCall()
}
func (w While) HasCall() bool   { return w.Condition.HasCall() || w.Body.HasCall() }
func (f For) HasCall() bool     { return f.Iterator.HasCall() || f.Body.HasCall() }
func (r Return) HasCall() bool  { return r.Target.HasCall() }
func (y Yield) HasCall() bool   { return y.Target.HasCall() }
func (r Read) HasCall() bool    { return false }
func (w Write) HasCall() bool   { return false }
func (a Aton) HasCall() bool    { return false }
func (t Toa) HasCall() bool     { return false }
func (n Name) HasCall() bool    { return false }
func (l Local) HasCall() bool   { return false }
func (c Closure) HasCall() bool { return false }
func (b Block) HasCall() bool {
	for _, t := range b.Body {
		if t.HasCall() {
			return true
		}
	}
	return false
}
func (e Exit) HasCall() bool { return false }
