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

// Squeue - Performant double-ended queue data structure
// 			Built in Go with slice internals
//
// Memory concerns involving slice-based queues have been handled:
//  - Values are set to nil once no longer needed
//  - Both the outer and underlying slices behave circularly, so
//    they both only need grow at full slice utilization.
//  - When an inner queue reaches capacity, it creates a new
//    slice to add/take elements to/from, with double the capacity
//    of the previous inner queue. The previous inner queue remains
//    untouched until its values are needed, it is not reallocated.
//  - When the outer queue reaches capacity and must grow, the
//    pointers in the outer queue are reallocated in proper order
//    into a bigger slice. The values at the pointers are untouched.
//  - At no point does the program allow for growth of slices beyond
//    what is needed, and the memory provided by the slices is
//    utilized maximally as elements are added and removed (due to
//    circular behavior). Dynamic growth/allocations and optimal
//    utilization of the slices prevents any kind of memory leak.
//
// Time amortization is prevalent in this implementation:
//  - The slice, by its nature, allocates a set number of memory
//    addresses to which we can add data. This behavior is faster at
//    scale, because we do not need to allocate memory for each node
//    upon element add, like in a linked list.
//  - This data structure allows for queue slices to be available at
//    the top level, while the data that lies in the middle of the
//    deque is stored as pointers and dereferenced when the data is
//    needed. This way, the add/delete operations must only index one
//    slice instead of two (outer and inner to get elem).
//  - The size is also found in a O(1) time amortized fashion, preventing
//    the need for an additional CPU operation at every add/delete.
//
// Data Structure Overview
//  - Consists of head/tail slices with elements, as well as an
//    underlying slice of memory references which emulate a queue.
//  - Add/delete operations are performed on the head/tail slices.
//    When a slice meets its capacity and it is no longer immediately
//    needed, its start index is cached, and a new slice is allocated.
//    A pointer to the newly allocated slice is saved in the outer slice,
//    indexed adjacent to the pointer of the slice that met its capacity.
//    The newly allocated slice becomes the new head/tail.
//  - When the head/tail empties, a pointer is taken from the outer slice.
//    The pointer is dereferenced, and used as the new head/tail.
//  - When only one inner slice reference remains, the tail is set to nil,
//    and all operations are done on the head slice. A head slice
//    containing no elements would indicate an empty queue.
//  - The inner and outer slices all behave as circular queues, for time
//    and memory optimization.
//  - The outer queue keeps a buffer slice outside the head and tail; the
//    buffer is not there from initialization, but as the queue grows and
//    shrinks, a buffer slice will be kept to avoid discarding and
//    reallocating the same slice redundantly during edge cases.

/* Data Types */

// Squeue: main type
// - Slice-based, circular queue that uses a cache to cut the time of necessary reallocations
type Squeue struct {
	head, tail                                 []interface{} // Head and tail slices, containing elements; add/delete operations of elements will occur on these slices
	cache                                      []*Cached     // Cache of pointers to slices and the index of their first element
	headF, headL, tailF, tailL, cacheF, cacheL int           // Circular pointers; F is the index to first element in queue, L is the index after the last element in queue (first available slot)
	cacheSize                                  int           // Size of cache; element counts recorded as slices enter the cache (time amortized)
}

// Cached: underlying type for Squeue
type Cached struct {
	ptr *[]interface{} // Pointer to a queue that exists in the Squeue data structure
	idx int            // Index of the first element in the queue that is pointed at (queue is circular)
}

/* Exports */

// New - queue constructor, convinience method
// Accepts initial values to be enqueued, in the order listed
func New(initial ...interface{}) Squeue {
	n := len(initial)
	head, cache := make([]interface{}, 2*max(n, 10)), make([]*Cached, 6)

	copy(head, initial)
	cache[0] = &Cached{&head, 0}

	return Squeue{head, nil, cache, 0, n, 0, 0, 0, 1, 0}
}

// Shift - add to front of queue
// Add element to the head, increments head pointer
func (sq *Squeue) Shift(elem interface{}) {
	// Check if head queue has room available
	if !(sq.headL == sq.headF && sq.head[sq.headL] != nil) {
		// Slots remain in head slice; set head pointer to next available, add elem
		sq.headF -= 1
		if sq.headF < 0 {
			sq.headF += len(sq.head)
		}
		sq.head[sq.headF] = elem
		return
	}
	// Head slice is full
	switch {
	case sq.tail == nil:
		// If tail is nil, full head becomes tail, so head can become new slice
		sq.tail = sq.head
		sq.tailF, sq.tailL = sq.headF, sq.headL
	default:
		// If cache at capacity, reallocate to bigger slice
		if sq.cacheF == sq.cacheL {
			sq.grow()
		}
		// Head needs to be cached, set starting elem before caching
		sq.cache[sq.cacheF].idx = sq.headF
		sq.cacheSize += len(sq.head)
	}
	// Set new index
	sq.cacheF -= 1
	if sq.cacheF < 0 {
		sq.cacheF += len(sq.cache)
	}
	// Check for slice in cache, set head
	if sq.cache[sq.cacheF] != nil {
		// Use empty slice from cache
		sq.head = (*sq.cache[sq.cacheF].ptr)
		sq.headF, sq.headL = 0, 0
	} else {
		// Create new head slice, save pointer to cache
		inner := make([]interface{}, min(2*max(len(sq.head), len(sq.tail)), 100000))
		sq.cache[sq.cacheF] = &Cached{&inner, 0}
		// Set head, pointers
		sq.head = inner
		sq.headF, sq.headL = 0, 1
	}
	// Add elem to head
	sq.head[sq.headF] = elem
}

// Push - add to back of queue (enqueue)
// Adds element to tail, increments tail pointer
func (sq *Squeue) Push(elem interface{}) {
	switch {
	case sq.tail == nil:
		// Perform operation on head slice
		if !(sq.headL == sq.headF && sq.head[sq.headL] != nil) {
			// Slots remain in head slice; add elem, inc tail pointer
			sq.head[sq.headL] = elem
			sq.headL = (sq.headL + 1) % len(sq.head)
			return
		} else {
			// Head full, check for slice in cache
			if sq.cache[sq.cacheL] != nil {
				// Use empty slice from cache
				sq.tail = (*sq.cache[sq.cacheL].ptr)
				sq.tailF, sq.tailL = 0, 0
			} else {
				// New slice allocated
				inner := make([]interface{}, min(2*len(sq.head), 100000))
				sq.cache[sq.cacheL] = &Cached{&inner, 0}
				// Set tail, pointers
				sq.tail = inner
				sq.tailF, sq.tailL = 0, 0
			}
			// Inc outer tail pointer
			sq.cacheL = (sq.cacheL + 1) % len(sq.cache)
		}
	default:
		// Tail slice exists
		if sq.tailL == sq.tailF && sq.tail[sq.tailL] != nil {
			// Tail at capacity
			if sq.cacheL == sq.cacheF {
				// Cache at capacity, grow outer slice
				sq.grow()
			}
			d1 := sq.cacheL - 1
			if d1 < 0 {
				d1 += len(sq.cache)
			}
			// Add tail pointer to cache, record size
			sq.cache[d1].idx = sq.tailF
			sq.cacheSize += len(sq.tail)
			// Check for slice in cache, set tail
			if sq.cache[sq.cacheL] != nil {
				// Use empty slice from cache
				sq.tail = (*sq.cache[sq.cacheL].ptr)
				sq.tailF, sq.tailL = 0, 0
			} else {
				// Create new tail
				inner := make([]interface{}, min(2*max(len(sq.head), len(sq.tail)), 100000))
				sq.cache[sq.cacheL] = &Cached{&inner, 0}
				// Set tail, pointers
				sq.tail = inner
				sq.tailF, sq.tailL = 0, 0
			}
			// Inc outer tail pointer
			sq.cacheL = (sq.cacheL + 1) % len(sq.cache)
		}
	}
	// Add elem to tail, inc tail pointer
	sq.tail[sq.tailL] = elem
	sq.tailL = (sq.tailL + 1) % len(sq.tail)
}

// PeekFront - retrieve first element from queue without removing it
// Checks for empty queue, if not returns first elem
func (sq *Squeue) PeekFront() (interface{}, error) {
	// Check for elem in head queue
	elem := sq.head[sq.headF]
	// Return elem if found
	if elem != nil {
		return elem, nil
	}
	// No elems in head; move to next slice in cache until only head left
	d1, d2 := (sq.cacheF - 1), ((sq.cacheF + 1) % len(sq.cache))
	if d1 < 0 {
		d1 += len(sq.cache)
	}
	// Check the last queue has been used
	if d2 == sq.cacheL {
		return nil, fmt.Errorf("no elements remaining in queue")
	}
	// Void cached slice pointer if not in use
	if sq.cacheF != sq.cacheL && d1 != sq.cacheL {
		sq.cache[d1] = nil
	}
	// Inc outer head pointer
	sq.cacheF = d2
	// Get next slice
	if ((sq.cacheF + 1) % len(sq.cache)) == sq.cacheL {
		// Only one slice remains: the head should take the tail's place, and tail should be void
		sq.head = sq.tail
		sq.headF, sq.headL = sq.tailF, sq.tailL
		sq.tail = nil
	} else {
		// Pointer taken from cache; dereference, and use as head
		sq.head = (*sq.cache[sq.cacheF].ptr)
		sq.cacheSize -= len(sq.head)
		hF := sq.cache[sq.cacheF].idx
		sq.headF, sq.headL = hF, hF
	}
	// Elem is nil; recurse until elem is found or cache is empty
	return sq.PeekFront()
}

// PeekBack - retrieve last element from queue without removing it
// Checks for empty queue, if not returns last elem
func (sq *Squeue) PeekBack() (interface{}, error) {
	var elem interface{}
	d1 := sq.cacheF - 1
	if d1 < 0 {
		d1 += len(sq.cache)
	}
	if sq.tail == nil {
		// Perform operation on head slice
		if sq.headL == 0 {
			elem = sq.head[len(sq.head)-1]
		} else {
			elem = sq.head[sq.headL-1]
		}
		// Return elem if found
		if elem != nil {
			return elem, nil
		}
		return nil, fmt.Errorf("no elements remaining in queue")
	}
	// Perform operation on tail slice
	if sq.tailL == 0 {
		elem = sq.tail[len(sq.tail)-1]
	} else {
		elem = sq.tail[sq.tailL-1]
	}
	// Return elem if found
	if elem != nil {
		return elem, nil
	}
	// No elems in tail; move to next slice in cache until only head left
	d2, d3 := sq.cacheL-1, sq.cacheL-2
	if d2 < 0 {
		d2 += len(sq.cache)
	}
	if d3 < 0 {
		d3 += len(sq.cache)
	}
	// Void cached slice pointer if not in use
	if sq.cacheL != sq.cacheF && sq.cacheL != d1 {
		sq.cache[sq.cacheL] = nil
	}
	// Dec outer tail pointer
	sq.cacheL = d2
	if sq.cacheL == ((sq.cacheF + 1) % len(sq.cache)) {
		// Last slice in cache, set tail to nil
		sq.tail = nil
	} else {
		// Take pointer from cache, dereference, use as tail
		sq.tail = (*sq.cache[d3].ptr)
		tF := sq.cache[d3].idx
		sq.tailF, sq.tailL = tF, tF
		sq.cacheSize -= len(sq.tail)
	}
	// Elem is nil; recurse until elem is found or cache is empty
	return sq.PeekBack()
}

// Unshift - remove element from front of queue (dequeue)
// Retrieves the elem, and if successful deletes its value in the slice
// Increments the head pointer to next elem in queue
func (sq *Squeue) Unshift() (interface{}, error) {
	// Get first elem in queue; if none, return error
	elem, err := sq.PeekFront()
	if err != nil {
		return nil, err
	}
	// Void element, move pointer
	sq.head[sq.headF] = nil
	sq.headF = (sq.headF + 1) % len(sq.head)

	return elem, nil
}

// Pop - remove element from back of queue
// Retrieves the elem, and if successful deletes its value in the slice
// Decrements the tail pointer to next elem in queue
func (sq *Squeue) Pop() (interface{}, error) {
	// Get last elem in queue; if none, return error
	elem, err := sq.PeekBack()
	if err != nil {
		return nil, err
	}
	// Void element, move pointer
	switch {
	case sq.tail == nil:
		// Perform operation on head slice
		sq.headL -= 1
		if sq.headL < 0 {
			sq.headL += len(sq.head)
		}
		sq.head[sq.headL] = nil
	default:
		// Perform operation on tail slice
		sq.tailL -= 1
		if sq.tailL < 0 {
			sq.tailL += len(sq.tail)
		}
		sq.tail[sq.tailL] = nil
	}

	return elem, nil
}

// Size - returns number of elements in queue
// O(1) amortized time complexity
// Cached slices record their length before caching, so only the size
// of the head and the tail must be found, which are O(1)
func (sq *Squeue) Size() int {
	return sq.headSize() + sq.cacheSize + sq.tailSize()
}

// Returns true if queue is empty
func (sq *Squeue) Empty() bool {
	return sq.Size() == 0
}

// Each - returns underlying slice for iteration - convinience method
// Method takes values from memory in O(n) time; iteration is done most performantly
// using delete operations (Unshift/Pop) until the queue is empty
// Ex.
// for !qq.Empty() {
//	   elem := qq.Pop()
//	   // do something with elem
// }
func (sq *Squeue) Each() []interface{} {
	s := make([]interface{}, 0)

	s = sq.appendHead(s)
	s = sq.appendCache(s)
	s = sq.appendTail(s)

	return s
}

// String - string representation: formats relevant slots of underlying slice as string
// Relies on Each() method to load values from memory - O(n)
func (sq *Squeue) String() string {
	return fmt.Sprint(sq.Each())
}

/* Internals */

// Resize slice to double the number of elements in the queue
// For small n, place values at beginning of larger slice to prevent unnecessary allocations
func (sq *Squeue) grow() {
	n := cap(sq.cache)
	if n < 6 {
		sq.resize(8)
	} else {
		sq.resize(2 * n)
	}
}

// Allocates bigger cache slices, copies elements from old onto new
// Slice is in circular order, and are reset to 0th index
// Reconfigures pointers to reflect shift
func (sq *Squeue) resize(m int) {
	// Allocate new cache slices
	var j int = 0
	qq := make([]*Cached, m)
	// Copy existing values to beginning of new slice
	switch {
	case sq.cacheF == 0:
		copy(qq, sq.cache)
		j = len(sq.cache)
	default:
		lenq := len(sq.cache)
		for i := sq.cacheF; i < lenq; i++ {
			qq[j] = sq.cache[i]
			j++
		}
		for i := 0; i < sq.cacheL; i++ {
			qq[j] = sq.cache[i]
			j++
		}
	}
	// Set pointers
	sq.cacheF = 0
	sq.cacheL = j
	// Set underlying slice as newly allocated slice
	sq.cache = qq
}

func (sq *Squeue) headSize() int {
	res := 0
	if sq.head == nil {
		return res
	}
	f, l := sq.headF, sq.headL
	switch {
	case f < l:
		res = l - f
	case f > l:
		res = len(sq.head) - f + l
	default:
		if sq.head[f] != nil {
			res = len(sq.head)
		}
	}
	return res
}

func (sq *Squeue) tailSize() int {
	res := 0
	if sq.tail == nil {
		return res
	}
	f, l := sq.tailF, sq.tailL
	switch {
	case f < l:
		res = l - f
	case f > l:
		res = len(sq.tail) - f + l
	default:
		if sq.tail[f] != nil {
			res = len(sq.tail)
		}
	}
	return res
}

// Add values from head into slice in queue order, return slice
func (sq *Squeue) appendHead(s []interface{}) []interface{} {
	q, p, c := sq.head, sq.headF, sq.headL
	lenq := len(q)
	if p < c {
		for j := p; j < c; j++ {
			s = append(s, q[j])
		}
	} else {
		for j := p; j < lenq; j++ {
			s = append(s, q[j])
		}
		for j := 0; j < c; j++ {
			s = append(s, q[j])
		}
	}
	return s
}

// Add values from tail into slice in queue order, return slice
func (sq *Squeue) appendTail(s []interface{}) []interface{} {
	q, p, c := sq.tail, sq.tailF, sq.tailL
	lenq := len(q)
	if p < c {
		for j := p; j < c; j++ {
			s = append(s, q[j])
		}
	} else {
		for j := p; j < lenq; j++ {
			s = append(s, q[j])
		}
		for j := 0; j < c; j++ {
			s = append(s, q[j])
		}
	}
	return s
}

// Get slice from reference in cache in queue order, then add their values in their queue order to a slice. Return the slice
func (sq *Squeue) appendCache(s []interface{}) []interface{} {
	switch {
	case sq.cacheF < sq.cacheL:
		for i := sq.cacheF + 1; i < sq.cacheL-1; i++ {
			s = sq.appendCacheInner(s, i)
		}
	default:
		lenQ := len(sq.cache)
		for i := sq.cacheF + 1; i < lenQ; i++ {
			s = sq.appendCacheInner(s, i)
		}
		for i := 0; i < sq.cacheL-1; i++ {
			s = sq.appendCacheInner(s, i)
		}
	}
	return s
}

// appendCache util
func (sq *Squeue) appendCacheInner(s []interface{}, i int) []interface{} {
	q, p := *(sq.cache[i].ptr), sq.cache[i].idx
	lenq := len(q)
	for j := p; j < lenq; j++ {
		s = append(s, q[j])
	}
	for j := 0; j < p; j++ {
		s = append(s, q[j])
	}
	return s
}

// Returns the maximum of two integers; if equal, returns the first arguemnt
func max(n, m int) int {
	if m > n {
		return m
	}
	return n
}

// Returns the minimum of two integers; if equal, returns the first arguemnt
func min(n, m int) int {
	if m < n {
		return m
	}
	return n
}
