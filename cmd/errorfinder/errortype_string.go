// Code generated by "stringer -type=ErrorType"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[errorTypeUnknown-0]
	_ = x[errorTypeSentinel-1]
	_ = x[errorTypeStructured-2]
}

const _ErrorType_name = "ErrorTypeUnknownErrorTypeSentinelErrorTypeStructured"

var _ErrorType_index = [...]uint8{0, 16, 33, 52}

func (i errorType) String() string {
	if i < 0 || i >= errorType(len(_ErrorType_index)-1) {
		return "ErrorType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ErrorType_name[_ErrorType_index[i]:_ErrorType_index[i+1]]
}
