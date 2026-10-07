package validation_tests

import (
	"testing"
)

func TestScopingNestedShadowing(t *testing.T) {
	assertEquivalent(t, "IfBlockShadowingEquivalence",
		`package p
func test() int {
	x := 10
	if x > 0 {
		x := 20
		return x
	}
	return x
}`,
		`package p
func test() int {
	a := 10
	if a > 0 {
		a := 20
		return a
	}
	return a
}`)

	assertDifferent(t, "InnerShadowedVarVsOuterVarReturn",
		`package p
func test() int {
	x := 10
	if x > 0 {
		x := 20
		return x
	}
	return x
}`,
		`package p
func test() int {
	x := 10
	if x > 0 {
		_ = 20
		return x
	}
	return x
}`)

	assertEquivalent(t, "ThreeLevelNestedShadowing",
		`package p
func deep() int {
	x := 1
	{
		x := 2
		{
			x := 3
			return x
		}
	}
}`,
		`package p
func deep() int {
	a := 1
	{
		a := 2
		{
			a := 3
			return a
		}
	}
}`)
}

func TestScopingShortVariableRedeclarations(t *testing.T) {
	assertEquivalent(t, "MultiVariableShortRedeclarationEquivalence",
		`package p
func run() int {
	x := 1
	x, y := 2, 3
	return x + y
}`,
		`package p
func run() int {
	a := 1
	a, b := 2, 3
	return a + b
}`)

	assertDifferent(t, "RedeclaredExistingVarVsNewDistinctVar",
		`package p
func run() int {
	x := 1
	x, y := 2, 3
	return x + y
}`,
		`package p
func run() int {
	x := 1
	z, y := 2, 3
	return x + y
}`)
}

func TestScopingBuiltinShadowing(t *testing.T) {
	assertEquivalent(t, "LocalVarShadowingBuiltinLen",
		`package p
func check() int {
	len := 42
	return len
}`,
		`package p
func check() int {
	size := 42
	return size
}`)

	assertDifferent(t, "BuiltinCallVsLocalShadowedVariable",
		`package p
func check(items []int) int {
	return len(items)
}`,
		`package p
func check(items []int) int {
	len := 100
	return len
}`)

	assertEquivalent(t, "LocalVarShadowingBuiltinInt",
		`package p
func typeShadow() int {
	int := 99
	return int
}`,
		`package p
func typeShadow() int {
	num := 99
	return num
}`)
}

func TestScopingClosuresAndCapture(t *testing.T) {
	assertEquivalent(t, "ClosureCapturingOuterScope",
		`package p
func factory() func() int {
	counter := 0
	return func() int {
		counter++
		return counter
	}
}`,
		`package p
func generator() func() int {
	tally := 0
	return func() int {
		tally++
		return tally
	}
}`)

	assertDifferent(t, "ClosureCapturedVarVsLocalInnerVar",
		`package p
func makeFn() func() int {
	x := 10
	return func() int {
		return x
	}
}`,
		`package p
func makeFn() func() int {
	x := 10
	_ = x
	return func() int {
		x := 20
		return x
	}
}`)

	assertEquivalent(t, "ClosureWithOwnParameters",
		`package p
func apply(fn func(a, b int) int) int {
	caller := func(x, y int) int {
		return fn(x, y)
	}
	return caller(1, 2)
}`,
		`package p
func apply(op func(i, j int) int) int {
	invoker := func(m, n int) int {
		return op(m, n)
	}
	return invoker(1, 2)
}`)
}

func TestScopingMethodReceiverVsParams(t *testing.T) {
	assertEquivalent(t, "ReceiverNameSameAsParameterInAnotherMethod",
		`package p
type Box struct{ value int }
func (b Box) Get() int { return b.value }
func (b Box) Set(bVal int) Box { b.value = bVal; return b }`,
		`package p
type Box struct{ value int }
func (self Box) Get() int { return self.value }
func (self Box) Set(newVal int) Box { self.value = newVal; return self }`)
}
