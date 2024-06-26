// Package compresult defines the compilation result.
package compresult

import (
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/dbginfo"
	"github.com/paulsonkoly/calc/types/value"
)

// Type is the compilation result.
type Type struct {
	CS  *[]bytecode.Type // Code segment
	DS  *[]value.Type    // Data segment
	Dbg *dbginfo.Type    // Debug info
}
