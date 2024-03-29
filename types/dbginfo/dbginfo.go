// Package dbginfo defines types holding debug information produced at
// compilation time useful at runtime.
package dbginfo

// Type describes a function call for stack dumping.
type Call struct {
	Name   string // variable holding the function when called
	ArgCnt int    // number of function arguments
}

// Type maps instruction pointer to Call.
type Type map[int]Call
