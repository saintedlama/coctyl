package validation_tests

import (
	"testing"
)

func TestCorpusBinarySearch(t *testing.T) {
	baseline := `package search

func BinarySearch(arr []int, target int) int {
	low := 0
	high := len(arr) - 1

	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}`

	refactoredEquivalent := `package algorithms

// FindIndex performs a binary search on sorted slice
func FindIndex(haystack []int, needle int) int {
	left := 0
	right := len(haystack) - 1

	for left <= right {
		middle := left + (right-left)/2
		if haystack[middle] == needle {
			return middle
		} else if haystack[middle] < needle {
			left = middle + 1
		} else {
			right = middle - 1
		}
	}
	return -1
}`

	mutatedBug := `package search

func BinarySearch(arr []int, target int) int {
	low := 0
	high := len(arr) - 1

	for low < high { // Mutated: strict inequality causes missed element
		mid := low + (high-low)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}`

	assertEquivalent(t, "BinarySearchRefactoringEquivalence", baseline, refactoredEquivalent)
	assertDifferent(t, "BinarySearchMutationDetection", baseline, mutatedBug)
}

func TestCorpusRingBuffer(t *testing.T) {
	baseline := `package queue

type RingBuffer struct {
	data     []int
	head     int
	tail     int
	capacity int
}

func (rb *RingBuffer) Push(val int) bool {
	next := (rb.tail + 1) % rb.capacity
	if next == rb.head {
		return false
	}
	rb.data[rb.tail] = val
	rb.tail = next
	return true
}

func (rb *RingBuffer) Pop() (int, bool) {
	if rb.head == rb.tail {
		return 0, false
	}
	val := rb.data[rb.head]
	rb.head = (rb.head + 1) % rb.capacity
	return val, true
}`

	refactoredEquivalent := `package collections

type RingBuffer struct {
	data     []int
	head     int
	tail     int
	capacity int
}

func (self *RingBuffer) Push(item int) bool {
	nextPos := (self.tail + 1) % self.capacity
	if nextPos == self.head {
		return false
	}
	self.data[self.tail] = item
	self.tail = nextPos
	return true
}

func (self *RingBuffer) Pop() (int, bool) {
	if self.head == self.tail {
		return 0, false
	}
	item := self.data[self.head]
	self.head = (self.head + 1) % self.capacity
	return item, true
}`

	mutatedBug := `package queue

type RingBuffer struct {
	data     []int
	head     int
	tail     int
	capacity int
}

func (rb *RingBuffer) Push(val int) bool {
	next := (rb.tail + 1) % rb.capacity
	if next == rb.head {
		return false
	}
	rb.data[rb.tail] = val
	rb.tail = next
	return true
}

func (rb *RingBuffer) Pop() (int, bool) {
	if rb.head == rb.tail {
		return 0, false
	}
	val := rb.data[rb.head]
	rb.head = (rb.head + 2) % rb.capacity // Mutated: step by 2 skips entries
	return val, true
}`

	assertEquivalent(t, "RingBufferRefactoringEquivalence", baseline, refactoredEquivalent)
	assertDifferent(t, "RingBufferMutationDetection", baseline, mutatedBug)
}

func TestCorpusWorkerPool(t *testing.T) {
	baseline := `package concurrency

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		res := j * 2
		results <- res
	}
}

func RunPool(numWorkers int, numJobs int) {
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
}`

	refactoredEquivalent := `package threading

func executeWorker(workerID int, tasks <-chan int, outputs chan<- int) {
	for task := range tasks {
		output := task * 2
		outputs <- output
	}
}

func RunPool(workerCount int, taskCount int) {
	taskQueue := make(chan int, taskCount)
	outputQueue := make(chan int, taskCount)

	for idx := 1; idx <= workerCount; idx++ {
		go executeWorker(idx, taskQueue, outputQueue)
	}

	for current := 1; current <= taskCount; current++ {
		taskQueue <- current
	}
	close(taskQueue)

	for received := 1; received <= taskCount; received++ {
		<-outputQueue
	}
}`

	mutatedBug := `package concurrency

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		res := j * 2
		results <- res
	}
}

func RunPool(numWorkers int, numJobs int) {
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= numWorkers; w++ {
		worker(w, jobs, results) // Mutated: synchronous worker blocks pool
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
}`

	assertEquivalent(t, "WorkerPoolRefactoringEquivalence", baseline, refactoredEquivalent)
	assertDifferent(t, "WorkerPoolMutationDetection", baseline, mutatedBug)
}

func TestCorpusTreeTraversal(t *testing.T) {
	baseline := `package tree

type Node struct {
	Value int
	Left  *Node
	Right *Node
}

func SumNodes(root *Node) int {
	if root == nil {
		return 0
	}
	leftSum := SumNodes(root.Left)
	rightSum := SumNodes(root.Right)
	return root.Value + leftSum + rightSum
}`

	refactoredEquivalent := `package tree

type Node struct {
	Value int
	Left  *Node
	Right *Node
}

func Total(n *Node) int {
	if n == nil {
		return 0
	}
	l := Total(n.Left)
	r := Total(n.Right)
	return n.Value + l + r
}`

	mutatedBug := `package tree

type Node struct {
	Value int
	Left  *Node
	Right *Node
}

func SumNodes(root *Node) int {
	if root == nil {
		return 0
	}
	leftSum := SumNodes(root.Left)
	rightSum := SumNodes(root.Left) // Mutated: recurses on Left twice
	return root.Value + leftSum + rightSum
}`

	assertEquivalent(t, "TreeTraversalRefactoringEquivalence", baseline, refactoredEquivalent)
	assertDifferent(t, "TreeTraversalMutationDetection", baseline, mutatedBug)
}

func TestCorpusMiddlewareChain(t *testing.T) {
	baseline := `package middleware

type Request struct {
	Path  string
	Token string
}

type Response struct {
	StatusCode int
	Body       string
}

type HandlerFunc func(Request) Response

func AuthMiddleware(next HandlerFunc) HandlerFunc {
	return func(req Request) Response {
		if req.Token == "" {
			return Response{StatusCode: 401, Body: "Unauthorized"}
		}
		return next(req)
	}
}`

	refactoredEquivalent := `package httpware

type Request struct {
	Path  string
	Token string
}

type Response struct {
	StatusCode int
	Body       string
}

type HandlerFunc func(Request) Response

func Authenticate(downstream HandlerFunc) HandlerFunc {
	return func(r Request) Response {
		if r.Token == "" {
			return Response{StatusCode: 401, Body: "Unauthorized"}
		}
		return downstream(r)
	}
}`

	mutatedBug := `package middleware

type Request struct {
	Path  string
	Token string
}

type Response struct {
	StatusCode int
	Body       string
}

type HandlerFunc func(Request) Response

func AuthMiddleware(next HandlerFunc) HandlerFunc {
	return func(req Request) Response {
		if req.Token != "" { // Mutated: inverted condition checks non-empty token
			return Response{StatusCode: 401, Body: "Unauthorized"}
		}
		return next(req)
	}
}`

	assertEquivalent(t, "MiddlewareChainRefactoringEquivalence", baseline, refactoredEquivalent)
	assertDifferent(t, "MiddlewareChainMutationDetection", baseline, mutatedBug)
}
