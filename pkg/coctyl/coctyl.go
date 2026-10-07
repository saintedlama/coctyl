package coctyl

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

// HashFile parses and computes the dactyl for a Go file on disk.
func HashFile(filePath string, opts Options) (*FileResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}

	result, err := HashSource(content, opts)
	if err != nil {
		return nil, fmt.Errorf("hashing file %s: %w", filePath, err)
	}
	result.Path = filePath
	return result, nil
}

// HashSource parses and computes the structural dactyl for Go source code in bytes
// using the two-pass approach:
// Pass 1: Normalizer transforms AST (scope resolution, alpha-conversion, import/package filtering).
// Pass 2: Serializer converts the normalized AST into canonical S-expression string.
func HashSource(src []byte, opts Options) (*FileResult, error) {
	canonical, err := CanonicalASTSource(src, opts)
	if err != nil {
		return nil, err
	}

	return &FileResult{
		Hash: HashString(canonical),
	}, nil
}

// CanonicalASTFile parses a Go file on disk and returns its normalized canonical S-expression string.
func CanonicalASTFile(filePath string, opts Options) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", filePath, err)
	}
	return CanonicalASTSource(content, opts)
}

// CanonicalASTSource parses Go source bytes and returns its normalized canonical S-expression string
// using the two-pass pipeline (Pass 1: Normalizer, Pass 2: Serializer).
func CanonicalASTSource(src []byte, opts Options) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return "", fmt.Errorf("parsing Go source: %w", err)
	}

	// Pass 1: AST Normalization & Scope-aware Alpha-conversion
	normalized := NewNormalizer(opts).Normalize(file)

	// Pass 2: Stateless Canonical Serialization
	canonical := NewSerializer().Serialize(normalized)
	return canonical, nil
}
