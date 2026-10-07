package validation_tests

import (
	"testing"
)

func TestIntegrityOperandAndDataflowOrder(t *testing.T) {
	assertDifferent(t, "SubtractionOperandOrderSwapped",
		`package p
func sub(a, b int) int {
	return a - b
}`,
		`package p
func sub(a, b int) int {
	return b - a
}`)

	assertDifferent(t, "DivisionOperandOrderSwapped",
		`package p
func div(x, y int) int {
	return x / y
}`,
		`package p
func div(x, y int) int {
	return y / x
}`)

	assertDifferent(t, "ComparisonOperandOrderSwapped",
		`package p
func cmp(x, y int) bool {
	return x < y
}`,
		`package p
func cmp(x, y int) bool {
	return y < x
}`)

	assertDifferent(t, "FunctionCallArgumentOrderSwapped",
		`package p
func caller(f func(int, string), count int, label string) {
	f(count, label)
}`,
		`package p
func caller(f func(string, int), count int, label string) {
	f(label, count)
}`)

	assertDifferent(t, "DataflowIdentitySameVarTwiceVsDistinct",
		`package p
func add(a, b int) int {
	return a + b
}`,
		`package p
func add(a, b int) int {
	return a + a
}`)

	assertDifferent(t, "MapAssignmentKeyValOrderSwapped",
		`package p
func store(m map[int]int, k, v int) {
	m[k] = v
}`,
		`package p
func store(m map[int]int, k, v int) {
	m[v] = k
}`)
}

func TestIntegrityOperators(t *testing.T) {
	assertDifferent(t, "PlusVsMinus",
		`package p
func calc(a, b int) int { return a + b }`,
		`package p
func calc(a, b int) int { return a - b }`)

	assertDifferent(t, "MultiplyVsDivide",
		`package p
func calc(a, b int) int { return a * b }`,
		`package p
func calc(a, b int) int { return a / b }`)

	assertDifferent(t, "EqualVsNotEqual",
		`package p
func test(a, b int) bool { return a == b }`,
		`package p
func test(a, b int) bool { return a != b }`)

	assertDifferent(t, "LessThanVsLessOrEqual",
		`package p
func test(a, b int) bool { return a < b }`,
		`package p
func test(a, b int) bool { return a <= b }`)

	assertDifferent(t, "LogicalAndVsLogicalOr",
		`package p
func test(a, b bool) bool { return a && b }`,
		`package p
func test(a, b bool) bool { return a || b }`)

	assertDifferent(t, "BitwiseAndVsBitwiseOr",
		`package p
func test(a, b int) int { return a & b }`,
		`package p
func test(a, b int) int { return a | b }`)

	assertDifferent(t, "BitwiseXorVsBitwiseClear",
		`package p
func test(a, b int) int { return a ^ b }`,
		`package p
func test(a, b int) int { return a &^ b }`)

	assertDifferent(t, "IncrementVsDecrement",
		`package p
func step(x int) int { x++; return x }`,
		`package p
func step(x int) int { x--; return x }`)
}

func TestIntegrityLiterals(t *testing.T) {
	assertDifferent(t, "IntegerConstantsDiffer",
		`package p
func get() int { return 100 }`,
		`package p
func get() int { return 200 }`)

	assertDifferent(t, "StringConstantsDiffer",
		`package p
func msg() string { return "success" }`,
		`package p
func msg() string { return "failure" }`)

	assertDifferent(t, "BooleanConstantsDiffer",
		`package p
func flag() bool { return true }`,
		`package p
func flag() bool { return false }`)

	assertDifferent(t, "FloatConstantsDiffer",
		`package p
func pi() float64 { return 3.14 }`,
		`package p
func pi() float64 { return 3.14159 }`)
}

func TestIntegrityControlFlowAndExecution(t *testing.T) {
	assertDifferent(t, "IfElseBranchesInverted",
		`package p
func pick(c bool) int {
	if c {
		return 1
	} else {
		return 2
	}
}`,
		`package p
func pick(c bool) int {
	if c {
		return 2
	} else {
		return 1
	}
}`)

	assertDifferent(t, "LoopBoundaryConditionStrictVsInclusive",
		`package p
func loop(n int) int {
	s := 0
	for i := 0; i < n; i++ {
		s += i
	}
	return s
}`,
		`package p
func loop(n int) int {
	s := 0
	for i := 0; i <= n; i++ {
		s += i
	}
	return s
}`)

	assertDifferent(t, "BreakVsContinue",
		`package p
func run(items []int) {
	for _, it := range items {
		if it == 0 {
			break
		}
	}
}`,
		`package p
func run(items []int) {
	for _, it := range items {
		if it == 0 {
			continue
		}
	}
}`)

	assertDifferent(t, "GoroutineVsSynchronousCall",
		`package p
func execute(fn func()) {
	fn()
}`,
		`package p
func execute(fn func()) {
	go fn()
}`)

	assertDifferent(t, "DeferVsImmediateCall",
		`package p
func run(cleanup func()) {
	cleanup()
}`,
		`package p
func run(cleanup func()) {
	defer cleanup()
}`)
}

func TestIntegrityTypesAndChannels(t *testing.T) {
	assertDifferent(t, "ChannelSendVsReceive",
		`package p
func pipe(ch chan int, val int) {
	ch <- val
}`,
		`package p
func pipe(ch chan int, val int) {
	_ = <-ch
}`)

	assertDifferent(t, "ChannelDirectionBidirectionalVsSendOnly",
		`package p
func handle(ch chan int) {}`,
		`package p
func handle(ch chan<- int) {}`)

	assertDifferent(t, "SliceVsFixedArrayType",
		`package p
func proc(arr []int) int { return len(arr) }`,
		`package p
func proc(arr [5]int) int { return len(arr) }`)

	assertDifferent(t, "PointerVsValueReceiverType",
		`package p
type Node struct{ val int }
func (n Node) Value() int { return n.val }`,
		`package p
type Node struct{ val int }
func (n *Node) Value() int { return n.val }`)

	assertDifferent(t, "ReturnTupleTypesSwapped",
		`package p
func result() (int, string) { return 1, "ok" }`,
		`package p
func result() (string, int) { return "ok", 1 }`)

	assertDifferent(t, "StructFieldSelectorAccessDiffer",
		`package p
type User struct{ Name string; Role string }
func check(u User) string { return u.Name }`,
		`package p
type User struct{ Name string; Role string }
func check(u User) string { return u.Role }`)

	assertDifferent(t, "StructCompositeLiteralKeysDiffer",
		`package p
type Point struct{ X, Y int }
func origin() Point { return Point{X: 1} }`,
		`package p
type Point struct{ X, Y int }
func origin() Point { return Point{Y: 1} }`)

	assertDifferent(t, "ExternalPackageQualifiersDiffer",
		`package p
import "fmt"
func log(msg string) { fmt.Println(msg) }`,
		`package p
import "log"
func log(msg string) { log.Println(msg) }`)
}

func TestIntegrityEquivalenceCounterparts(t *testing.T) {
	baseline := `package a
func calculate() int {
	a := 10
	b := 20
	c := a + b
	return c
}`

	assertDifferent(t, "MutateFirstLiteral10to99", baseline,
		`package a
func calculate() int {
	a := 99
	b := 20
	c := a + b
	return c
}`)

	assertDifferent(t, "MutateSecondLiteral20to99", baseline,
		`package a
func calculate() int {
	a := 10
	b := 99
	c := a + b
	return c
}`)

	assertDifferent(t, "SwapInitialLiteralValues", baseline,
		`package a
func calculate() int {
	a := 20
	b := 10
	c := a + b
	return c
}`)

	assertDifferent(t, "SwapAdditionOperandOrder", baseline,
		`package a
func calculate() int {
	a := 10
	b := 20
	c := b + a
	return c
}`)

	assertDifferent(t, "RepeatOperandInAddition", baseline,
		`package a
func calculate() int {
	a := 10
	b := 20
	c := a + a
	return c
}`)

	assertDifferent(t, "ReturnDifferentVariableA", baseline,
		`package a
func calculate() int {
	a := 10
	b := 20
	c := a + b
	return a
}`)

	assertDifferent(t, "ReturnDifferentVariableB", baseline,
		`package a
func calculate() int {
	a := 10
	b := 20
	c := a + b
	return b
}`)

	assertDifferent(t, "MutateOperatorPlusToMinus", baseline,
		`package a
func calculate() int {
	a := 10
	b := 20
	c := a - b
	return c
}`)

	assertDifferent(t, "MutateOperatorPlusToMultiply", baseline,
		`package a
func calculate() int {
	a := 10
	b := 20
	c := a * b
	return c
}`)
}

func TestIntegrityFunctionParametersCounterparts(t *testing.T) {
	baseline := `package p
func divide(dividend, divisor int) int {
	return dividend / divisor
}`

	assertDifferent(t, "SwappedParameterEvaluationOrder", baseline,
		`package p
func divide(dividend, divisor int) int {
	return divisor / dividend
}`)

	assertDifferent(t, "RepeatedFirstParameter", baseline,
		`package p
func divide(dividend, divisor int) int {
	return dividend / dividend
}`)

	assertDifferent(t, "RepeatedSecondParameter", baseline,
		`package p
func divide(dividend, divisor int) int {
	return divisor / divisor
}`)

	assertDifferent(t, "ParameterTypeChanged", baseline,
		`package p
func divide(dividend, divisor int64) int {
	return int(dividend / divisor)
}`)
}
