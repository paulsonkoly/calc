package node

import "github.com/paulsonkoly/calc/types/value"

// Constanter converts a constant node to a value.
//
// A constant node is something that can be put in the data segment. Function
// literal can't be a constant node as they require code.
type Constanter interface {
	Constant() (value.Type, bool) // ToValue converts a constant node to a value.
}

func (i Invalid) ToValue() (value.Type, bool)  { return value.Nil, false }
func (c Call) Constant() (value.Type, bool)     { return value.Nil, false }
func (f Function) Constant() (value.Type, bool) { return value.Nil, false }
func (i Int) Constant() (value.Type, bool)      { return value.NewInt(int(i)), true }
func (f Float) Constant() (value.Type, bool)    { return value.NewFloat(float64(f)), true }
func (s String) Constant() (value.Type, bool)   { return value.NewString(string(s)), true }
func (b Bool) Constant() (value.Type, bool)     { return value.NewBool(bool(b)), true }
func (l List) Constant() (value.Type, bool) {
	ary := make([]value.Type, 0, len(l.Elems))
	for _, t := range l.Elems {
		v, ok := t.Constant()
		if !ok {
			return value.Nil, false
		}
		ary = append(ary, v)
	}
	return value.NewArray(ary), true
}
func (b BinOp) Constant() (value.Type, bool)       { return value.Nil, false }
func (a Assign) Constant() (value.Type, bool)      { return value.Nil, false }
func (u UnOp) Constant() (value.Type, bool)        { return value.Nil, false }
func (u IndexAt) Constant() (value.Type, bool)     { return value.Nil, false }
func (u IndexFromTo) Constant() (value.Type, bool) { return value.Nil, false }
func (i If) Constant() (value.Type, bool)          { return value.Nil, false }
func (i IfElse) Constant() (value.Type, bool)      { return value.Nil, false }
func (w While) Constant() (value.Type, bool)       { return value.Nil, false }
func (f For) Constant() (value.Type, bool)         { return value.Nil, false }
func (r Return) Constant() (value.Type, bool)      { return value.Nil, false }
func (y Yield) Constant() (value.Type, bool)       { return value.Nil, false }
func (r Read) Constant() (value.Type, bool)        { return value.Nil, false }
func (w Write) Constant() (value.Type, bool)       { return value.Nil, false }
func (a Aton) Constant() (value.Type, bool)        { return value.Nil, false }
func (t Toa) Constant() (value.Type, bool)         { return value.Nil, false }
func (n Name) Constant() (value.Type, bool)        { return value.Nil, false }
func (l Local) Constant() (value.Type, bool)       { return value.Nil, false }
func (c Closure) Constant() (value.Type, bool)     { return value.Nil, false }
func (b Block) Constant() (value.Type, bool)       { return value.Nil, false }
func (e Exit) Constant() (value.Type, bool)        { return value.Nil, false }
