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
	outerQ                     []*[]interface{} // slice of elements
	head, tail                 []interface{}
	headF, headL, tailF, tailL int
	firstIdx, lastIdx, size    int // index of first element, first available slot
}

// Constructor - convinience method
// Accepts initial values to be enqueued, in the order listed
func New(initial ...interface{}) Squeue {
	n, m := len(initial), len(initial)
	if m < 10 {
		m = 10
	}
	qq := make([]interface{}, 2*m)
	copy(qq, initial)
	//inner := &InnerQueue{qq, 0, n}

	outer := make([]*[]interface{}, 10)
	outer[0] = &qq

	return Squeue{outer, qq, nil, 0, n, 0, 0, 0, 1, n}
}

// Add element to queue
// Adds element to tail, increments tail pointer
func (sq *Squeue) Enqueue(elem interface{}) {
	if sq.tail == nil {
		if sq.headL == sq.headF && sq.head[sq.headL] != nil {
			inner := make([]interface{}, 2*sq.Size())
			sq.outerQ[sq.lastIdx] = &inner
			// Set tail, pointers
			sq.tail = inner
			sq.tailF, sq.tailL = 0, 0
			sq.lastIdx = (sq.lastIdx + 1) % len(sq.outerQ)
		} else {
			sq.head[sq.headL] = elem
			sq.headL = (sq.headL + 1) % len(sq.head)
			return
		}
	} else {
		if sq.tailL == sq.tailF && sq.tail[sq.tailL] != nil {
			sq.size += len(sq.tail)
			inner := make([]interface{}, 2*sq.Size())
			sq.outerQ[sq.lastIdx] = &inner
			// Set tail, pointers
			sq.tail = inner
			sq.tailF, sq.tailL = 0, 0
			sq.lastIdx = (sq.lastIdx + 1) % len(sq.outerQ)
			if sq.lastIdx == sq.firstIdx && sq.outerQ[sq.lastIdx] != nil {
				sq.grow()
			}
		}
	}
	sq.tail[sq.tailL] = elem
	sq.tailL = (sq.tailL + 1) % len(sq.tail)
}

// Retrieve element from queue without removing it
// Checks for empty queue, if not returns first elem
func (sq *Squeue) Peek() (interface{}, error) {
	// Check for elem in head queue
	elem := sq.head[sq.headF]
	// If not there, move to next queue until only head left
	if elem == nil {
		disp := (sq.firstIdx + 1) % len(sq.outerQ)
		if sq.outerQ[disp] == nil {
			return nil, fmt.Errorf("no elements remaining in queue")
		}
		sq.outerQ[sq.firstIdx] = nil
		sq.firstIdx = disp
		sq.head = (*sq.outerQ[sq.firstIdx])
		if sq.outerQ[sq.firstIdx] == &sq.tail {
			sq.headF, sq.headL = sq.tailF, sq.tailL
			sq.tail = nil
		} else {
			sq.size -= len(sq.head)
			sq.headF, sq.headL = 0, 0
		}
		return sq.Peek()
	}
	return elem, nil
}

// Remove element from queue
// Retrieves the elem, and if successful deletes its value in the slice
// Increments the head pointer to next elem in queue
func (sq *Squeue) Dequeue() (interface{}, error) {
	elem, err := sq.Peek()
	if err != nil {
		return nil, err
	}
	// Void element, move head pointer
	sq.head[sq.headF] = nil
	sq.headF = (sq.headF + 1) % len(sq.head)
	return elem, nil
}

// Returns number of elements in queue
func (sq *Squeue) Size() int {
	total := sq.size
	if sq.head != nil {
		f1, f2 := sq.headF, sq.headL
		switch {
		case f1 < f2:
			total += f2 - f1
		case f1 > f2:
			total += len(sq.head) - f1 + f2
		default:
			if sq.head[f1] != nil {
				total += len(sq.head)
			}
		}
	}
	if sq.tail != nil {
		l1, l2 := sq.tailF, sq.tailL
		switch {
		case l1 < l2:
			total += l2 - l1
		case l1 > l2:
			total += len(sq.tail) - l1 + l2
		default:
			if sq.tail[l1] != nil {
				total += len(sq.tail)
			}
		}
	}
	return total
}

// Returns true if queue is empty
func (sq *Squeue) Empty() bool {
	return sq.Size() == 0
}

// Returns underlying slice for iteration - convinience method
func (sq *Squeue) Each() []interface{} {
	s := make([]interface{}, 0)
	for _, q := range sq.outerQ {
		s = append(s, *q...)
	}
	return s
}

// String representation: formats relevant slots of underlying slice as string
func (sq *Squeue) String() string {
	return fmt.Sprint(sq.Each())
}

// Resize slice to double the number of elements in the queue
// For small n, place values at beginning of larger slice to prevent unnecessary allocations
func (sq *Squeue) grow() {
	fmt.Println("GROW")
	n := cap(sq.outerQ)
	if n < 6 {
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
	//fmt.Println("RESIZE CALLED", len(sq.outerQ), m, sq.firstIdx, sq.lastIdx)
	qq := make([]*[]interface{}, m)
	// Copy existing values to beginning of new slice
	j := 0
	if sq.firstIdx == 0 {
		copy(qq, sq.outerQ)
		sq.outerQ = qq
		return
	}
	lenq := len(sq.outerQ)
	for i := sq.firstIdx; i < lenq; i++ {
		qq[j] = sq.outerQ[i]
		j++
	}
	for i := 0; i < sq.lastIdx; i++ {
		qq[j] = sq.outerQ[i]
		j++
	}
	// Set pointers
	sq.firstIdx = 0
	sq.lastIdx = j
	// Set underlying slice as newly allocated slice
	sq.outerQ = qq
}
