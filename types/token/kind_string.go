// Code generated by "stringer -type Kind token.go"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Invalid-0]
	_ = x[EOL-1]
	_ = x[EOF-2]
	_ = x[IntLit-3]
	_ = x[FloatLit-4]
	_ = x[StringLit-5]
	_ = x[Name-6]
	_ = x[Sticky-7]
	_ = x[NotSticky-8]
}

const _Kind_name = "InvalidEOLEOFIntLitFloatLitStringLitNameStickyNotSticky"

var _Kind_index = [...]uint8{0, 7, 10, 13, 19, 27, 36, 40, 46, 55}

func (i Kind) String() string {
	if i < 0 || i >= Kind(len(_Kind_index)-1) {
		return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}
