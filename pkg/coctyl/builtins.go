package coctyl

// predeclaredIdentifiers contains Go's predeclared types, constants, and built-in functions.
// These should not be alpha-normalized unless shadowed by a local variable.
var predeclaredIdentifiers = map[string]bool{
	// Constants
	"true":  true,
	"false": true,
	"iota":  true,
	"nil":   true,

	// Types
	"any":        true,
	"bool":       true,
	"byte":       true,
	"comparable": true,
	"complex64":  true,
	"complex128": true,
	"error":      true,
	"float32":    true,
	"float64":    true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"rune":       true,
	"string":     true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,

	// Functions
	"append":  true,
	"cap":     true,
	"clear":   true,
	"close":   true,
	"complex": true,
	"copy":    true,
	"delete":  true,
	"imag":    true,
	"len":     true,
	"make":    true,
	"max":     true,
	"min":     true,
	"new":     true,
	"panic":   true,
	"print":   true,
	"println": true,
	"real":    true,
	"recover": true,
}

// IsPredeclared returns true if the name is a Go predeclared identifier.
func IsPredeclared(name string) bool {
	return predeclaredIdentifiers[name]
}
