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
// I am a Chinese programmer，So the code comment is in Chinese.
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
	MAX_LEVEL = 25
	EPS       = 0.00001
)

var debug = false

func OpenDebugMode() {
	debug = true
}

func CloseDebugMode() {
	debug = false
}

/* 

跳跃表结构，每个结点指向同一层或下层的next结点，比如 B.next->F、C、C， C.next->F、D
prev指向第一层的前置结点，比如 B.prev->A，D.prev->C

3层     B           F
2层     B  C        F
1层  A  B  C  D  E  F  G  H
*/
type SkipListElement struct {
	next  [MAX_LEVEL]*SkipListElement /*指向同一层或下层的next结点*/
	prev  *SkipListElement            /*第一层的prev结点*/
	name  string                      /*名称*/
	score float64                     /*分值*/
	level int                         /*结点占据层数*/
}

func (e *SkipListElement) Name() string {
	return e.name
}

func (e *SkipListElement) Score() float64 {
	return e.score
}

func (e *SkipListElement) Next() *SkipListElement {
	return e.next[0]
}

func (e *SkipListElement) Prev() *SkipListElement {
	return e.prev
}

func (e *SkipListElement) less(score float64, name string) bool { /*先比较score，再比较name*/
	if math.Abs(e.score-score) > EPS {
		return e.score < score
	}
	return e.name < name
}

func (e *SkipListElement) equal(score float64, name string) bool {
	return math.Abs(e.score-score) <= EPS && e.name == name
}

func (e *SkipListElement) greater(score float64, name string) bool {
	if math.Abs(e.score-score) > EPS {
		return e.score > score
	}
	return e.name > name
}

type SkipList struct {
	startLevels [MAX_LEVEL]*SkipListElement /*指向每一层的起始结点*/
	endLevels   [MAX_LEVEL]*SkipListElement /*指向每一层的最后结点*/
	maxLevel    int                         /*当前最大层数*/
	elements    map[string]float64          /*每个结点的分值*/
}

func New() *SkipList {
	return NewSeed(time.Now().UTC().UnixNano())
}

func NewSeed(seed int64) *SkipList {
	rand.Seed(seed)
	return &SkipList{
		startLevels: [MAX_LEVEL]*SkipListElement{},
		endLevels:   [MAX_LEVEL]*SkipListElement{},
		elements:    make(map[string]float64),
	}
}

func (t *SkipList) Insert(name string, score float64) {
	t.Delete(name) /*删除旧结点*/

	elemLevel := generateLevel() /*获取随机层数*/
	if elemLevel > t.maxLevel {  /*最大层数+1*/
		elemLevel = t.maxLevel + 1
		t.maxLevel = elemLevel
	}

	t.elements[name] = score
	elem := &SkipListElement{
		next:  [MAX_LEVEL]*SkipListElement{},
		name:  name,
		score: score,
		level: elemLevel,
	}

	/*是否最小，最大，中间元素，默认为首次插入*/
	isMin, isMax, isMid := true, true, false
	if !t.IsEmpty() {
		isMin = elem.less(t.startLevels[0].score, t.startLevels[0].name)
		isMax = elem.greater(t.endLevels[0].score, t.endLevels[0].name)
	}

	if !isMin && !isMax { /*中间插入*/

		/*获取level层索引，或者>level层的 首结点<=新结点的 层索引*/
		index := t.findStartLevel(elem.score, elem.name, elemLevel)
		isMid = true
		var currNode, nextNode *SkipListElement
		var iters int

		for {
			iters++
			if currNode == nil {
				nextNode = t.startLevels[index]
			} else {
				nextNode = currNode.next[index]
			}

			/*在层数范围内，每层需要插入一个新结点； 插入第一个大于新结点的结点前面，若没有，插入层尾*/
			if index <= elemLevel && (nextNode == nil || nextNode.greater(elem.score, elem.name)) {
				elem.next[index] = nextNode /*在 currNode 和 nextNode 中间插入*/
				if currNode != nil {
					currNode.next[index] = elem
				}

				if index == 0 { /*在第一层时，处理prev指针*/
					elem.prev = currNode
					nextNode.prev = elem
				}
			}

			/*优先向本层右侧扫描，搜索本层最后一个小于新结点的结点，作为currNode；
			否则下降一层，继续向右侧扫描，currNode可以保持不变*/
			if nextNode != nil && nextNode.less(elem.score, elem.name) {
				currNode = nextNode
			} else {
				if index--; index < 0 {
					break
				}
			}
		}

		if debug {
			fmt.Println("Insert Iteration times ", iters)
		}
	}

	for i := elemLevel; i >= 0; i-- { /*处理层首和层尾*/
		/*最小元素必为第一层的层首，中间元素在顶部某些层也可为层首*/
		if isMin || isMid {
			/*层首为空，或层首大于新结点，需要将新结点作为层首*/
			if t.startLevels[i] == nil || t.startLevels[i].greater(elem.score, elem.name) {
				if i == 0 && t.startLevels[i] != nil { /*第一层时，更新prev*/
					t.startLevels[i].prev = elem
				}

				elem.next[i] = t.startLevels[i]
				t.startLevels[i] = elem
			}

			if elem.next[i] == nil { /*也可以为层尾*/
				t.endLevels[i] = elem
			}
		}

		if isMax {
			if !isMin { /*避免将第一个元素（同时满足isMax和isMin）链接到自身*/
				if t.endLevels[i] != nil {
					t.endLevels[i].next[i] = elem
				}
				if i == 0 {
					elem.prev = t.endLevels[i]
				}
				t.endLevels[i] = elem
			}

			/*最大元素，也有可能是层首*/
			if t.startLevels[i] == nil || t.startLevels[i].greater(elem.score, elem.name) {
				t.startLevels[i] = elem
			}
		}
	}
}

/*返回level层索引，或者>level层的 首结点<=新结点的 层索引*/
func (t *SkipList) findStartLevel(score float64, name string, level int) int {
	for i := t.maxLevel; i >= 0; i-- {
		if i <= level {
			return i
		}
		if t.startLevels[i] != nil && !t.startLevels[i].greater(score, name) {
			return i
		}
	}
	return 0
}

func (t *SkipList) Find(name string) (foundItem *SkipListElement) {
	score, ok := t.elements[name]
	if !ok {
		return
	}

	index := t.findStartLevel(score, name, 0) /*从上到下，获取第一个 首结点<=新结点的 层索引*/
	var currNode, nextNode *SkipListElement
	var iterNodes []string

	for {
		if currNode == nil {
			nextNode = t.startLevels[index]
		} else {
			nextNode = currNode.next[index]
		}

		if debug {
			if nextNode != nil {
				iterNodes = append(iterNodes, nextNode.name)
			}
		}

		if nextNode != nil && nextNode.equal(score, name) { /*找到目标*/
			foundItem = nextNode
			break
		}

		if nextNode != nil && nextNode.less(score, name) {
			currNode = nextNode
		} else {
			if index--; index < 0 {
				break
			}
		}
	}

	if debug {
		fmt.Println("Find Iteration times ", len(iterNodes), ", path: "+strings.Join(iterNodes, ", "))
	}
	return
}

func (t *SkipList) FindGreaterOrEqual(score float64) (foundItem *SkipListElement) {
	if t.IsEmpty() {
		return
	}
	if lessThan(score, t.startLevels[0].score) || equalFloat(score, t.startLevels[0].score) {
		foundItem = t.startLevels[0] /*score<=最小分值，直接返回最小结点*/
		return
	}
	if greaterThan(score, t.endLevels[0].score) {
		return /*score>最大分值，返回nil*/
	}

	index := 0
	for i := t.maxLevel; i >= 0; i-- { /*从上到下，获取第一个 首结点分值<score 的层索引*/
		if lessThan(t.startLevels[i].score, score) {
			index = i
			break
		}
	}

	var currNode, nextNode *SkipListElement
	var iters int

	for {
		iters++
		if currNode == nil {
			nextNode = t.startLevels[index]
		} else {
			nextNode = currNode.next[index]
		}

		if index == 0 { /*进入第一层，从currNode向右搜索第一个 >=score的结点*/
			for node := currNode; node != nil; {
				if equalFloat(node.score, score) || greaterThan(node.score, score) {
					foundItem = node
					break
				}
				node = node.next[0]
			}
			break
		}

		if nextNode != nil && lessThan(nextNode.score, score) {
			currNode = nextNode
		} else {
			if index--; index < 0 {
				break
			}
		}
	}

	if debug {
		fmt.Println("FindGreaterOrEqual Iteration times ", iters)
	}
	return
}

func (t *SkipList) Delete(name string) {
	score, ok := t.elements[name]
	if !ok {
		return
	}

	index := t.findStartLevel(score, name, 0)
	var currNode, nextNode *SkipListElement
	var iters int

	for {
		iters++
		if currNode == nil {
			nextNode = t.startLevels[index]
		} else {
			nextNode = currNode.next[index]
		}

		if nextNode != nil && nextNode.equal(score, name) { /*删除node*/
			delNode := nextNode
			if currNode != nil {
				currNode.next[index] = delNode.next[index]
			}

			if index == 0 { /*处理第1层*/
				if delNode.next[index] != nil {
					delNode.next[index].prev = currNode
				}
				delete(t.elements, name)
			}

			if t.startLevels[index] == delNode { /*更新层首*/
				t.startLevels[index] = delNode.next[index]

				if t.startLevels[index] == nil { /*层数-1*/
					t.maxLevel = index - 1
				}
			}

			if delNode.next[index] == nil { /*更新层尾*/
				t.endLevels[index] = currNode
			}
			delNode.next[index] = nil
		}

		if nextNode != nil && nextNode.less(score, name) {
			currNode = nextNode
		} else {
			if index--; index < 0 {
				break
			}
		}
	}

	if debug {
		fmt.Println("Delete Iteration times ", iters)
	}
}

func (t *SkipList) IsEmpty() bool {
	return t.startLevels[0] == nil
}

func (t *SkipList) GetSmallestNode() *SkipListElement {
	return t.startLevels[0]
}

func (t *SkipList) GetLargestNode() *SkipListElement {
	return t.endLevels[0]
}

func (t *SkipList) GetNodeCount() int {
	return len(t.elements)
}

func (t *SkipList) PrintNodes() string {
	levels := make([]string, 0, MAX_LEVEL)
	var buff bytes.Buffer

	for i := t.maxLevel; i >= 0; i-- {
		buff.Reset()
		buff.WriteString("[" + strconv.Itoa(i) + "] ")

		currNode := t.startLevels[i]
		for currNode != nil {
			buff.WriteString(currNode.name)
			buff.WriteString(" -> ")

			currNode = currNode.next[i]
		}

		buff.WriteString("(" + t.endLevels[i].name + ")")
		levels = append(levels, buff.String())
	}

	return strings.Join(levels, "\n")
}

func (t *SkipList) PrintLevels() string {
	levels := make([]string, 0, MAX_LEVEL)
	wholeCount := 0

	for i := t.maxLevel; i >= 0; i-- {
		count := 0

		currNode := t.startLevels[i]
		for currNode != nil {
			count++
			currNode = currNode.next[i]
		}

		levels = append(levels, fmt.Sprintf("[%02d] %d", i, count))
		wholeCount += count
	}

	levels = append(levels, "whole count="+strconv.Itoa(wholeCount))
	return strings.Join(levels, "\n")
}

func generateLevel() int { /*返回随机层数*/
	var x uint64 = rand.Uint64() & ((1 << uint(MAX_LEVEL-1)) - 1) /*随机值x，bit位数<=MAX_LEVEL*/
	zeroes := bits.TrailingZeros64(x)                             /*从尾部开始，bit位为0的个数*/

	level := MAX_LEVEL - 1
	if zeroes < MAX_LEVEL {
		level = zeroes
	}
	return level
}

func greaterThan(a, b float64) bool {
	return a > b && math.Abs(a-b) > EPS
}

func lessThan(a, b float64) bool {
	return a < b && math.Abs(a-b) > EPS
}

func equalFloat(a, b float64) bool {
	return math.Abs(a-b) <= EPS
}
