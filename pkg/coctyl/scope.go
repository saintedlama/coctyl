package coctyl

import "fmt"

// Scope tracks local identifier bindings in a lexical environment.
type Scope struct {
	parent  *Scope
	symbols map[string]string // original local identifier -> canonical symbol ($0, $1, ...)
	nextID  *int              // shared monotonic counter across the function scope hierarchy
}

// NewRootScope creates a new top-level lexical scope for a function.
func NewRootScope() *Scope {
	counter := 0
	return &Scope{
		parent:  nil,
		symbols: make(map[string]string),
		nextID:  &counter,
	}
}

// NewChild creates an inner lexical scope inheriting the variable counter.
func (s *Scope) NewChild() *Scope {
	return &Scope{
		parent:  s,
		symbols: make(map[string]string),
		nextID:  s.nextID,
	}
}

// Define allocates the next canonical slot ($0, $1, ...) to name within this scope.
// If name is blank ("_"), it returns "_" without consuming a slot.
func (s *Scope) Define(name string) string {
	if name == "" || name == "_" {
		return "_"
	}
	canonical := fmt.Sprintf("$%d", *s.nextID)
	*s.nextID++
	s.symbols[name] = canonical
	return canonical
}

// DefineOrReuse binds name in the current immediate scope. If name was already declared
// in the immediate scope (e.g. Go short variable := redeclaration where at least one variable is new),
// it reuses the existing canonical slot. If shadowing an outer scope, a new slot is allocated.
func (s *Scope) DefineOrReuse(name string) string {
	if name == "" || name == "_" {
		return "_"
	}
	if canonical, ok := s.symbols[name]; ok {
		return canonical
	}
	return s.Define(name)
}

// Lookup searches for an identifier in the scope hierarchy.
// Returns the canonical name and true if found, or empty string and false if undeclared locally.
func (s *Scope) Lookup(name string) (string, bool) {
	if name == "_" {
		return "_", true
	}
	for curr := s; curr != nil; curr = curr.parent {
		if canonical, ok := curr.symbols[name]; ok {
			return canonical, true
		}
	}
	return "", false
}
