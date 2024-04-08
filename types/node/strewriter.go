package node

type SymTbl []map[string]int

// STRewriter is a recursive node transformation that resolves local and
// closure variable names to indices into the call frame
//
// It is not idempotent. It is supposed to be called after the parser produced
// the AST and before evaluation.
// Global variables are untouched. Local variables are changed from Name nodes
// to either Local or Closure nodes.
type STRewriter interface {
	STRewrite(symTbl SymTbl) Type
}

func (c Call) STRewrite(symTbl SymTbl) Type {
	return Call{Name: c.Name.STRewrite(symTbl), Arguments: c.Arguments.STRewrite(symTbl).(List)}
}

func (f Function) STRewrite(symTbl SymTbl) Type {
	// new lexical scope
	scope := map[string]int{}

	// assign parameters to scope
	for i, t := range f.Parameters.Elems {
		name := t.(Name)
		scope[string(name)] = i
	}

	// push scope
	symTbl = append(symTbl, scope)

	parameters := f.Parameters.STRewrite(symTbl).(List)
	body := f.Body.STRewrite(symTbl)

	localCnt := len(symTbl[len(symTbl)-1])

	// pop the lexical scope by ignoring slc

	return Function{Parameters: parameters, Body: body, LocalCnt: localCnt}
}

func (i Int) STRewrite(_ SymTbl) Type    { return (i) }
func (f Float) STRewrite(_ SymTbl) Type  { return (f) }
func (s String) STRewrite(_ SymTbl) Type { return (s) }
func (b Bool) STRewrite(_ SymTbl) Type   { return (b) }

func (b BinOp) STRewrite(symTbl SymTbl) Type {
	return BinOp{Op: b.Op, Left: b.Left.STRewrite(symTbl), Right: b.Right.STRewrite(symTbl)}
}

func (u UnOp) STRewrite(symTbl SymTbl) Type {
	return UnOp{Op: u.Op, Target: u.Target.STRewrite(symTbl)}
}

func (i IndexAt) STRewrite(symTbl SymTbl) Type {
	return IndexAt{Ary: i.Ary.STRewrite(symTbl), At: i.At.STRewrite(symTbl)}
}

func (i IndexFromTo) STRewrite(symTbl SymTbl) Type {
	return IndexFromTo{Ary: i.Ary.STRewrite(symTbl), From: i.From.STRewrite(symTbl), To: i.To.STRewrite(symTbl)}
}

func (i If) STRewrite(symTbl SymTbl) Type {
	return If{Condition: i.Condition.STRewrite(symTbl), TrueCase: i.TrueCase.STRewrite(symTbl)}
}

func (i IfElse) STRewrite(symTbl SymTbl) Type {
	return IfElse{Condition: i.Condition.STRewrite(symTbl), TrueCase: i.TrueCase.STRewrite(symTbl), FalseCase: i.FalseCase.STRewrite(symTbl)}
}

func (w While) STRewrite(symTbl SymTbl) Type {
	return While{Condition: w.Condition.STRewrite(symTbl), Body: w.Body.STRewrite(symTbl)}
}

func (f For) STRewrite(symTbl SymTbl) Type {
	iterator := f.Iterators.STRewrite(symTbl).(List)

	if len(symTbl) < 1 {
		return For{VarRefs: f.VarRefs, Iterators: iterator, Body: f.Body.STRewrite(symTbl)}
	}

	varRefs := []Type{}
	for _, varRef := range f.VarRefs.Elems {
		varRef := varRef.(Name)
		name := string(varRef)

		ix, ok := symTbl[len(symTbl)-1][name]
		if !ok {
			l := len(symTbl[len(symTbl)-1])
			symTbl[len(symTbl)-1][name] = l
			ix = l
		}
		varRefs = append(varRefs, Local{Ix: ix, VarName: name})
	}

	return For{VarRefs: List{Elems: varRefs}, Iterators: iterator, Body: f.Body.STRewrite(symTbl)}
}

func (r Return) STRewrite(symTbl SymTbl) Type {
	return Return{Target: r.Target.STRewrite(symTbl)}
}

func (y Yield) STRewrite(symTbl SymTbl) Type {
	return Yield{Target: y.Target.STRewrite(symTbl)}
}

func (n Name) STRewrite(symTbl SymTbl) Type {
	name := string(n)

	// look up variable at local scope
	if len(symTbl) > 0 {
		if ix, ok := symTbl[len(symTbl)-1][string(n)]; ok {
			return Local{Ix: ix, VarName: name}
		}
	}

	// look up variable in the enclosing lexical scope
	if len(symTbl) > 1 {
		if ix, ok := symTbl[len(symTbl)-2][string(n)]; ok {
			return Closure{Ix: ix, VarName: name}
		}
	}

	// variable not defined in either, assume global variable
	return n
}

func (a Assign) STRewrite(symTbl SymTbl) Type {
	value := a.Value.STRewrite(symTbl)
	varRef := a.VarRef.(Name)
	name := string(varRef)

	if len(symTbl) < 1 {
		return Assign{VarRef: varRef, Value: value}
	}

	ix, ok := symTbl[len(symTbl)-1][name]
	if !ok {
		l := len(symTbl[len(symTbl)-1])
		symTbl[len(symTbl)-1][name] = l
		ix = l
	}

	return Assign{VarRef: Local{Ix: ix, VarName: name}, Value: value}
}

func (l Local) STRewrite(_ SymTbl) Type   { panic("STRewrite called on local") }
func (c Closure) STRewrite(_ SymTbl) Type { panic("STRewrite called on closure") }

func (b Block) STRewrite(symTbl SymTbl) Type {
	body := []Type{}

	for _, t := range b.Body {
		body = append(body, t.STRewrite(symTbl))
	}

	return Block{Body: body}
}

func (l List) STRewrite(symTbl SymTbl) Type {
	elems := []Type{}

	for _, t := range l.Elems {
		elems = append(elems, t.STRewrite(symTbl))
	}

	return List{Elems: elems}
}

func (r Read) STRewrite(_ SymTbl) Type       { return r }
func (w Write) STRewrite(symTbl SymTbl) Type { return Write{Value: w.Value.STRewrite(symTbl)} }
func (a Aton) STRewrite(symTbl SymTbl) Type  { return Aton{Value: a.Value.STRewrite(symTbl)} }
func (t Toa) STRewrite(symTbl SymTbl) Type   { return Toa{Value: t.Value.STRewrite(symTbl)} }
func (e Exit) STRewrite(symTbl SymTbl) Type  { return Exit{Value: e.Value.STRewrite(symTbl)} }
