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
	option() opt
	label() string
}

func (i Invalid) option() opt     { return defaultOpts }
func (c Call) option() opt        { return defaultOpts }
func (f Function) option() opt    { return defaultOpts }
func (i Int) option() opt         { return constOpts }
func (f Float) option() opt       { return constOpts }
func (s String) option() opt      { return constOpts }
func (b Bool) option() opt        { return constOpts }
func (l List) option() opt        { return constOpts }
func (b BinOp) option() opt       { return opteratorOpts }
func (a Assign) option() opt      { return defaultOpts }
func (u UnOp) option() opt        { return opteratorOpts }
func (u IndexAt) option() opt     { return opteratorOpts }
func (u IndexFromTo) option() opt { return opteratorOpts }
func (i If) option() opt          { return defaultOpts }
func (i IfElse) option() opt      { return defaultOpts }
func (w While) option() opt       { return defaultOpts }
func (f For) option() opt         { return defaultOpts }
func (r Return) option() opt      { return defaultOpts }
func (y Yield) option() opt       { return defaultOpts }
func (r Read) option() opt        { return defaultOpts }
func (w Write) option() opt       { return defaultOpts }
func (a Aton) option() opt        { return defaultOpts }
func (t Toa) option() opt         { return defaultOpts }
func (n Name) option() opt        { return variableOpts }
func (l Local) option() opt       { return variableOpts }
func (c Closure) option() opt     { return variableOpts }
func (b Block) option() opt       { return defaultOpts }
func (e Error) option() opt       { return defaultOpts }
func (e Exit) option() opt        { return defaultOpts }

func (i Invalid) label() string     { return fmt.Sprintf("%T", i) }
func (c Call) label() string        { return fmt.Sprintf("%T", c) }
func (f Function) label() string    { return fmt.Sprintf("%T", f) }
func (i Int) label() string         { return fmt.Sprintf("int:%v", i) }
func (f Float) label() string       { return fmt.Sprintf("float:%v", f) }
func (s String) label() string      { return strings.Trim(string(s), "\"") }
func (b Bool) label() string        { return fmt.Sprint(b) }
func (l List) label() string        { return "[]" }
func (b BinOp) label() string       { return b.Op }
func (a Assign) label() string      { return fmt.Sprintf("%T", a) }
func (u UnOp) label() string        { return u.Op }
func (u IndexAt) label() string     { return "@" }
func (u IndexFromTo) label() string { return "@" }
func (i If) label() string          { return fmt.Sprintf("%T", i) }
func (i IfElse) label() string      { return fmt.Sprintf("%T", i) }
func (w While) label() string       { return fmt.Sprintf("%T", w) }
func (f For) label() string         { return fmt.Sprintf("%T", f) }
func (r Return) label() string      { return fmt.Sprintf("%T", r) }
func (y Yield) label() string       { return fmt.Sprintf("%T", y) }
func (r Read) label() string        { return fmt.Sprintf("%T", r) }
func (w Write) label() string       { return fmt.Sprintf("%T", w) }
func (a Aton) label() string        { return fmt.Sprintf("%T", a) }
func (t Toa) label() string         { return fmt.Sprintf("%T", t) }
func (n Name) label() string        { return string(n) }
func (l Local) label() string       { return fmt.Sprintf("lvar:%d", int(l)) }
func (c Closure) label() string     { return fmt.Sprintf("cvar:%d", int(c)) }
func (b Block) label() string       { return fmt.Sprintf("%T", b) }
func (e Error) label() string       { return fmt.Sprintf("%T", e) }
func (e Exit) label() string        { return fmt.Sprintf("%T", e) }

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

func graphviz(parentID int, arrowLbl string, t graphvizzer) {
	id := nodecnt
	if parentID >= 0 {
		fmt.Printf("node_%d -> node_%d [label=\"%s\"]\n", parentID, id, arrowLbl)
	}

	o := t.option()

	fmt.Printf("node_%d [label=\"%s\" %v]\n", id, t.label(), o)

	nodecnt++

	for k, t2 := range children(t) {
		graphviz(id, k, t2)
	}
}
