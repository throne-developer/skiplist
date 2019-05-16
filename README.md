# skiplist
Using **REDIS-ZSET mode** to operate skiplist

Improvement on the basis of https://github.com/MauriceGit/skiplist project，Using REDIS-ZSET mode to operate skiplist

# Example

```golang
func TestSkipListSimple(t *testing.T) {
	sl := New()

	sl.Insert("A", 1)
	sl.Insert("B", 2)
	sl.Insert("C", 3)
	sl.Insert("D1", 6)
	sl.Insert("D2", 6)
	sl.Insert("D3", 6)

	if elem := sl.Find("D1"); elem != nil {
		fmt.Println("Find D1, score=", elem.Score())
	}

	if rank, ok := sl.GetRank("D1"); ok {
		fmt.Println("D1 rank is ", rank)
	}

	if score, ok := sl.GetScore("D1"); ok {
		fmt.Println("D1 score is ", score)
	}

	if elem := sl.FindByRank(5); elem != nil {
		fmt.Println("FindByRank 5, name=", elem.Name())
	}

	for elem := sl.FindGreaterOrEqual(3); elem != nil; elem = elem.Next() {
		fmt.Println("FindGreaterOrEqual 3, name=", elem.Name())
	}

	sl.Delete("A")
}
```

**output**
```bash
Find D1, score= 6
D1 rank is  4
D1 score is  6
FindByRank 5, name= D2
FindGreaterOrEqual 3, name= C
FindGreaterOrEqual 3, name= D1
FindGreaterOrEqual 3, name= D2
FindGreaterOrEqual 3, name= D3
```


