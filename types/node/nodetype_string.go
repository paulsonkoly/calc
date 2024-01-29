// Code generated by "stringer -type=NodeType node.go"; DO NOT EDIT.

package node

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Invalid-0]
	_ = x[Call-1]
	_ = x[Int-2]
	_ = x[Float-3]
	_ = x[Op-4]
	_ = x[If-5]
	_ = x[While-6]
	_ = x[Return-7]
	_ = x[Name-8]
	_ = x[Block-9]
}

const _NodeType_name = "InvalidCallIntFloatOpIfWhileReturnNameBlock"

var _NodeType_index = [...]uint8{0, 7, 11, 14, 19, 21, 23, 28, 34, 38, 43}

func (i NodeType) String() string {
	if i < 0 || i >= NodeType(len(_NodeType_index)-1) {
		return "NodeType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _NodeType_name[_NodeType_index[i]:_NodeType_index[i+1]]
}
