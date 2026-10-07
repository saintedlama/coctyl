package coctyl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saintedlama/coctyl/pkg/coctyl"
)

func hashSource(t *testing.T, src string) string {
	t.Helper()
	res, err := coctyl.HashSource([]byte(src), coctyl.Options{})
	if err != nil {
		t.Fatalf("failed to hash source: %v\nSource:\n%s", err, src)
	}
	return res.Hash
}

// ----------------------------------------------------------------------------
// Equivalence Tests: Files that MUST produce identical fingerprints
// ----------------------------------------------------------------------------

func TestUserExampleScopeOrderArithmetic(t *testing.T) {
	// Equivalent functional logic: a - b vs x - y with different package & func names
	src1 := `package p1
func calc() int {
	a := 1
	b := 2
	c := a - b
	return c
}`
	src2 := `package p2
func compute() int {
	x := 1
	y := 2
	z := x - y
	return z
}`
	// Swapped order: b - a
	src3 := `package p3
func computeReversed() int {
	x := 1
	y := 2
	z := y - x
	return z
}`

	h1 := hashSource(t, src1)
	h2 := hashSource(t, src2)
	h3 := hashSource(t, src3)

	if h1 != h2 {
		t.Errorf("expected identical hashes for renamed variables in same order: %s vs %s", h1, h2)
	}
	if h1 == h3 {
		t.Errorf("expected DIFFERENT hashes when variable subtraction order is reversed: got identical %s", h1)
	}
}

func TestPackageDeclarationIgnoredByDefault(t *testing.T) {
	src1 := `package alpha
func fn() int { return 42 }`

	src2 := `package beta
func fn() int { return 42 }`

	// Default: package declaration ignored -> identical hashes
	h1 := hashSource(t, src1)
	h2 := hashSource(t, src2)
	if h1 != h2 {
		t.Errorf("expected identical hashes when package declaration differs: %s vs %s", h1, h2)
	}

	// When IncludePackage: true -> hashes MUST differ
	optsWithPkg := coctyl.Options{IncludePackage: true}
	res1, _ := coctyl.HashSource([]byte(src1), optsWithPkg)
	res2, _ := coctyl.HashSource([]byte(src2), optsWithPkg)
	if res1.Hash == res2.Hash {
		t.Errorf("expected different hashes when IncludePackage: true")
	}
}

func TestAlphaEquivalenceRenamedVariables(t *testing.T) {
	src1 := `package main
func add(x, y int) int {
	res := x + y
	return res
}`
	src2 := `package main
func sum(alpha, beta int) int {
	gamma := alpha + beta
	return gamma
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("expected identical hashes, got %s and %s", h1, h2)
	}
}

func TestMultiVariableShortDeclarationRedeclaration(t *testing.T) {
	src1 := `package main
func process(a int) (int, int) {
	x := a
	x, y := 10, 20
	return x, y
}`
	src2 := `package main
func handle(input int) (int, int) {
	val := input
	val, other := 10, 20
	return val, other
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("expected identical hashes for redeclaration with renamed vars, got %s and %s", h1, h2)
	}
}

func TestForLoopVariants(t *testing.T) {
	// 1. Classic 3-part for loop
	src1Classic := `package main
func loopClassic(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}`
	src2Classic := `package main
func loopClassic(count int) int {
	total := 0
	for idx := 0; idx < count; idx++ {
		total += idx
	}
	return total
}`
	if h1, h2 := hashSource(t, src1Classic), hashSource(t, src2Classic); h1 != h2 {
		t.Errorf("classic for loop: expected %s, got %s", h1, h2)
	}

	// 2. For range with key and value
	src1Range := `package main
func rangeKV(items []string) string {
	var out string
	for k, v := range items {
		if k > 0 {
			out += v
		}
	}
	return out
}`
	src2Range := `package main
func rangeKV(elements []string) string {
	var result string
	for idx, elem := range elements {
		if idx > 0 {
			result += elem
		}
	}
	return result
}`
	if h1, h2 := hashSource(t, src1Range), hashSource(t, src2Range); h1 != h2 {
		t.Errorf("for range KV: expected %s, got %s", h1, h2)
	}

	// 3. For range with blank key
	src1BlankKey := `package main
func rangeBlank(items []int) int {
	s := 0
	for _, v := range items {
		s += v
	}
	return s
}`
	src2BlankKey := `package main
func rangeBlank(nums []int) int {
	acc := 0
	for _, n := range nums {
		acc += n
	}
	return acc
}`
	if h1, h2 := hashSource(t, src1BlankKey), hashSource(t, src2BlankKey); h1 != h2 {
		t.Errorf("for range blank key: expected %s, got %s", h1, h2)
	}

	// 4. For range without variables
	src1NoVars := `package main
func drain(ch chan int) {
	for range ch {}
}`
	src2NoVars := `package main
func drain(channel chan int) {
	for range channel {}
}`
	if h1, h2 := hashSource(t, src1NoVars), hashSource(t, src2NoVars); h1 != h2 {
		t.Errorf("for range no vars: expected %s, got %s", h1, h2)
	}
}

func TestIfElseScoping(t *testing.T) {
	src1 := `package main
func check(val int) int {
	if x := val * 2; x > 10 {
		return x
	} else {
		return x + 1
	}
}`
	src2 := `package main
func check(n int) int {
	if calc := n * 2; calc > 10 {
		return calc
	} else {
		return calc + 1
	}
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("if-else scoping: expected %s, got %s", h1, h2)
	}
}

func TestSwitchAndTypeSwitch(t *testing.T) {
	// Standard switch with init statement
	src1Switch := `package main
func handle(v int) string {
	switch x := v % 3; x {
	case 0:
		return "zero"
	case 1:
		return "one"
	default:
		return "other"
	}
}`
	src2Switch := `package main
func handle(input int) string {
	switch rem := input % 3; rem {
	case 0:
		return "zero"
	case 1:
		return "one"
	default:
		return "other"
	}
}`
	if h1, h2 := hashSource(t, src1Switch), hashSource(t, src2Switch); h1 != h2 {
		t.Errorf("switch with init: expected %s, got %s", h1, h2)
	}

	// Type switch
	src1TypeSwitch := `package main
func inspect(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case string:
		return len(val)
	default:
		return 0
	}
}`
	src2TypeSwitch := `package main
func inspect(item any) int {
	switch target := item.(type) {
	case int:
		return target
	case string:
		return len(target)
	default:
		return 0
	}
}`
	if h1, h2 := hashSource(t, src1TypeSwitch), hashSource(t, src2TypeSwitch); h1 != h2 {
		t.Errorf("type switch: expected %s, got %s", h1, h2)
	}
}

func TestSelectAndChannels(t *testing.T) {
	src1 := `package main
func multiplex(c1, c2 chan int, out chan int) {
	select {
	case msg := <-c1:
		out <- msg
	case msg := <-c2:
		out <- msg
	default:
		out <- 0
	}
}`
	src2 := `package main
func multiplex(chA, chB chan int, dest chan int) {
	select {
	case data := <-chA:
		dest <- data
	case data := <-chB:
		dest <- data
	default:
		dest <- 0
	}
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("select statement: expected %s, got %s", h1, h2)
	}
}

func TestGenericFunctions(t *testing.T) {
	src1 := `package main
func Filter[T any](items []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}`
	src2 := `package main
func Filter[Elem any](elements []Elem, matcher func(Elem) bool) []Elem {
	var out []Elem
	for _, elem := range elements {
		if matcher(elem) {
			out = append(out, elem)
		}
	}
	return out
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("generic functions: expected %s, got %s", h1, h2)
	}
}

func TestGenericIndexList(t *testing.T) {
	src1 := `package main
type Pair[K, V any] struct {
	First  K
	Second V
}
func Make[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{First: a, Second: b}
}`
	src2 := `package main
type Pair[K, V any] struct {
	First  K
	Second V
}
func Make[X, Y any](x X, y Y) Pair[X, Y] {
	return Pair[X, Y]{First: x, Second: y}
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("generic IndexList: expected %s, got %s", h1, h2)
	}
}

func TestReceiverVariations(t *testing.T) {
	// Pointer receiver renamed
	src1Ptr := `package main
type Server struct{ port int }
func (s *Server) GetPort() int {
	return s.port
}`
	src2Ptr := `package main
type Server struct{ port int }
func (srv *Server) GetPort() int {
	return srv.port
}`
	if h1, h2 := hashSource(t, src1Ptr), hashSource(t, src2Ptr); h1 != h2 {
		t.Errorf("pointer receiver: expected %s, got %s", h1, h2)
	}

	// Value receiver renamed
	src1Val := `package main
type Point struct{ X, Y int }
func (p Point) Sum() int {
	return p.X + p.Y
}`
	src2Val := `package main
type Point struct{ X, Y int }
func (pt Point) Sum() int {
	return pt.X + pt.Y
}`
	if h1, h2 := hashSource(t, src1Val), hashSource(t, src2Val); h1 != h2 {
		t.Errorf("value receiver: expected %s, got %s", h1, h2)
	}
}

func TestNamedReturnValuesAndNakedReturn(t *testing.T) {
	src1 := `package main
func divide(dividend, divisor int) (quotient int, remainder int) {
	quotient = dividend / divisor
	remainder = dividend % divisor
	return
}`
	src2 := `package main
func divide(a, b int) (q int, r int) {
	q = a / b
	r = a % b
	return
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("named returns and naked return: expected %s, got %s", h1, h2)
	}
}

func TestDeepNestedScopesAndMultiShadowing(t *testing.T) {
	src1 := `package main
func deep(x int) int {
	if x > 0 {
		x := x + 1
		for i := 0; i < x; i++ {
			x := i * 2
			if x > 5 {
				return x
			}
		}
		return x
	}
	return x
}`
	src2 := `package main
func deep(val int) int {
	if val > 0 {
		val := val + 1
		for idx := 0; idx < val; idx++ {
			val := idx * 2
			if val > 5 {
				return val
			}
		}
		return val
	}
	return val
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("deep nested shadowing: expected %s, got %s", h1, h2)
	}
}

func TestSliceExpressions(t *testing.T) {
	src1 := `package main
func sliceOps(s []int) ([]int, []int, []int, []int) {
	return s[1:2], s[1:], s[:3], s[1:2:3]
}`
	src2 := `package main
func sliceOps(arr []int) ([]int, []int, []int, []int) {
	return arr[1:2], arr[1:], arr[:3], arr[1:2:3]
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("slice expressions: expected %s, got %s", h1, h2)
	}
}

func TestCompositeLiterals(t *testing.T) {
	src1 := `package main
type Config struct { Host string; Port int }
func makeConfig(h string, p int) (Config, map[string]int, []int) {
	cfg := Config{Host: h, Port: p}
	m := map[string]int{h: p}
	s := []int{p, p + 1}
	return cfg, m, s
}`
	src2 := `package main
type Config struct { Host string; Port int }
func makeConfig(serverHost string, serverPort int) (Config, map[string]int, []int) {
	conf := Config{Host: serverHost, Port: serverPort}
	mp := map[string]int{serverHost: serverPort}
	slice := []int{serverPort, serverPort + 1}
	return conf, mp, slice
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("composite literals: expected %s, got %s", h1, h2)
	}
}

func TestLabelsAndBranching(t *testing.T) {
	src1 := `package main
func search(matrix [][]int, target int) bool {
Outer:
	for _, row := range matrix {
		for _, val := range row {
			if val == target {
				break Outer
			}
		}
	}
	return false
}`
	src2 := `package main
func search(grid [][]int, needle int) bool {
Outer:
	for _, line := range grid {
		for _, item := range line {
			if item == needle {
				break Outer
			}
		}
	}
	return false
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("labels and branching: expected %s, got %s", h1, h2)
	}
}

func TestDeferAndGoStatements(t *testing.T) {
	src1 := `package main
func concurrency(fn func(int), val int) {
	defer fn(val)
	go fn(val + 1)
}`
	src2 := `package main
func concurrency(callback func(int), arg int) {
	defer callback(arg)
	go callback(arg + 1)
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("defer and go: expected %s, got %s", h1, h2)
	}
}

func TestImmediatelyInvokedFuncLiteral(t *testing.T) {
	src1 := `package main
func compute(n int) int {
	return func(x int) int {
		return x * 2
	}(n)
}`
	src2 := `package main
func compute(val int) int {
	return func(item int) int {
		return item * 2
	}(val)
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("IIFE: expected %s, got %s", h1, h2)
	}
}

func TestTypeAssertionTwoValueAssignment(t *testing.T) {
	src1 := `package main
func convert(data any) (string, bool) {
	res, ok := data.(string)
	return res, ok
}`
	src2 := `package main
func convert(input any) (string, bool) {
	text, success := input.(string)
	return text, success
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("type assertion 2-value: expected %s, got %s", h1, h2)
	}
}

func TestSingleValueTypeAssertion(t *testing.T) {
	src1 := `package main
func asString(x any) string {
	return x.(string)
}`
	src2 := `package main
func asString(val any) string {
	return val.(string)
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("type assertion 1-value: expected %s, got %s", h1, h2)
	}
}

func TestChannelReceiveAssignment(t *testing.T) {
	src1 := `package main
func recv(ch chan int) (int, bool) {
	msg, ok := <-ch
	return msg, ok
}`
	src2 := `package main
func recv(queue chan int) (int, bool) {
	item, open := <-queue
	return item, open
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("channel receive assignment: expected %s, got %s", h1, h2)
	}
}

func TestMultipleVarGroupDeclarations(t *testing.T) {
	src1 := `package main
func setup(initVal int) int {
	var (
		first  = initVal + 1
		second = first * 2
	)
	return second
}`
	src2 := `package main
func setup(seed int) int {
	var (
		step1 = seed + 1
		step2 = step1 * 2
	)
	return step2
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("var group declaration: expected %s, got %s", h1, h2)
	}
}

func TestBitwiseAndUnaryOperators(t *testing.T) {
	src1 := `package main
func bitwise(mask, shift int) int {
	inv := ^mask
	shifted := (inv << shift) | (mask >> 1)
	cleared := shifted &^ mask
	return cleared
}`
	src2 := `package main
func bitwise(flags, offset int) int {
	neg := ^flags
	shifted := (neg << offset) | (flags >> 1)
	cleared := shifted &^ flags
	return cleared
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("bitwise and unary operators: expected %s, got %s", h1, h2)
	}
}

func TestEmptyInterfaceType(t *testing.T) {
	src1 := `package main
func process(v interface{}) bool {
	return v != nil
}`
	src2 := `package main
func process(item interface{}) bool {
	return item != nil
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("empty interface: expected %s, got %s", h1, h2)
	}
}

func TestLocalShadowingPredeclaredBuiltin(t *testing.T) {
	// A local variable named 'len' shadows the builtin len in Go
	src1 := `package main
func shadowBuiltin(x int) int {
	len := x * 10
	return len + 1
}`
	src2 := `package main
func shadowBuiltin(num int) int {
	len := num * 10
	return len + 1
}`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 != h2 {
		t.Errorf("shadowing builtin: expected %s, got %s", h1, h2)
	}
}

func TestHashFileOnDisk(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.go")
	src := `package disktest

func Double(x int) int {
	return x * 2
}
`
	if err := os.WriteFile(filePath, []byte(src), 0644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	res, err := coctyl.HashFile(filePath, coctyl.Options{})
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	if res.Path != filePath {
		t.Errorf("expected Path %s, got %s", filePath, res.Path)
	}
	if res.Hash == "" {
		t.Errorf("expected non-empty file hash")
	}
}

// ----------------------------------------------------------------------------
// Negative Integrity Tests: Files that MUST produce DIFFERENT fingerprints
// ----------------------------------------------------------------------------

func TestIntegrityDataflowDifference(t *testing.T) {
	// a + b vs a + a
	src1 := `package main
func calc(a, b int) int { return a + b }`
	src2 := `package main
func calc(a, b int) int { return a + a }`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 == h2 {
		t.Errorf("expected different hashes for 'a + b' vs 'a + a', got %s", h1)
	}
}

func TestIntegrityParameterOrder(t *testing.T) {
	// (a int, b string) vs (b string, a int)
	src1 := `package main
func log(a int, b string) string { return b }`
	src2 := `package main
func log(b string, a int) string { return b }`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 == h2 {
		t.Errorf("expected different hashes for swapped parameter types, got %s", h1)
	}
}

func TestIntegrityReturnOrder(t *testing.T) {
	// return a, b vs return b, a
	src1 := `package main
func swap(a, b int) (int, int) { return a, b }`
	src2 := `package main
func swap(a, b int) (int, int) { return b, a }`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 == h2 {
		t.Errorf("expected different hashes for swapped return expressions, got %s", h1)
	}
}

func TestIntegrityOperatorDifference(t *testing.T) {
	operators := []string{"+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">="}
	hashes := make(map[string]string)

	for _, op := range operators {
		src := `package main
func op(a, b int) any { return a ` + op + ` b }`
		h := hashSource(t, src)
		if existingOp, exists := hashes[h]; exists {
			t.Errorf("operator collision: '%s' produced same hash as '%s' (%s)", op, existingOp, h)
		}
		hashes[h] = op
	}
}

func TestIntegrityStructFieldDifference(t *testing.T) {
	src1 := `package main
type User struct { First, Last string }
func get(u User) string { return u.First }`
	src2 := `package main
type User struct { First, Last string }
func get(u User) string { return u.Last }`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 == h2 {
		t.Errorf("expected different hashes for u.First vs u.Last, got %s", h1)
	}
}

func TestIntegrityReceiverTypeDifference(t *testing.T) {
	src1 := `package main
type ServiceA struct{}
func (s *ServiceA) Do() int { return 1 }`
	src2 := `package main
type ServiceB struct{}
func (s *ServiceB) Do() int { return 1 }`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 == h2 {
		t.Errorf("expected different hashes for methods on different receiver types, got %s", h1)
	}
}

func TestIntegrityConstantLiteralDifference(t *testing.T) {
	src1 := `package main
func get() int { return 100 }`
	src2 := `package main
func get() int { return 200 }`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 == h2 {
		t.Errorf("expected different hashes for literals 100 vs 200, got %s", h1)
	}
}

func TestIntegrityBuiltinDifference(t *testing.T) {
	src1 := `package main
func check(s []int) int { return len(s) }`
	src2 := `package main
func check(s []int) int { return cap(s) }`
	if h1, h2 := hashSource(t, src1), hashSource(t, src2); h1 == h2 {
		t.Errorf("expected different hashes for builtins len vs cap, got %s", h1)
	}
}

func TestIntegrityControlFlowDifference(t *testing.T) {
	srcIf := `package main
func flow(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}`
	srcFor := `package main
func flow(x int) int {
	for x > 0 {
		return 1
	}
	return 0
}`
	if h1, h2 := hashSource(t, srcIf), hashSource(t, srcFor); h1 == h2 {
		t.Errorf("expected different hashes for if vs for control flow, got %s", h1)
	}
}

func TestIntegrityShadowedVsBuiltin(t *testing.T) {
	// One function shadows 'len' as a variable, the other uses builtin 'len'
	srcShadow := `package main
func testLen(items []int) int {
	len := 42
	return len
}`
	srcBuiltin := `package main
func testLen(items []int) int {
	return len(items)
}`
	if h1, h2 := hashSource(t, srcShadow), hashSource(t, srcBuiltin); h1 == h2 {
		t.Errorf("expected different hashes for shadowed variable vs builtin call")
	}
}

func TestIgnoreImportsOption(t *testing.T) {
	src1 := `package main
import "fmt"
func greet(name string) { fmt.Println(name) }`

	src2 := `package main
import (
	"fmt"
	"strings"
	"time"
)
func greet(name string) { fmt.Println(name) }`

	// 1. By default, imports ARE ignored, so file hashes are identical!
	res1Default, err := coctyl.HashSource([]byte(src1), coctyl.Options{})
	if err != nil {
		t.Fatal(err)
	}
	res2Default, err := coctyl.HashSource([]byte(src2), coctyl.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res1Default.Hash != res2Default.Hash {
		t.Errorf("expected identical file hashes with default options (imports ignored)")
	}

	// 2. When explicitly requiring imports (IncludeImports: true), file hashes MUST differ
	optsWithImports := coctyl.Options{IncludeImports: true}
	res1With, err := coctyl.HashSource([]byte(src1), optsWithImports)
	if err != nil {
		t.Fatal(err)
	}
	res2With, err := coctyl.HashSource([]byte(src2), optsWithImports)
	if err != nil {
		t.Fatal(err)
	}
	if res1With.Hash == res2With.Hash {
		t.Errorf("expected different file hashes when IncludeImports: true")
	}
}
