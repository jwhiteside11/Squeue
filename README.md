# Squeue | Slice-based queue

Go does not have a built-in queue data type. I took to the internet to view common solutions for this, and found that there were two that were widely used:
- 1) Use a slice; the append method handles adding to the queue, removing is done with the slicing syntax. This works well, but the internals give rise to memory concerns.
- 2) Use a linked list; Go provides a built-in container/list module with a doubly-linked list implementation. This is a great thing, but it is more information than we need to implement a queue. Therefore, it's going to be underoptimized.

This module provides a slice-based queue implementation, hereafter referred to as a squeue, and several common methods for the queue data type.

## Getting Started

You can install this package in your project by running:

```bash
go get github.com/jwhiteside11/squeue
```

and then use it in your project like so:

```go
import "github.com/jwhiteside11/squeue"

func someFunctionName() {
	queue := squeue.New()

	queue.Enqueue("Hello")
	queue.Enqueue(2)
	queue.Enqueue("The")
	queue.Enqueue("World")

	fmt.Println(queue.Peek())
	
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
```

## Available methods

- **New(elems ...interface{}) Squeue** - Create a queue
- **(queue Squeue) Enqueue(elem interface{})** - Add element to back of queue
- **(queue Squeue) Peek () (interface{}, error)** - Retrieve, but do not remove, element from head of queue
- **(queue Squeue) Dequeue() (interface{}, error)** - Remove element from head of queue
- **(queue Squeue) Size() int** - Get size of queue
- **(queue Squeue) Empty() bool** - Returns true if queue is empty
- **(queue Squeue) Each() []interface{}** - Returns a new slice pointing containing only the underlying slice values; convinience method
- **(queue Squeue) String() string** - String representation of underlying slice values

## Performance

The slice-based queue implementation is more performant than a linked list queue, in both time and memory. The memory requirement is similar between the two, but the squeue is generally over twice as fast. That said, slices are a linear data structure, so we must consider certain memory implications.

The squeue is faster because it amortizes the time cost of allocation by growing and shrinking the underlying slice as needed. This behavior allows for quicker writes to the data structure, because we don't need to allocate anything in order to add an element; we simply assign it a memory address. The linked list queue must spend time allocating an element each time one is added.

The concerns regarding memory in a slice-based queue implementation have been addressed in this module. Unused values are discarded, and the underlying slice reallocates periodically as the queue gets used. The reallocation tells the GC that we are no longer using the slice's old underlying array. This way, the queue will not leak memory over time.

## Contributions

I'm open to any and all contributions. Make a PR.
