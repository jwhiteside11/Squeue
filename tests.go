package squeue

import (
	"container/list"
	"fmt"
	"math"
	"runtime"
	"time"
)

// Squeue test suites
//
// Run Example() to see the example output.
//
// Various methods are included for testing the squeue vs. a linked-list
// queue: CompareQueues, SQTest, and LLQTest. They report runtime stastics,
// and each take an optional scale parameter that sets the desired throughput
// of the test.
//
// It is demonstrable that the squeue does have edge cases, but the average
// case is much more favorable than a linked list queue.

var mem runtime.MemStats
var scale int = 10000

func Example() {
	queue := New()
	queue.Enqueue("Hello")
	queue.Enqueue(2)
	queue.Enqueue("The")
	queue.Enqueue("World")
	for _, v := range queue.Each() {
		fmt.Println(v)
	}
	for !queue.Empty() {
		queue.Dequeue()
	}

	_, err := queue.Dequeue()
	if err != nil {
		fmt.Println(err)
	}
}

// Test LinkedList queue
func CompareQueues(n ...int) {
	if len(n) > 0 {
		scale = n[0]
	}
	tL1, mL1 := linearLLQTest()
	fmt.Printf("LQ linear:  %vns, used ~%vKB\n", tL1, mL1/1000)
	tS1, mS1 := linearSQTest()
	fmt.Printf("SQ linear:  %vns, used ~%vKB\n", tS1, mS1/1000)
	tL2, mL2 := ladderLLQTest()
	fmt.Printf("LQ ladder:  %vns, used ~%vKB\n", tL2, mL2/1000)
	tS2, mS2 := ladderSQTest()
	fmt.Printf("SQ ladder:  %vns, used ~%vKB\n", tS2, mS2/1000)
	tL3, mL3 := pushPopLLQTest()
	fmt.Printf("LQ pushpop: %vns, used ~%vKB\n", tL3, mL3/1000)
	tS3, mS3 := pushPopSQTest()
	fmt.Printf("SQ pushpop: %vns, used ~%vKB\n", tS3, mS3/1000)
	tL4, mL4 := upDownLLQTest()
	fmt.Printf("LQ up-down: %vns, used ~%vKB\n", tL4, mL4/1000)
	tS4, mS4 := upDownSQTest()
	fmt.Printf("SQ up-down: %vns, used ~%vKB\n", tS4, mS4/1000)

	// Get average ratio between times/memory allocs, print result
	meanRatT := (float64(tS1)/float64(tL1) + float64(tS2)/float64(tL2) + float64(tS3)/float64(tL3) + float64(tS4)/float64(tL4)) / 4
	meanRatM := (float64(mS1)/float64(mL1) + float64(mS2)/float64(mL2) + float64(mS3)/float64(mL3) + float64(mS4)/float64(mL4)) / 4
	fmt.Printf("SliceQueue runs on average in %.2f%% time and %.2fs%% memory of a LinkedQueue\n\n", meanRatT*100, meanRatM*100)
}

func SQTest(m ...int) (int64, uint64) {
	if len(m) > 0 {
		scale = m[0]
	}
	startT, startM := runtimeStats()

	qq := New()
	// Linear
	for i := 0; i < scale; i++ {
		qq.Enqueue(i)
	}
	for i := 0; i < scale; i++ {
		qq.Dequeue()
	}

	// PushPop
	for i := 0; i < scale; i++ {
		qq.Enqueue(i)
		qq.Dequeue()
	}

	// UpDown
	n := scale / 10
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			qq.Enqueue(i)
		}
		for j := 0; j < 10; j++ {
			qq.Dequeue()
		}
	}

	// Ladder
	n = int(math.Sqrt(float64(scale)))
	for i := 0; i < n; i++ {
		for j := 0; j < n-i; j++ {
			qq.Enqueue(i)
		}
		for j := 0; j <= i; j++ {
			qq.Dequeue()
		}
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)
	//fmt.Printf("SliceQueue  - total time:  %vns, used ~%vKB\n", elapsed, used/1000)

	return elapsed, used
}

func LLQTest(m ...int) (int64, uint64) {
	if len(m) > 0 {
		scale = m[0]
	}
	startT, startM := runtimeStats()

	ll := list.New()
	// Linear
	for i := 0; i < scale; i++ {
		ll.PushBack(i)
	}
	for i := 0; i < scale; i++ {
		e := ll.Front() // First element
		ll.Remove(e)
	}

	// PushPop
	for i := 0; i < scale; i++ {
		ll.PushBack(i)
		e := ll.Front() // First element
		ll.Remove(e)
	}

	// UpDown
	n := scale / 10
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			ll.PushBack(i)
		}
		for j := 0; j < 10; j++ {
			e := ll.Front()
			ll.Remove(e)
		}
	}

	// Ladder
	n = int(math.Sqrt(float64(scale)))
	for i := 0; i < n; i++ {
		for j := 0; j < n-i; j++ {
			ll.PushBack(i)
		}
		for j := 0; j <= i; j++ {
			e := ll.Front() // First element
			ll.Remove(e)
		}
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)
	//fmt.Printf("LinkedQueue - total time:  %vns, used ~%vKB\n", elapsed, used/1000)

	return elapsed, used
}

func upDownSQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	qq := New()
	n := scale / 10
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			qq.Enqueue(i)
		}
		for j := 0; j < 10; j++ {
			qq.Dequeue()
		}
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)

	return elapsed, used
}

func upDownLLQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	ll := list.New()
	n := scale / 10
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			ll.PushBack(i)
		}
		for j := 0; j < 10; j++ {
			e := ll.Front()
			ll.Remove(e)
		}
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)

	return elapsed, used
}

func ladderSQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	qq := New()
	n := int(math.Sqrt(float64(scale)))
	for i := 0; i < n; i++ {
		for j := 0; j < n-i; j++ {
			qq.Enqueue(i)
		}
		for j := 0; j <= i; j++ {
			qq.Dequeue()
		}
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)

	return elapsed, used
}

func ladderLLQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	ll := list.New()
	n := int(math.Sqrt(float64(scale)))
	for i := 0; i < n; i++ {
		for j := 0; j < n-i; j++ {
			ll.PushBack(i)
		}
		for j := 0; j <= i; j++ {
			e := ll.Front() // First element
			ll.Remove(e)
		}
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)

	return elapsed, used
}

func pushPopSQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	qq := New()
	for i := 0; i < scale; i++ {
		qq.Enqueue(i)
		qq.Dequeue()
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)

	return elapsed, used
}

func pushPopLLQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	ll := list.New()
	for i := 0; i < scale; i++ {
		ll.PushBack(i)
		e := ll.Front() // First element
		ll.Remove(e)
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)

	return elapsed, used
}

func linearSQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	qq := New()
	for i := 0; i < scale; i++ {
		qq.Enqueue(i)
	}
	for i := 0; i < scale; i++ {
		qq.Dequeue()
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, uint64(endM-startM)

	return elapsed, used
}

func linearLLQTest() (int64, uint64) {
	startT, startM := runtimeStats()

	ll := list.New()
	for i := 0; i < scale; i++ {
		ll.PushBack(i)
	}
	for i := 0; i < scale; i++ {
		e := ll.Front() // First element
		ll.Remove(e)
	}

	endT, endM := runtimeStats()

	elapsed, used := endT-startT, endM-startM

	return elapsed, used
}

func runtimeStats() (int64, uint64) {
	runtime.ReadMemStats(&mem)
	return time.Now().UnixNano(), mem.TotalAlloc
}
