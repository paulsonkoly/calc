package flags

import "flag"

var ByteCodeFlag = flag.Bool("bytecode", false, "calc prints expression bytecode")
var AstFlag = flag.Bool("ast", false, `calc outputs AST in graphviz dot format
% ./cmd --ast ../examples/euler_35.calc > x.dot # remove any output values
% gvpack -u x.dot > packed.dot
% dot -Tsvg packed.dot -o x.svg`)
var EvalFlag = flag.String("eval", "", "string to evaluate")
var CPUProfFlag = flag.String("cpuprof", "", "filename for go pprof")
