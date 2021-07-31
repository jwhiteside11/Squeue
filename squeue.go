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
// Two pointers are used to track the index of first element and
// the first empty array slot. Reallocation occurs when the number
// of elements added exceeds the capacity of the underlying slice,
// or when less than half the slice's slots are being utilized.
//
// This implementation beats the built-in linked list as a queue,
// both in time and in memory.
// If you instead want to use a Stack, just use a slice outright;
// there exists no memory issue.
// If you want a double-ended queue (Deque), use a linked list.
// The container/list package provides a doubly-linked list,
// which performs better as a deque than an array.

type Squeue struct {
	squeue            []interface{} // slice of elements
	firstIdx, lastIdx int           // index of first element, first available slot
}

// Constructor - convinience method
func New(initial ...interface{}) Squeue {
	n, m := len(initial), len(initial)
	if m < 10 {
		m = 10
	}
	qq := make([]interface{}, 2*m)
	copy(qq, initial)
	return Squeue{qq, 0, n}
}

// Add element to queue (add last)
func (sq *Squeue) Enqueue(elem interface{}) {
	if sq.lastIdx == cap(sq.squeue) {
		sq.grow()
	}
	sq.squeue[sq.lastIdx] = elem
	sq.lastIdx++
}

// Returns first element in queue without removing it
func (sq *Squeue) Peek() (interface{}, error) {
	if sq.firstIdx == sq.lastIdx {
		return nil, fmt.Errorf("no elements remaining in queue")
	}
	return sq.squeue[sq.firstIdx], nil
}

// Remove element from queue (remove first)
func (sq *Squeue) Dequeue() (interface{}, error) {
	elem, err := sq.Peek()
	if err != nil {
		return nil, err
	}
	sq.squeue[sq.firstIdx] = nil
	sq.firstIdx++
	if (2 * sq.firstIdx) > len(sq.squeue) {
		sq.shrink()
	}
	return elem, nil

}

// Returns length of underlying slice
func (sq *Squeue) Size() int {
	return sq.lastIdx - sq.firstIdx
}

// Returns underlying slice for iteration - convinience method
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

// Invoked when backing array reaches capacity
func (sq *Squeue) grow() {
	n := cap(sq.squeue)
	if n < 15 {
		n = 15
	}
	m := 2 * n
	sq.resize(m)
}

// Invoked when half or more of backing array is no longer usable
func (sq *Squeue) shrink() {
	n := cap(sq.squeue)
	// For small n, place values at beginning of larger slice
	if n <= 30 {
		sq.resize(30)
		return
	}
	// Resize slice to 3/4 capacity
	m := (3 * n) / 4
	sq.resize(m)
}

// Allocates a new slice, copies elements to new slice starting at 0th index, and reconfigures pointers
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
