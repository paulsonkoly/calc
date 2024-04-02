package node

// Depther determines the depth of binary or unary operators
type Depther interface {
	Depth() int
}

func (c Call) Depth() int        { return 0 }
func (f Function) Depth() int    { return 0 }
func (i Int) Depth() int         { return 0 }
func (f Float) Depth() int       { return 0 }
func (s String) Depth() int      { return 0 }
func (b Bool) Depth() int        { return 0 }
func (l List) Depth() int        { return 0 }
func (b BinOp) Depth() int       { return b.Left.Depth() + 1 }
func (a Assign) Depth() int      { return 0 }
func (u UnOp) Depth() int        { return u.Target.Depth() + 1 }
func (u IndexAt) Depth() int     { return 0 }
func (u IndexFromTo) Depth() int { return 0 }
func (i If) Depth() int          { return 0 }
func (i IfElse) Depth() int      { return 0 }
func (w While) Depth() int       { return 0 }
func (f For) Depth() int         { return 0 }
func (r Return) Depth() int      { return 0 }
func (y Yield) Depth() int       { return 0 }
func (r Read) Depth() int        { return 0 }
func (w Write) Depth() int       { return 0 }
func (a Aton) Depth() int        { return 0 }
func (t Toa) Depth() int         { return 0 }
func (n Name) Depth() int        { return 0 }
func (l Local) Depth() int       { return 0 }
func (c Closure) Depth() int     { return 0 }
func (b Block) Depth() int       { return 0 }
func (e Exit) Depth() int        { return 0 }
