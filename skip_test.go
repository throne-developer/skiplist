package skiplist

import (
	"fmt"
	"strings"
	"testing"
)

func TestSkipList1(t *testing.T) {
	data := make(map[string]float64)
	for i := 1; i <= 1000; i++ {
		data[fmt.Sprintf("v_%d", i)] = float64(i)
	}

	for i := 1; i <= 100; i++ {
		if msg := checkSkiplist(data); msg != "" {
			fmt.Println("TestSkipList1 checkSkiplist err: " + msg)
		}
	}
	fmt.Println("TestSkipList1 success")
}

func checkSkiplist(data map[string]float64) string {
	slist := New()

	/*插入数据*/
	for name, score := range data {
		slist.Insert(name, score)
	}

	if slist.GetNodeCount() != len(data) {
		return fmt.Sprintf("GetNodeCount failed")
	}

	for name, score := range data {
		/*查找元素*/
		if node := slist.Find(name); node == nil || node.Name() != name || int(node.Score()) != int(score) {
			return fmt.Sprintf("Find failed, name=%s", name)
		}
		/*获取rank*/
		rank, ok := slist.GetRank(name)
		if !ok || rank != int(score) {
			return fmt.Sprintf("GetRank failed, name=%s, rank=%d", name, rank)
		}
		/*根据rank查找*/
		if node := slist.FindByRank(rank); node == nil || node.Name() != name {
			return fmt.Sprintf("FindByRank wrong, name=%s, rank=%d", name, rank)
		}
		/*根据score查找*/
		if node := slist.FindGreaterOrEqual(score); node == nil || node.Name() != name {
			return fmt.Sprintf("FindGreaterOrEqual wrong, name=%s, rank=%d", name, rank)
		}
		/*获取score*/
		if getScore, ok := slist.GetScore(name); !ok || int(getScore) != int(score) {
			return fmt.Sprintf("GetScore wrong, name=%s, rank=%d", name, rank)
		}
	}

	/*删除元素*/
	for name := range data {
		slist.Delete(name)
		if slist.Find(name) != nil {
			return fmt.Sprintf("Delete failed, name=%s", name)
		}
	}

	if slist.GetNodeCount() != 0 {
		return fmt.Sprintf("GetNodeCount failed")
	}

	return ""
}

func TestSkipList2(t *testing.T) {
	sl := NewSeed(1)

	sl.Insert("a10", 1)
	sl.Insert("a20", 2)
	sl.Insert("a30", 3)
	sl.Insert("a40", 4)
	sl.Insert("a50", 5)
	sl.Insert("a61", 6)
	sl.Insert("a62", 6)
	sl.Insert("a63", 6)
	sl.Insert("a70", 7)
	sl.Insert("a81", 8)
	sl.Insert("a82", 8)
	sl.Insert("a83", 8)
	sl.Insert("a90", 9)

	sl.Insert("a63", 6)
	sl.Insert("a70", 7)

	fmt.Println(sl.PrintNodes())

	name := "a82"
	sl.Delete("a82")
	fmt.Println("delete " + name)
	fmt.Println(sl.PrintNodes())

	sl.Insert(name, 8)
	fmt.Println("recovery " + name)
	fmt.Println(sl.PrintNodes())

	for _, name := range []string{"a10", "a61", "a63", "a64", "a81", "a90", "a11"} {
		if elem := sl.Find(name); elem != nil {
			fmt.Println("Find "+name+" score=", elem.score)
		} else {
			fmt.Println("Find " + name + " failed")
		}
	}

	for _, score := range []float64{0, 1, 5, 9, 10} {
		results := make([]string, 0)
		if elem := sl.FindGreaterOrEqual(score); elem != nil {
			curr := elem
			for curr != nil {
				results = append(results, curr.name)
				curr = curr.Next()
			}
		}
		fmt.Println("FindGreaterOrEqual", score, ":", strings.Join(results, ", "))
	}

	names := []string{"a10", "a61", "a63", "a64", "a81", "a90", "a11", "a62"}
	for _, name := range names {
		sl.Delete(name)
	}
	fmt.Println(sl.PrintNodes())
}
