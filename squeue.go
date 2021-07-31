package squeue

/*
Copyright 2021 John D Whiteside

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissionsand limitations under the License.
*/

import (
	"fmt"
)

// Squeue - simple queue data structure built with slice internals
//
// Memory concerns involving slice-based queues have been handled:
// 	- Values are set to nil once no longer needed
//  - Queue lifecycle guarantees that the underlying slice will be
//	  reallocated as elements are added and removed, preventing
// 	  memory leaks
//
// Two pointers are used to track the index of first element and the
// first empty array slot. The slice and pointers reset when the end
// pointer goes beyond the capacity of the underlying slice. The slice
// is reallocated if it is under/over-utilizing its capacity.
//
// If you instead want to implement a Stack, just use a slice outright;
// there exists no memory issue.
//
// If you want a double-ended queue (Deque), use a linked list.
// The container/list package provides a built-in doubly-linked list;
// this will perform better as a deque than an array.

// Squeue type
// Uses an underlying slice and two pointers to emulate a queue
type Squeue struct {
	squeue            []interface{} // slice of elements
	firstIdx, lastIdx int           // index of first element, first available slot
}

// Constructor - convinience method
// Accepts initial values to be enqueued, in the order listed
func New(initial ...interface{}) Squeue {
	n, m := len(initial), len(initial)
	if m < 5 {
		m = 5
	}
	qq := make([]interface{}, 2*m)
	copy(qq, initial)
	return Squeue{qq, 0, n}
}

// Add element to queue
// Adds element to tail, increments tail pointer
func (sq *Squeue) Enqueue(elem interface{}) {
	if sq.lastIdx == cap(sq.squeue) {
		sq.refresh()
	}
	sq.squeue[sq.lastIdx] = elem
	sq.lastIdx++
}

// Retrieve element from queue without removing it
// Checks for empty queue, if not returns first elem
func (sq *Squeue) Peek() (interface{}, error) {
	// If pointers are the same, the queue is empty; return nil, error
	if sq.firstIdx == sq.lastIdx {
		return nil, fmt.Errorf("no elements remaining in queue")
	}
	return sq.squeue[sq.firstIdx], nil
}

// Remove element from queue
// Retrieves the elem, and if successful deletes its value in the slice
// Increments the head pointer to next elem in queue
func (sq *Squeue) Dequeue() (interface{}, error) {
	elem, err := sq.Peek()
	if err != nil {
		return nil, err
	}
	sq.squeue[sq.firstIdx] = nil
	sq.firstIdx++
	return elem, nil

}

// Returns number of elements in queue
func (sq *Squeue) Size() int {
	return sq.lastIdx - sq.firstIdx
}

// Returns true if queue is empty
func (sq *Squeue) Empty() bool {
	return sq.Size() == 0
}

// Returns underlying slice for iteration - convinience method
func (sq *Squeue) Each() []interface{} {
	return sq.squeue[sq.firstIdx:sq.lastIdx]
}

// String representation: formats relevant slots of underlying slice as string
func (sq *Squeue) String() string {
	return fmt.Sprint(sq.Each())
}

// Rewrite slice based on current array utilization
func (sq *Squeue) refresh() {
	lenq, capq := sq.Size(), cap(sq.squeue)
	// Three cases: stack too big, stack ok, stack too small
	switch {
	// Stack big for underlying slice
	case (3 * lenq) > (2 * capq):
		sq.grow()
	// Stack ok for underlying slice
	case (4 * lenq) > capq:
		sq.reset()
	// Stack small for underlying slice
	default:
		sq.shrink()
	}
}

// Uses the existing slice
// Swaps elements to the beginning of the slice
// Reconfigures pointers to reflect shift
func (sq *Squeue) reset() {
	// Use existing slice
	qq := sq.squeue
	// Copy existing values to beginning of new slice
	j := 0
	for i := sq.firstIdx; i < sq.lastIdx; i++ {
		qq[j], qq[i] = qq[i], qq[j]
		j++
	}
	// Set pointers
	sq.firstIdx = 0
	sq.lastIdx = j
}

// Resize slice to double the number of elements in the queue
func (sq *Squeue) grow() {
	n := sq.Size()
	sq.resize(2 * n)
}

// Resize slice to double the number of elements in the queue
// For small n, place values at beginning of larger slice to prevent unnecessary allocations
func (sq *Squeue) shrink() {
	n := sq.Size()
	if n < 5 {
		sq.resize(8)
	} else {
		sq.resize(2 * n)
	}
}

// Allocates a new slice
// Copies elements from queue into new slice starting at 0th index
// Reconfigures pointers to reflect shift
func (sq *Squeue) resize(m int) {
	// Allocate new slice
	qq := make([]interface{}, m)
	// Copy existing values to beginning of new slice
	j := 0
	for i := sq.firstIdx; i < sq.lastIdx; i++ {
		qq[j] = sq.squeue[i]
		j++
	}
	// Set pointers
	sq.firstIdx = 0
	sq.lastIdx = j
	// Set underlying slice as newly allocated slice
	sq.squeue = qq
}
