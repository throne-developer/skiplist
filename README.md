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
		fmt.Println("Find D1 score=", elem.Score())
	}

	if elem := sl.FindGreaterOrEqual(3); elem != nil {
		fmt.Println("FindGreaterOrEqual 3 name=", elem.Name())
	}

	sl.Delete("A")
}
```

**output**
>Find D1 score= 6<br>FindGreaterOrEqual 3 name= C

# Bench
```golang
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
```

**output**
```bash
InitMap use 564ms, average  0.000564 ms

Insert use 3.464s, average  0.003464 ms

[level 20] 1
[level 19] 2
[level 18] 4
[level 17] 6
[level 16] 14
[level 15] 29
[level 14] 63
[level 13] 121
[level 12] 240
[level 11] 465
[level 10] 928
[level 09] 1886
[level 08] 3774
[level 07] 7768
[level 06] 15567
[level 05] 31114
[level 04] 62503
[level 03] 124536
[level 02] 249941
[level 01] 499940
[level 00] 1000000
whole count=1998902

Find use 2.576s, average  0.002576 ms

FindGreaterOrEqual Iteration times  29
FindGreaterOrEqual Iteration times  37
FindGreaterOrEqual Iteration times  39
FindGreaterOrEqual Iteration times  41
FindGreaterOrEqual Iteration times  29

FindGreaterOrEqual use 2.334s, average  0.002334 ms

Delete use 2.38s, average  0.00238 ms
```
