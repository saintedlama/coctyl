package validation_tests

import (
	"testing"

	"github.com/saintedlama/coctyl/pkg/coctyl"
)

// hashSource parses and hashes source code with optional options.
func hashSource(t *testing.T, src string, opts ...coctyl.Options) string {
	t.Helper()
	opt := coctyl.Options{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	res, err := coctyl.HashSource([]byte(src), opt)
	if err != nil {
		t.Fatalf("unexpected HashSource error: %v\nSource:\n%s", err, src)
	}
	return res.Hash
}

// assertEquivalent verifies that src1 and src2 produce identical dactyl hashes.
func assertEquivalent(t *testing.T, name, src1, src2 string, opts ...coctyl.Options) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		h1 := hashSource(t, src1, opts...)
		h2 := hashSource(t, src2, opts...)
		if h1 != h2 {
			t.Errorf("expected IDENTICAL hashes for %q, but got:\nHash 1: %s\nHash 2: %s", name, h1, h2)
		}
	})
}

// assertDifferent verifies that src1 and src2 produce different dactyl hashes.
func assertDifferent(t *testing.T, name, src1, src2 string, opts ...coctyl.Options) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		h1 := hashSource(t, src1, opts...)
		h2 := hashSource(t, src2, opts...)
		if h1 == h2 {
			t.Errorf("expected DIFFERENT hashes for %q, but both produced identical hash: %s", name, h1)
		}
	})
}
