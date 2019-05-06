package skiplist

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSkipListSimple(t *testing.T) {
	sl := New()

	sl.Insert("A", 1)
	sl.Insert("B", 2)
	sl.Insert("C", 3)
	sl.Insert("D1", 6)
	sl.Insert("D2", 6)
	sl.Insert("D3", 6)

	if elem := sl.Find("D1"); elem != nil {
		fmt.Println("Find D1 score=", elem.Score())
	}

	if elem := sl.FindGreaterOrEqual(3); elem != nil {
		fmt.Println("FindGreaterOrEqual 3 name=", elem.Name())
	}

	sl.Delete("A")
}

func TestSkipList(t *testing.T) {
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

	OpenDebugMode()
	for _, name := range []string{"a10", "a61", "a63", "a64", "a81", "a90", "a11"} {
		if elem := sl.Find(name); elem != nil {
			fmt.Println("Find "+name+" score=", elem.score)
		} else {
			fmt.Println("Find " + name + " failed")
		}
	}
	CloseDebugMode()

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
		fmt.Println("delete", name)
		fmt.Println(sl.PrintNodes())
	}

}

func TestSkipBench(t *testing.T) {
	sl := New()
	N := 100 * 10000

	/*Init*/
	st := time.Now()
	datas := make(map[string]float64)
	for i := 1; i <= N; i++ {
		datas[strconv.Itoa(i)] = float64(i)
	}
	dur := time.Since(st)
	fmt.Println("InitMap use "+dur.String()+", average ", dur.Seconds()*1000/float64(N), "ms")

	/*Insert*/
	st = time.Now()
	for name, score := range datas {
		sl.Insert(name, score)
	}
	dur = time.Since(st)
	fmt.Println("Insert use "+dur.String()+", average ", dur.Seconds()*1000/float64(N), "ms")
	fmt.Println(sl.PrintLevels())

	/*Find*/
	st = time.Now()
	for name := range datas {
		sl.Find(name)
	}
	dur = time.Since(st)
	fmt.Println("Find use "+dur.String()+", average ", dur.Seconds()*1000/float64(N), "ms")

	/*FindGreaterOrEqual*/
	st = time.Now()
	for _, score := range datas {
		if rand.Intn(100000) == 1 {
			OpenDebugMode()
			sl.FindGreaterOrEqual(score)
			CloseDebugMode()

		} else {
			sl.FindGreaterOrEqual(score)
		}
	}
	dur = time.Since(st)
	fmt.Println("FindGreaterOrEqual use "+dur.String()+", average ", dur.Seconds()*1000/float64(N), "ms")

	/*Delete*/
	st = time.Now()
	for name := range datas {
		sl.Delete(name)
	}
	dur = time.Since(st)
	fmt.Println("Delete use "+dur.String()+", average ", dur.Seconds()*1000/float64(N), "ms")
}
