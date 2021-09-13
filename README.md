# Squeue = Slice-based queue

Go does not have a built-in queue data type. Upon learning this, I took to the internet to view common solutions for this. I found that there were two that were widely used:
- 1) Use a slice; the append method handles adding to the queue, removing is done with the slicing syntax. This works well, but the internals give rise to memory concerns.
- 2) Use a linked list; Go provides a built-in module with a doubly-linked list implementation. This is an fine answer, but a bit naive for my taste. 

I began creating a slice-backed queue data type in Go, and what I ended up with was a unique data structure. Backed by slices, it uses circular queue behaviors with a nested array structure to outperform other queue implementation in both time and memory. I would love more feedback on its performance vs. other queues available in Go and other languages.

This module offers a Go implementation of that data structure, along with several common methods for the queue data type. Enjoy.

## Getting Started

You can install this package in your project by running:

```bash
go get github.com/jwhiteside11/squeue
```

and then use it in your project like so:

```go
import "github.com/jwhiteside11/squeue"

func someFunction() {
	queue := squeue.New()

    // Use as queue
	queue.Push("Hello")
	queue.Push("World")
    el, _ := queue.Unshift()
    fmt.Println(el) // el == "Hello"

    // Use as deque
	queue.Shift("Hello")
    queue.Push("Welcome!")
    queue.Push(2)
    el, _ := queue.Pop()
    fmt.Println(el) // el == 2


    // Get element w/o removal
	queue.Shift("First")
	queue.Push("Last")
	fmt.Println(queue.PeekFront())
	fmt.Println(queue.PeekBack())

    // Iterate through elements in queue until queue is empty
	for !queue.Empty() {
		el, _ := queue.Unshift()
        fmt.Println(el)
	}

	// Iterate (slower than the above dequeueing pattern)
	for _, v := range queue.Each() {
        // do something with elem
	}

    // Catch error from delete operation (Unshift/Pop)
	_, err := queue.Pop()
	if err != nil {
		log.Fatal(err)
	}
}
```

## Available methods

- **New(elems ...interface{}) Squeue** - Create a new double-ended queue
- **(queue Squeue) Push(elem interface{})** - Add element to back of queue (enqueue)
- **(queue Squeue) Pop() (interface{}, error)** - Remove the last element from the queue
- **(queue Squeue) Shift(elem interface{})** - Add element to front of queue
- **(queue Squeue) Unshift() (interface{}, error)** - Remove the first element from the queue (dequeue)
- **(queue Squeue) PeekFront() (interface{}, error)** - Retrieve, but do not remove, the first element of the queue
- **(queue Squeue) PeekBack() (interface{}, error)** - Retrieve, but do not remove, the last element of the queue
- **(queue Squeue) Size() int** - Get size of queue
- **(queue Squeue) Empty() bool** - Returns true if queue is empty
- **(queue Squeue) Each() []interface{}** - Returns a new slice containing elements in queue order; convinience method
- **(queue Squeue) String() string** - String representation of queue

## Performance

This queue implementation is generally more performant than a linked list-based queue and a common circular array queue, in both time and memory. The performance improves as the throughput of the queue grows.

For small throughput (n < 100), a linked list is generally faster. This implementation seeks to eliminate the edge cases present in more common circular queue implementations, so for n > 100, there are very few cases where it is beaten by a linked list in any capacity. For very large throughput (n > 10^6), it's not even close. This can be verified using the included test file.

This module was originally built to hinder the memory leakage existent in a slice-based queue implementation. It accomplishes this by discarding unused values from the queue, and also by reallocating the underlying slice periodically as the queue gets used. The reallocation tells the GC that we are no longer using the slice's old underlying array, and discarding the value does much the same. This way, the queue will not leak memory over time. Analysis of runtime memory of long-standing use-cases confirm this.

## Contributions

This is an ongoing open-source project, open to any and all contributions. Any help is appreciated!
