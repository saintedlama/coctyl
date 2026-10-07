package validation_tests

import (
	"testing"
)

func TestEquivalenceIdentifierRenaming(t *testing.T) {
	assertEquivalent(t, "LocalShortVariableDeclarations",
		`package a
func calculate() int {
	a := 10
	b := 20
	c := a + b
	return c
}`,
		`package b
func compute() int {
	x := 10
	y := 20
	z := x + y
	return z
}`)

	assertEquivalent(t, "LocalVarDeclarations",
		`package p1
func run() int {
	var alpha int = 5
	var beta int = 15
	return alpha * beta
}`,
		`package p2
func execute() int {
	var first int = 5
	var second int = 15
	return first * second
}`)

	assertEquivalent(t, "GroupedVarDeclarations",
		`package p1
func process() int {
	var (
		count = 1
		total = 100
	)
	return count + total
}`,
		`package p2
func process() int {
	var (
		cnt = 1
		sum = 100
	)
	return cnt + sum
}`)

	assertEquivalent(t, "FunctionParametersRenamed",
		`package p1
func divide(dividend, divisor int) int {
	return dividend / divisor
}`,
		`package p2
func quotient(numerator, denominator int) int {
	return numerator / denominator
}`)

	assertEquivalent(t, "NamedReturnValuesRenamed",
		`package p1
func split(total int) (left, right int) {
	left = total / 2
	right = total - left
	return left, right
}`,
		`package p2
func partition(sum int) (first, second int) {
	first = sum / 2
	second = sum - first
	return first, second
}`)

	assertEquivalent(t, "NakedReturnsWithRenamedVariables",
		`package p1
func counter() (result int) {
	result = 42
	return
}`,
		`package p2
func tally() (ans int) {
	ans = 42
	return
}`)

	assertEquivalent(t, "MethodValueReceiverRenamed",
		`package p1
type Counter struct{ n int }
func (c Counter) Value() int {
	return c.n
}`,
		`package p2
type Counter struct{ n int }
func (self Counter) Value() int {
	return self.n
}`)

	assertEquivalent(t, "MethodPointerReceiverRenamed",
		`package p1
type Service struct{ active bool }
func (s *Service) IsActive() bool {
	return s.active
}`,
		`package p2
type Service struct{ active bool }
func (srv *Service) IsActive() bool {
	return srv.active
}`)
}

func TestEquivalenceControlFlowBindings(t *testing.T) {
	assertEquivalent(t, "ForClassicLoopIndexRenamed",
		`package p1
func sum(items []int) int {
	total := 0
	for i := 0; i < len(items); i++ {
		total += items[i]
	}
	return total
}`,
		`package p2
func aggregate(nums []int) int {
	acc := 0
	for idx := 0; idx < len(nums); idx++ {
		acc += nums[idx]
	}
	return acc
}`)

	assertEquivalent(t, "ForRangeWithKeyAndValueRenamed",
		`package p1
func keysValues(m map[string]int) int {
	sum := 0
	for k, v := range m {
		_ = k
		sum += v
	}
	return sum
}`,
		`package p2
func traverse(dict map[string]int) int {
	total := 0
	for key, val := range dict {
		_ = key
		total += val
	}
	return total
}`)

	assertEquivalent(t, "ForRangeWithBlankKey",
		`package p1
func total(nums []int) int {
	s := 0
	for _, n := range nums {
		s += n
	}
	return s
}`,
		`package p2
func total(list []int) int {
	sum := 0
	for _, item := range list {
		sum += item
	}
	return sum
}`)

	assertEquivalent(t, "IfWithInitStatementVariableRenamed",
		`package p1
func check(fn func() error) bool {
	if err := fn(); err != nil {
		return false
	}
	return true
}`,
		`package p2
func verify(task func() error) bool {
	if failure := task(); failure != nil {
		return false
	}
	return true
}`)

	assertEquivalent(t, "SwitchWithInitStatementRenamed",
		`package p1
func evaluate(val int) string {
	switch mod := val % 2; mod {
	case 0:
		return "even"
	default:
		return "odd"
	}
}`,
		`package p2
func classify(n int) string {
	switch remainder := n % 2; remainder {
	case 0:
		return "even"
	default:
		return "odd"
	}
}`)

	assertEquivalent(t, "TypeSwitchBindingRenamed",
		`package p1
func inspect(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case string:
		return len(val)
	default:
		return 0
	}
}`,
		`package p2
func examine(anyVal interface{}) int {
	switch concrete := anyVal.(type) {
	case int:
		return concrete
	case string:
		return len(concrete)
	default:
		return 0
	}
}`)
}

func TestEquivalenceTuplesAndUnpacking(t *testing.T) {
	assertEquivalent(t, "MapLookupCommaOkRenamed",
		`package p1
func lookup(m map[string]int, key string) (int, bool) {
	val, ok := m[key]
	return val, ok
}`,
		`package p2
func fetch(tbl map[string]int, k string) (int, bool) {
	res, found := tbl[k]
	return res, found
}`)

	assertEquivalent(t, "ChannelReceiveCommaOkRenamed",
		`package p1
func drain(ch <-chan int) (int, bool) {
	msg, open := <-ch
	return msg, open
}`,
		`package p2
func readNext(in <-chan int) (int, bool) {
	elem, active := <-in
	return elem, active
}`)

	assertEquivalent(t, "TypeAssertionCommaOkRenamed",
		`package p1
func cast(x interface{}) (string, bool) {
	str, success := x.(string)
	return str, success
}`,
		`package p2
func convert(obj interface{}) (string, bool) {
	text, valid := obj.(string)
	return text, valid
}`)
}

func TestEquivalenceFormattingAndComments(t *testing.T) {
	assertEquivalent(t, "RedundantParenthesesFormatting",
		`package p1
func calc(a, b, c int) int {
	return ((a + b) * (c + 1))
}`,
		`package p2
func calc(x, y, z int) int {
	return (x + y) * (z + 1)
}`)

	assertEquivalent(t, "CommentsAndDocstringInvariance",
		`// Package p1 implements math
package p1

/* Multi-line
   package explanation */

// Add sums two numbers
func Add(x, y int) int {
	// Add x and y
	res := x + y /* inline comment */
	return res // done
}`,
		`package p2

func Add(a, b int) int {
	out := a + b
	return out
}`)

	assertEquivalent(t, "WhitespaceAndIndentationInvariance",
		`package p1
func f(x int) int {


	if x > 0 {

		return x * 2

	}


	return 0
}`,
		`package p2
func f(a int) int {
	if a > 0 {
		return a * 2
	}
	return 0
}`)
}

func TestEquivalenceGenerics(t *testing.T) {
	assertEquivalent(t, "GenericFunctionTypeParamsRenamed",
		`package p1
func MapSlice[T, R any](in []T, f func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}`,
		`package p2
func Transform[In, Out any](list []In, op func(In) Out) []Out {
	res := make([]Out, len(list))
	for idx, elt := range list {
		res[idx] = op(elt)
	}
	return res
}`)

	assertEquivalent(t, "GenericConstraintInterfaceRenamed",
		`package p1
func Equal[T comparable](a, b T) bool {
	return a == b
}`,
		`package p2
func Identical[Elem comparable](x, y Elem) bool {
	return x == y
}`)
}
