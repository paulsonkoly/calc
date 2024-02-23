package node

import (
	"fmt"
	"reflect"
	"strings"
)

// Graphviz writes dot format graphviz output of AST to stdout
//
// print function for debugging calc. It prints an AST transforming it into a
// graphviz dotfile. One can convert the resulting file into an svg or other
// image formats.
//
//	% ./cmd --ast ../examples/euler_35.calc > x.dot
//	% gvpack -u x.dot > packed.dot
//	% dot -Tsvg packed.dot -o x.svg
func Graphviz(top graphvizzer) {
	fmt.Println("digraph AST {")
	graphviz(-1, "", top)
	fmt.Println("}")
}

type opt struct {
	shape     string
	style     string
	color     string
	fillcolor string
}

var (
	defaultOpts   = opt{shape: "box"}
	constOpts     = opt{shape: "box", style: "filled", color: "bisque4", fillcolor: "bisque"}
	variableOpts  = opt{shape: "box", style: "filled", color: "darkgoldenrod", fillcolor: "darkgoldenrod1"}
	opteratorOpts = opt{shape: "circle", style: "filled", color: "lightblue", fillcolor: "lightblue1"}
)

func (o opt) String() string {
	return fmt.Sprintf("shape=\"%s\" style=\"%s\" color=\"%s\" fillcolor=\"%s\"", o.shape, o.style, o.color, o.fillcolor)
}

type graphvizzer interface {
	Option() opt
	Label() string
}

func (i Invalid) Option() opt     { return defaultOpts }
func (c Call) Option() opt        { return defaultOpts }
func (f Function) Option() opt    { return defaultOpts }
func (i Int) Option() opt         { return constOpts }
func (f Float) Option() opt       { return constOpts }
func (s String) Option() opt      { return constOpts }
func (b Bool) Option() opt        { return constOpts }
func (l List) Option() opt        { return constOpts }
func (b BinOp) Option() opt       { return opteratorOpts }
func (a Assign) Option() opt      { return defaultOpts }
func (u UnOp) Option() opt        { return opteratorOpts }
func (u IndexAt) Option() opt     { return opteratorOpts }
func (u IndexFromTo) Option() opt { return opteratorOpts }
func (i If) Option() opt          { return defaultOpts }
func (i IfElse) Option() opt      { return defaultOpts }
func (w While) Option() opt       { return defaultOpts }
func (r Return) Option() opt      { return defaultOpts }
func (r Read) Option() opt        { return defaultOpts }
func (w Write) Option() opt       { return defaultOpts }
func (a Aton) Option() opt        { return defaultOpts }
func (t Toa) Option() opt         { return defaultOpts }
func (n Name) Option() opt        { return variableOpts }
func (l Local) Option() opt       { return variableOpts }
func (c Closure) Option() opt     { return variableOpts }
func (b Block) Option() opt       { return defaultOpts }
func (r Repl) Option() opt        { return defaultOpts }
func (e Error) Option() opt       { return defaultOpts }

func (i Invalid) Label() string     { return fmt.Sprintf("%T", i) }
func (c Call) Label() string        { return fmt.Sprintf("%T", c) }
func (f Function) Label() string    { return fmt.Sprintf("%T", f) }
func (i Int) Label() string         { return fmt.Sprintf("int:%v", i) }
func (f Float) Label() string       { return fmt.Sprintf("float:%v", f) }
func (s String) Label() string      { return strings.Trim(string(s), "\"") }
func (b Bool) Label() string        { return b.Token() }
func (l List) Label() string        { return "[]" }
func (b BinOp) Label() string       { return b.Op }
func (a Assign) Label() string      { return fmt.Sprintf("%T", a) }
func (u UnOp) Label() string        { return u.Op }
func (u IndexAt) Label() string     { return "@" }
func (u IndexFromTo) Label() string { return "@" }
func (i If) Label() string          { return fmt.Sprintf("%T", i) }
func (i IfElse) Label() string      { return fmt.Sprintf("%T", i) }
func (w While) Label() string       { return fmt.Sprintf("%T", w) }
func (r Return) Label() string      { return fmt.Sprintf("%T", r) }
func (r Read) Label() string        { return fmt.Sprintf("%T", r) }
func (w Write) Label() string       { return fmt.Sprintf("%T", w) }
func (a Aton) Label() string        { return fmt.Sprintf("%T", a) }
func (t Toa) Label() string         { return fmt.Sprintf("%T", t) }
func (n Name) Label() string        { return string(n) }
func (l Local) Label() string       { return fmt.Sprintf("lvar:%d", int(l)) }
func (c Closure) Label() string     { return fmt.Sprintf("cvar:%d", int(c)) }
func (b Block) Label() string       { return fmt.Sprintf("%T", b) }
func (r Repl) Label() string        { return fmt.Sprintf("%T", r) }
func (e Error) Label() string       { return fmt.Sprintf("%T", e) }

func children(t graphvizzer) map[string]graphvizzer {
	typ := reflect.TypeOf(t)
	r := map[string]graphvizzer{}

	if typ.Kind() == reflect.Struct {
		for _, sf := range reflect.VisibleFields(typ) {
			field := reflect.ValueOf(t).FieldByName(sf.Name)
			if fType, ok := field.Interface().(graphvizzer); ok {
				r[sf.Name] = fType
				continue
			}

			if fType, ok := field.Interface().([]Type); ok {
				for i, g := range fType {
					r[fmt.Sprintf("%d", i)] = g
				}
			}
		}
	}
	return r
}

var nodecnt int

func graphviz(parentId int, arrowLbl string, t graphvizzer) {
	id := nodecnt
	if parentId >= 0 {
		fmt.Printf("node_%d -> node_%d [label=\"%s\"]\n", parentId, id, arrowLbl)
	}

	o := t.Option()

	fmt.Printf("node_%d [label=\"%s\" %v]\n", id, t.Label(), o)

	nodecnt++

	for k, t2 := range children(t) {
		graphviz(id, k, t2)
	}
}
