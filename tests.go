package squeue

import (
	"container/list"
	"fmt"
	"log"
	"math"
	"runtime"
	"time"
)

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

func CompareQueues(n ...int) {
	// Test LinkedList queue
	if len(n) > 0 {
		scale = n[0]
	}
	tS1, mS1 := linearSQTest()
	tL1, mL1 := linearLLQTest()
	tS2, mS2 := ladderSQTest()
	tL2, mL2 := ladderLLQTest()
	tS3, mS3 := pushPopSQTest()
	tL3, mL3 := pushPopLLQTest()

	// Get average difference, std. deviation, print
	meanRatT := (float64(tS1)/float64(tL1) + float64(tS2)/float64(tL2) + float64(tS3)/float64(tL3)) / 3
	meanRatM := (float64(mS1)/float64(mL1) + float64(mS2)/float64(mL2) + float64(mS3)/float64(mL3)) / 3
	fmt.Printf("SliceQueue runs on average in %.2f%% time and %.2fs%% memory of a LinkedQueue\n", meanRatT, meanRatM)
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
	fmt.Printf("SliceQueue: %vns, used ~%vKB\n", elapsed, used/1000)

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
	fmt.Printf("SliceQueue: %vns, used ~%vKB\n", elapsed, used/1000)

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
	fmt.Printf("SliceQueue: %vns, used ~KB%v\n", elapsed, used/1000)

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
	fmt.Printf("SliceQueue: %vns, used ~KB%v\n", elapsed, used/1000)

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
	fmt.Printf("SliceQueue: %vns, used ~%vKB\n", elapsed, used/1000)

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
	fmt.Printf("LinkedQueue: %vns, used ~%vKB\n", elapsed, used/1000)

	return elapsed, used
}

func runtimeStats() (int64, uint64) {
	runtime.ReadMemStats(&mem)
	return time.Now().UnixNano(), mem.TotalAlloc
}

// Disregard

func testQueues() {
	e1, u1 := testArrayQueue()
	e2, u2 := testArrayQueue()
	fmt.Printf("Squeue used ~%.2f%% time and ~%.2f%% memory of LinkedList queue\n", float64(e1)/float64(e2)*100, float64(u1)/float64(u2)*100)
}

func testArrayQueue() (int64, uint64) {
	// Test ArrayQueue
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	start1T := time.Now().UnixNano()
	start1M := mem.TotalAlloc

	qq := New()
	for i := 0; i < 10000; i++ {
		qq.Enqueue(i)
	}
	for i := 0; i < 10000; i++ {
		qq.Dequeue()
	}

	for i := 0; i < 1000; i++ {

		for j := 0; j < 1000-i; j++ {
			qq.Enqueue(i)
		}
		for j := 0; j <= i; j++ {
			_, err := qq.Dequeue()
			if err != nil {
				fmt.Println(qq)
				log.Fatal(err)
			}
		}
	}

	runtime.ReadMemStats(&mem)
	end1T := time.Now().UnixNano()
	end1M := mem.TotalAlloc

	elapsed, used := end1T-start1T, uint64(end1M-start1M)
	fmt.Printf("SliceQueue:  %vns, used ~%v\n", elapsed, used/1000)

	return elapsed, used
}
func testLinkedDeque() (int64, uint64) {
	// Test LinkedList queue
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	start2T := time.Now().UnixNano()
	start2M := mem.TotalAlloc

	ll := list.New()

	for i := 0; i < 10000; i++ {
		ll.PushBack(i)
		ll.PushFront(i)
	}
	for i := 0; i < 10000; i++ {
		e := ll.Front() // First element
		ll.Remove(e)
		f := ll.Back() // First element
		ll.Remove(f)
	}
	for i := 0; i < 1000; i++ {
		for j := 0; j < 1000-i; j++ {
			ll.PushFront(i)
		}
		for j := 0; j <= i; j++ {
			e := ll.Back() // First element
			ll.Remove(e)
		}
	}

	runtime.ReadMemStats(&mem)
	end2T := time.Now().UnixNano()
	end2M := mem.TotalAlloc

	elapsed, used := end2T-start2T, uint64(end2M-start2M)
	fmt.Printf("LinkedDeque: %vns, used ~%v\n", elapsed, used/1000)

	return elapsed, used
}
