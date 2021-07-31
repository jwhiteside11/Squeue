# Squeue = Slice-based queue

Go does not have a built-in queue data type. Upon learning this, I took to the internet to view common solutions for this. I found that there were two that were widely used:
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
- **(queue Squeue) Each() []interface{}** - Returns a new slice containing only the underlying slice values; convinience method
- **(queue Squeue) String() string** - String representation of underlying slice values

## Performance

The slice-based queue implementation is generally more performant than a linked list-based queue, in both time and memory. The performance gains are had by decreasing memory complexity, and amortizing the time expense of memory allocation. The performance improves as the throughput of the queue grows.

For small throughput (n < 50), a linked list is generally faster. For large throughput (n > 1000), the squeue is generally faster (50% time req). For very large throughput (n > 10^6), the edge cases are few, and this implementation becomes highly performant. This can be verified using the included test file.

For data analysis or high-load applications, the time and memory requirements of this queue implementation are highly desirable.

Amortization is acheived by growing and shrinking the underlying slice when its capacity is reached. This behavior allows for quicker writes to the data structure, because generally we don't need to allocate anything in order to add an element; we simply assign it to a memory address. The linked list queue, on the other hand, must allocate a new node for each element as they are added.

The squeue is also a more lightweight data structure; the edges between elements in the linked list use **O(N)** memory, whereas the squeue's two-pointer technique uses **O(1)** memory. Even with the reallocations being made, the total memory allocated by the program using a squeue is generally less than that needed by a program using a linked list as a queue.

This module was originally built to hinder the memory leakage existent in a slice-based queue implementation. It accomplishes this by discarding unused values from the queue, and also by reallocating the underlying slice periodically as the queue gets used. The reallocation tells the GC that we are no longer using the slice's old underlying array, and discarding the value does much the same. This way, the queue will not leak memory over time.

## Contributions

This is an ongoing open-source project, open to any and all contributions. Any help is appreciated!
