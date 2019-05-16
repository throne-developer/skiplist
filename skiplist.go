// MIT License
//
// Copyright (c) 2018 Maurice Tollmien (maurice.tollmien@gmail.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package skiplist is an implementation of a skiplist to store elements in increasing order.
// It allows finding, insertion and deletion operations in approximately O(n log(n)).
// Additionally, there are methods for retrieving the next and previous element as well as changing the actual value
// without the need for re-insertion (as long as the key stays the same!)
// Skiplist is a fast alternative to a balanced tree.

// Improvement on the basis of https://github.com/MauriceGit/skiplist project，Using REDIS-ZSET mode to operate skiplist

package skiplist

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	MaxLevel     = 25
	Eps          = 0.00001
	HeadNodeName = "-inf-"
	TailNodeName = "+inf+"
)

/* Example data:
Lv4  -inf
Lv3  -inf     B
Lv2  -inf     B           F
Lv1  -inf  A  B  C  D  E  F  G  H  +inf

Each node points to the next node of the same layer or lower layer, such as:
B.next[0] = C
B.next[1] = F
B.next[2] = +inf

B.span[0] = 1
B.span[1] = 4

Prev points to the front node of the first layer, such as:
B.prev = A
A.prev = -inf
+inf.prev = H
*/
type Element struct {
	next  [MaxLevel]*Element /*points to the next node of the same layer or lower layer*/
	span  [MaxLevel]int32    /*distance to next node*/
	prev  *Element
	name  string
	score float64
}

func (e *Element) Name() string {
	return e.name
}

func (e *Element) Score() float64 {
	return e.score
}

func (e *Element) Next() *Element {
	if e.next[0] != nil && e.next[0].name == TailNodeName {
		return nil
	}
	return e.next[0]
}

func (e *Element) Prev() *Element {
	if e.prev != nil && e.prev.name == HeadNodeName {
		return nil
	}
	return e.prev
}

/*compare score first, then name*/
func (e *Element) Less(score float64, name string) bool {
	if math.Abs(e.score-score) > Eps {
		return e.score < score
	}
	return e.name < name
}

func (e *Element) Equal(score float64, name string) bool {
	return math.Abs(e.score-score) <= Eps && e.name == name
}

func (e *Element) Greater(score float64, name string) bool {
	if math.Abs(e.score-score) > Eps {
		return e.score > score
	}
	return e.name > name
}

type SkipList struct {
	headNode *Element
	tailNode *Element
	maxLevel int
	elements map[string]float64
}

func New() *SkipList {
	return NewSeed(time.Now().UTC().UnixNano())
}

func NewSeed(seed int64) *SkipList {
	rand.Seed(seed)

	headNode := &Element{
		next:  [MaxLevel]*Element{},
		span:  [MaxLevel]int32{},
		prev:  nil,
		name:  HeadNodeName,
		score: math.Inf(-1),
	}

	tailNode := &Element{
		next:  [MaxLevel]*Element{},
		span:  [MaxLevel]int32{},
		prev:  nil,
		name:  TailNodeName,
		score: math.Inf(1),
	}

	/*The height of the headNode layer is MaxLevel, and each layer points to tailNode.
	The height of the tailNode layer is 1. */
	for i := MaxLevel - 1; i >= 0; i-- {
		headNode.next[i] = tailNode
		headNode.span[i] = 1
	}
	tailNode.prev = headNode

	return &SkipList{
		headNode: headNode,
		tailNode: tailNode,
		maxLevel: 0,
		elements: make(map[string]float64),
	}
}

func (t *SkipList) Insert(name string, score float64) {
	if name == HeadNodeName || name == TailNodeName {
		return
	}

	if currScore, ok := t.elements[name]; ok {
		if equalFloat(currScore, score) { /*no change*/
			return
		} else {
			t.Delete(name) /*delete old node*/
		}
	}

	elemLevel := generateLevel() /*Number of new node layers*/
	if elemLevel > t.maxLevel {
		elemLevel = t.maxLevel + 1
		t.maxLevel = elemLevel
	}

	elem := &Element{
		next:  [MaxLevel]*Element{},
		span:  [MaxLevel]int32{},
		prev:  nil,
		name:  name,
		score: score,
	}
	t.elements[name] = score

	var (
		index    = t.maxLevel
		currNode = t.headNode

		prevs = [MaxLevel]struct {
			node *Element
			rank int32
		}{}
	)

	for {
		nextNode := currNode.next[index]

		if !nextNode.Less(elem.score, elem.name) {
			prevs[index].node = currNode

			/*Within the elemLevel layer range, a new node needs to be inserted into each layer*/
			if index <= elemLevel {
				elem.next[index] = nextNode
				currNode.next[index] = elem

				if index == 0 { /*In the first layer, update prev*/
					elem.prev = currNode
					nextNode.prev = elem
				}
			}
		}

		if nextNode.Less(elem.score, elem.name) { /*Search right or down*/
			prevs[index].rank += currNode.span[index] /*Accumulate span as node rank*/
			currNode = nextNode

		} else {
			if index--; index < 0 {
				break
			} else {
				prevs[index].rank = prevs[index+1].rank
			}
		}
	}

	/*Update the node span, according to the rank and span of the pre-node and the rank of the new node*/
	elemRank := prevs[0].rank + 1
	for i := 0; i <= t.maxLevel; i++ {
		elem.span[i] = prevs[i].rank + prevs[i].node.span[i] - elemRank
		prevs[i].node.span[i] = elemRank - prevs[i].rank
	}
}

func (t *SkipList) Find(name string) (foundItem *Element) {
	score, ok := t.elements[name]
	if !ok {
		return
	}

	currNode := t.headNode

	for i := t.maxLevel; i >= 0; i-- {

		nextNode := currNode.next[i]
		for nextNode.Less(score, name) {
			currNode = nextNode
			nextNode = nextNode.next[i]
		}

		if nextNode.Equal(score, name) {
			foundItem = nextNode
			break
		}
	}
	return
}

func (t *SkipList) FindGreaterOrEqual(score float64) (foundItem *Element) {
	if t.IsEmpty() {
		return
	}

	/*Score <= minimum score, return the minimum node*/
	if first := t.headNode.next[0]; !greaterThan(score, first.score) {
		foundItem = first
		return
	}

	/*Score > maximum score, return nil*/
	if last := t.tailNode.prev; greaterThan(score, last.score) {
		return
	}

	currNode := t.headNode

	for i := t.maxLevel; i >= 0; i-- {

		nextNode := currNode.next[i]
		for lessThan(nextNode.score, score) {

			currNode = nextNode
			nextNode = nextNode.next[i]
		}

		/*Go to the first level, search for the first >= score node from currNode*/
		if i == 0 {
			for curr := currNode; curr != t.tailNode; curr = curr.next[0] {
				if !lessThan(curr.score, score) {
					foundItem = curr
					break
				}
			}
		}
	}
	return
}

func (t *SkipList) Delete(name string) {
	score, ok := t.elements[name]
	if !ok {
		return
	}

	currNode := t.headNode

	for i := t.maxLevel; i >= 0; i-- {

		nextNode := currNode.next[i]
		for nextNode.Less(score, name) {

			currNode = nextNode
			nextNode = nextNode.next[i]
		}

		if nextNode.Equal(score, name) { /*delete node*/
			delNode := nextNode
			currNode.next[i] = delNode.next[i]

			if i == 0 {
				delNode.next[i].prev = currNode
				delete(t.elements, name)
			}

			if t.headNode.next[i] == t.tailNode && i > 0 { /*empty layer*/
				t.maxLevel = i - 1
			}
		}
	}
}

func (t *SkipList) GetRank(name string) (rank int, exist bool) {
	score, ok := t.elements[name]
	if !ok {
		return
	}

	currNode := t.headNode
	var elemRank int32

	for i := t.maxLevel; i >= 0; i-- {

		nextNode := currNode.next[i]
		for nextNode.Less(score, name) {
			elemRank += currNode.span[i]

			currNode = nextNode
			nextNode = nextNode.next[i]
		}

		if nextNode.Equal(score, name) {
			rank = int(elemRank + currNode.span[i])
			exist = true
			break
		}
	}
	return
}

func (t *SkipList) FindByRank(rank int) (foundItem *Element) {
	if rank < 1 || rank > len(t.elements) {
		return nil
	}

	currNode := t.headNode
	var elemRank int32

	for i := t.maxLevel; i >= 0; i-- {

		for currNode.next[i] != t.tailNode {

			if nextRank := elemRank + currNode.span[i]; int(nextRank) <= rank {
				elemRank = nextRank
				currNode = currNode.next[i]
			} else {
				break
			}
		}

		if int(elemRank) == rank {
			foundItem = currNode
			break
		}
	}
	return
}

func (t *SkipList) IsEmpty() bool {
	return t.headNode.next[0] == t.tailNode
}

func (t *SkipList) GetSmallestNode() *Element {
	if !t.IsEmpty() {
		return t.headNode.next[0]
	}
	return nil
}

func (t *SkipList) GetLargestNode() *Element {
	if !t.IsEmpty() {
		return t.tailNode.prev
	}
	return nil
}

func (t *SkipList) GetNodeCount() int {
	return len(t.elements)
}

func (t *SkipList) GetScore(name string) (float64, bool) {
	score, ok := t.elements[name]
	return score, ok
}

func (t *SkipList) PrintNodes() string {
	levels := make([]string, 0, MaxLevel)
	var buff bytes.Buffer

	for i := t.maxLevel; i >= 0; i-- {
		buff.Reset()
		buff.WriteString("[" + strconv.Itoa(i) + "] ")

		for node := t.headNode; node != nil; node = node.next[i] {
			span := node.span[i]
			if node.next[i] == t.tailNode {
				span = 0
			}

			buff.WriteString(node.name)
			buff.WriteString(fmt.Sprintf(" -(%d)> ", span))
		}

		levels = append(levels, buff.String())
	}

	return strings.Join(levels, "\n")
}

func (t *SkipList) PrintLevels() string {
	levels := make([]string, 0, MaxLevel)
	wholeCount := 0

	for i := t.maxLevel; i >= 0; i-- {
		count := 0
		for node := t.headNode.next[i]; node != t.tailNode; node = node.next[i] {
			count++
		}

		levels = append(levels, fmt.Sprintf("[%02d] %d", i, count))
		wholeCount += count
	}

	levels = append(levels, "whole count="+strconv.Itoa(wholeCount))
	return strings.Join(levels, "\n")
}

/*Return random layers*/
func generateLevel() int {
	var x uint64 = rand.Uint64() & ((1 << uint(MaxLevel-1)) - 1) /*Random value x, bit number < MAX_LEVEL*/
	zeroes := bits.TrailingZeros64(x)                            /*Starting from the tail, the number of bits 0*/

	level := MaxLevel - 1
	if zeroes < MaxLevel {
		level = zeroes
	}
	return level
}

func greaterThan(a, b float64) bool {
	return a > b && math.Abs(a-b) > Eps
}

func lessThan(a, b float64) bool {
	return a < b && math.Abs(a-b) > Eps
}

func equalFloat(a, b float64) bool {
	return math.Abs(a-b) <= Eps
}
