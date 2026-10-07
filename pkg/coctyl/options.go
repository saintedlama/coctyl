package coctyl

// Options configures the dactyl generation.
type Options struct {
	// IncludeImports, if true, includes import declarations in the dactyl.
	// By default (false), imports are ignored so code parts can move dynamically.
	IncludeImports bool

	// IncludePackage, if true, includes the package declaration in the dactyl.
	// By default (false), package declarations are ignored to ensure functional equality.
	IncludePackage bool
}
