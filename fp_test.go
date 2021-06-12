package fp

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TestFPTestSuite struct {
	suite.Suite
}

func (suite *TestFPTestSuite) SetupTest() {
}

func (suite *TestFPTestSuite) TearDownTest() {
}

func TestTestFPTestSuite(t *testing.T) {
	suite.Run(t, new(TestFPTestSuite))
}

func (suite *TestFPTestSuite) TestMapString() {
	slice := []string{"a", "b", "c"}
	var out []string
	StreamOf(slice).Map(strings.ToUpper).ToSlice(&out)
	suite.ElementsMatch(out, []string{"A", "B", "C"})
}

func (suite *TestFPTestSuite) TestMapSelectString() {
	slice := []string{"a", "b", "c"}
	var out []string
	StreamOf(slice).FlatMap(func(e string) (string, bool) {
		return strings.ToUpper(e), e == "b"
	}).ToSlice(&out)
	suite.ElementsMatch(out, []string{"B"})

	out = StreamOf(slice).FlatMap(func(e string) (string, bool) {
		return strings.ToUpper(e), e == "x"
	}).Strings()
	suite.ElementsMatch(out, []string{})

	out = StreamOf(slice).FlatMap(func(e string) (string, bool) {
		return strings.ToUpper(e), e == "a" || e == "c"
	}).Strings()
	suite.ElementsMatch(out, []string{"A", "C"})
}

func (suite *TestFPTestSuite) TestFlatMapErr() {
	gerr := func(c bool) error {
		if c {
			return errors.New("ERR")
		}
		return nil
	}
	slice := []string{"a", "b", "c"}
	var out []string
	StreamOf(slice).FlatMap(func(e string) (string, error) {
		return strings.ToUpper(e), gerr(e == "a" || e == "c")
	}).ToSlice(&out)
	suite.ElementsMatch(out, []string{"B"})

	out = StreamOf(slice).FlatMap(func(e string) (string, error) {
		return strings.ToUpper(e), gerr(true)
	}).Strings()
	suite.ElementsMatch(out, []string{})

	out = StreamOf(slice).FlatMap(func(e string) (string, error) {
		return strings.ToUpper(e), gerr(e == "b")
	}).Strings()
	suite.ElementsMatch(out, []string{"A", "C"})
}

func (suite *TestFPTestSuite) TestRepeatableGetValueMapString() {
	slice := []string{"a", "b", "c"}
	q := StreamOf(slice).Map(strings.ToUpper)
	out := q.Strings()
	suite.ElementsMatch(out, []string{"A", "B", "C"})

	out = q.Strings()
	suite.ElementsMatch(out, []string{"A", "B", "C"})
}

func (suite *TestFPTestSuite) TestRepeatableGetValueMapChanString() {
	slice := make(chan string, 10)
	slice <- "a"
	slice <- "b"
	slice <- "c"
	close(slice)
	q := StreamOf(slice).Map(strings.ToUpper)
	out := q.Strings()
	suite.ElementsMatch([]string{"A", "B", "C"}, out)

	out = q.Strings()
	suite.ElementsMatch([]string{"A", "B", "C"}, out)
}

func (suite *TestFPTestSuite) TestMapStringPtr() {
	ptr := func(s string) *string { return &s }
	slice := []*string{nil, ptr("a"), nil, ptr("c")}
	var out []string
	StreamOf(slice).Map(func(s *string) string {
		if s == nil {
			return ""
		}
		return strings.ToUpper(*s)
	}).ToSlice(&out)
	suite.ElementsMatch(out, []string{"", "A", "", "C"})
}

func (suite *TestFPTestSuite) TestMapEmptySlice() {
	slice := []string{}
	var out []string
	StreamOf(slice).Map(strings.ToUpper).ToSlice(&out)
	suite.Len(out, 0)

	slice = nil
	StreamOf(slice).Map(strings.ToUpper).ToSlice(&out)
	suite.Len(out, 0)
}

func (suite *TestFPTestSuite) TestMapEmptySliceResultType() {
	var slice []string
	out := StreamOf(slice).Map(strings.ToUpper).Result()
	suite.Nil(out)

	ret, ok := out.([]string)
	suite.True(ok)
	suite.Nil(ret)
}

func (suite *TestFPTestSuite) TestLazyMap() {
	var cnt int
	slice := []string{"a", "b", "c"}
	q := StreamOf(slice).Map(func(s string) string {
		cnt++
		return s
	})
	suite.Equal(0, cnt)

	q.Result()
	suite.Equal(3, cnt)
}

func (suite *TestFPTestSuite) TestFilter() {
	slice := []string{"a", "b", "c"}
	out := StreamOf(slice).Filter(func(s string) bool {
		return s == "b"
	}).Strings()
	suite.Equal([]string{"b"}, out)
}

func (suite *TestFPTestSuite) TestReject() {
	slice := []string{"a", "b", "c"}
	out := StreamOf(slice).Reject(func(s string) bool {
		return s == "b"
	}).Strings()
	suite.Equal([]string{"a", "c"}, out)
}

func (suite *TestFPTestSuite) TestLazyFilterMap() {
	slice := []string{"a", "b", "c"}
	var cnt int
	out := StreamOf(slice).Filter(func(s string) bool {
		return s == "b"
	}).Map(func(s string) string {
		cnt++
		return strings.ToUpper(s)
	}).Strings()
	suite.Equal([]string{"B"}, out)
	suite.Equal(1, cnt)
}

func (suite *TestFPTestSuite) TestForeach() {
	var out string
	slice := []string{"abc", "de", "f"}
	out1 := StreamOf(slice).Foreach(func(s string) {
		out += s
	}).Strings()
	suite.Equal("abcdef", out)
	suite.ElementsMatch(slice, out1)
}

func (suite *TestFPTestSuite) TestForeachWithIndex() {
	var out string
	slice := []string{"abc", "de", "f"}
	var indics []int
	out1 := StreamOf(slice).Foreach(func(s string, i int) {
		out += s
		indics = append(indics, i)
	}).Strings()
	suite.Equal("abcdef", out)
	suite.ElementsMatch(slice, out1)
	suite.ElementsMatch([]int{0, 1, 2}, indics)
}

func (suite *TestFPTestSuite) TestReadFirst() {
	slice := []string{"abc", "de", "f"}
	q := StreamOf(slice)
	out := q.First()
	suite.Equal("abc", out.String())

	out = q.First()
	suite.Equal("abc", out.String())

	ch := make(chan string, 3)
	ch <- "a"
	ch <- "b"
	ch <- "c"
	close(ch)

	q = StreamOf(ch)
	out = q.First()
	suite.Equal("a", out.String())

	out = q.First()
	suite.Equal("a", out.String())
}

func (suite *TestFPTestSuite) TestReadChan() {
	ch := make(chan string, 3)
	ch <- "a"
	ch <- "b"
	ch <- "c"
	close(ch)

	out := StreamOf(ch).Strings()
	suite.ElementsMatch([]string{"a", "b", "c"}, out)
}

func (suite *TestFPTestSuite) TestReduce() {
	source := []string{"a", "b", "c", "d", "a", "c"}

	out := StreamOf(source).Reduce(map[string]int{}, func(memo map[string]int, s string) map[string]int {
		memo[s] += 1
		return memo
	}).Result().(map[string]int)
	suite.Equal(map[string]int{
		"a": 2,
		"b": 1,
		"c": 2,
		"d": 1,
	}, out)
}

func (suite *TestFPTestSuite) TestReduce0() {
	max := func(i, j int) int {
		if i > j {
			return i
		}
		return j
	}
	min := func(i, j int) int {
		if i < j {
			return i
		}
		return j
	}
	sum := func(i, j int) int { return i + j }

	source := []int{1, 2, 3, 4, 5, 6, 7}
	ret := StreamOf(source).Reduce0(max).Result().(int)
	suite.Equal(int(7), ret)

	ret = StreamOf(source).Reduce0(min).Result().(int)
	suite.Equal(int(1), ret)

	ret = StreamOf(source).Reduce0(sum).Result().(int)
	suite.Equal(int(28), ret)
}

func (suite *TestFPTestSuite) TestReduce0Empty() {
	max := func(i, j int) int {
		if i > j {
			return i
		}
		return j
	}

	source := []int{1}
	ret := StreamOf(source).Reduce0(max).Int()
	suite.Equal(int(1), ret)

	source = []int{}
	ret = StreamOf(source).Reduce0(max).Int()
	suite.Equal(int(0), ret)
}

func (suite *TestFPTestSuite) TestReduceChan() {
	ch := make(chan string, 3)
	ch <- "a"
	ch <- "c"
	ch <- "c"
	close(ch)

	out := StreamOf(ch).Reduce(map[string]int{}, func(memo map[string]int, s string) map[string]int {
		memo[s] += 1
		return memo
	}).Result().(map[string]int)
	suite.Equal(map[string]int{
		"a": 1,
		"c": 2,
	}, out)
}

func (suite *TestFPTestSuite) TestPartition() {
	source := []string{"a", "b", "c", "d"}

	out := StreamOf(source).Partition(3).StringsList()
	suite.Equal([][]string{
		{"a", "b", "c"},
		{"d"},
	}, out)
}

func (suite *TestFPTestSuite) TestIsEmpty() {
	slice := []string{"abc", "de", "f"}
	q := StreamOf(slice)
	suite.False(q.IsEmpty())
	out := q.First()
	suite.Equal("abc", out.String())
}

func (suite *TestFPTestSuite) TestHasSomething() {
	slice := []string{"abc", "de", "f"}
	q := StreamOf(slice)
	suite.True(q.HasSomething())
	out := q.First()
	suite.Equal("abc", out.String())
}

func (suite *TestFPTestSuite) TestTake() {
	slice := []string{"abc", "de", "f"}
	out := strings.Join(StreamOf(slice).Take(2).Strings(), "")
	suite.Equal("abcde", out)

	out = strings.Join(StreamOf(slice).Take(20).Strings(), "")
	suite.Equal("abcdef", out)
}

func (suite *TestFPTestSuite) TestSkip() {
	slice := []string{"abc", "de", "f"}
	out := strings.Join(StreamOf(slice).Skip(2).Strings(), "")
	suite.Equal("f", out)

	out = strings.Join(StreamOf(slice).Skip(3).Strings(), "")
	suite.Equal("", out)

	out = strings.Join(StreamOf(slice).Skip(20).Strings(), "")
	suite.Equal("", out)
}

func (suite *TestFPTestSuite) TestPageRange() {
	slice := []string{"abc", "de", "f", "g", "i"}
	out := StreamOf(slice).Skip(1).Take(2).Strings()
	suite.Equal([]string{"de", "f"}, out)

	out = StreamOf(slice).Skip(0).Take(2).Strings()
	suite.Equal([]string{"abc", "de"}, out)

	out = StreamOf(slice).Skip(0).Take(0).Strings()
	suite.Nil(out)

	out = StreamOf(slice).SkipWhile(func(s string) bool {
		return s == ""
	}).Take(2).Strings()
	suite.Equal([]string{"abc", "de"}, out)

	out = StreamOf(slice).SkipWhile(func(s string) bool {
		return s == "abc"
	}).Take(2).Strings()
	suite.Equal([]string{"de", "f"}, out)

	out = StreamOf(slice).TakeWhile(func(s string) bool {
		return s == "abc"
	}).Take(2).Strings()
	suite.Equal([]string{"abc"}, out)
}

func (suite *TestFPTestSuite) TestSort() {
	slice := []int{1, 3, 2}
	out := StreamOf(slice).Sort().Ints()
	suite.Equal([]int{1, 2, 3}, out)
}

func (suite *TestFPTestSuite) TestSortBy() {
	slice := []string{"abc", "de", "f"}
	out := StreamOf(slice).SortBy(func(a, b string) bool {
		return len(a) < len(b)
	}).Strings()
	suite.Equal([]string{"f", "de", "abc"}, out)
}

func (suite *TestFPTestSuite) TestContains() {
	slice := []string{"abc", "de", "f"}
	q := StreamOf(slice)
	suite.True(q.Contains("de"))
	suite.False(q.Contains("e"))
	suite.Equal([]string{"ABC", "DE", "F"}, q.Map(strings.ToUpper).Strings())

	ptr := func(s string) *string { return &s }
	slice1 := []string{"abc", "de", "f"}
	q = StreamOf(slice1).Map(func(s string) *string { return &s })
	suite.True(q.Contains(ptr("de")))
	suite.False(q.Contains(ptr("e")))
	suite.Equal([]string{"ABC", "DE", "F"}, q.Map(func(s *string) string { return strings.ToUpper(*s) }).Strings())
}

func (suite *TestFPTestSuite) TestContainsBy() {
	slice := []string{"abc", "de", "f"}
	q := StreamOf(slice)
	suite.True(q.ContainsBy(func(s string) bool { return strings.ToUpper(s) == "F" }))
	suite.False(q.ContainsBy(func(s string) bool { return s == "e" }))
	suite.False(q.ContainsBy(func(s string) bool { return s == "F" }))
	suite.Equal([]string{"ABC", "DE", "F"}, q.Map(strings.ToUpper).Strings())
}

func (suite *TestFPTestSuite) TestUniq() {
	slice := []int{1, 3, 2, 1, 2, 1, 3}
	out := StreamOf(slice).Uniq().Ints()
	suite.ElementsMatch([]int{1, 2, 3}, out)
}

func (suite *TestFPTestSuite) TestUniqKeepFirst() {
	slice := []string{"a", "A", "B", "c", "b"}
	out := StreamOf(slice).UniqBy(func(s string) string { return strings.ToLower(s) }).Strings()
	suite.ElementsMatch([]string{"a", "B", "c"}, out)
}

func (suite *TestFPTestSuite) TestUniqBy() {
	slice := []int{1, 3, 2, 1, 2, 1, 3}
	out := StreamOf(slice).UniqBy(func(i int) bool {
		return i%2 == 0
	}).Ints()
	suite.ElementsMatch([]int{1, 2}, out)
}

func (suite *TestFPTestSuite) TestResult() {
	type S interface {
		String() string
	}
	var slice []S
	size := 2
	for i := 0; i < size; i++ {
		buf := bytes.NewBuffer(nil)
		buf.WriteString(fmt.Sprint(i))
		slice = append(slice, buf)
	}

	out := StreamOf(slice).Map(func(s S) string {
		return s.String()
	}).Strings()
	suite.Equal([]string{"0", "1"}, out)
}

func (suite *TestFPTestSuite) TestFlatten() {
	slice := []string{"abc", "de", "f"}
	out := StreamOf(slice).Map(func(s string) []byte {
		return []byte(s)
	}).Flatten().Bytes()
	suite.Equal("abcdef", string(out))
}

func (suite *TestFPTestSuite) TestFlatMap() {
	slice := []string{"abc", "de", "f"}
	out := StreamOf(slice).FlatMap(func(s string) []byte {
		return []byte(s)
	}).Bytes()
	suite.Equal("abcdef", string(out))
}

func (suite *TestFPTestSuite) TestDeepFlatten() {
	slice := [][]string{
		{"abc", "de", "f"},
		{"g", "hi"},
	}
	out := StreamOf(slice).Map(func(s []string) [][]byte {
		return StreamOf(s).Map(func(st string) []byte {
			return []byte(st)
		}).Result().([][]byte)
	}).Flatten().Flatten().Bytes()
	suite.Equal("abcdefghi", string(out))

	slice = [][]string{
		{"abc", "f"},
		{"g"},
	}
	out1 := StreamOf(slice).Flatten().Strings()
	suite.Equal([]string{"abc", "f", "g"}, out1)
}

func (suite *TestFPTestSuite) TestHybridFlatten() {
	slice := []chan string{
		make(chan string, 3),
		make(chan string, 3),
		make(chan string, 3),
		make(chan string, 3),
	}
	slice[0] <- "a"
	slice[1] <- "b"
	slice[1] <- "c"
	slice[2] <- "d"
	slice[2] <- "e"
	slice[3] <- "f"
	for _, ch := range slice {
		close(ch)
	}
	out := StreamOf(slice).
		Flatten().
		Strings()
	suite.Equal([]string{"a", "b", "c", "d", "e", "f"}, out)
}

func (suite *TestFPTestSuite) TestRepeatableGetValueOfHybridFlatten() {
	slice := []chan string{
		make(chan string, 3),
		make(chan string, 3),
		make(chan string, 3),
	}
	slice[0] <- "a"
	slice[1] <- "b"
	slice[1] <- "c"
	slice[2] <- "d"
	slice[2] <- "e"
	for _, ch := range slice {
		close(ch)
	}
	q := StreamOf(slice).Flatten()
	out := q.Strings()
	suite.Equal([]string{"a", "b", "c", "d", "e"}, out)

	out = q.Strings()
	suite.Equal([]string{"a", "b", "c", "d", "e"}, out)

}

func (suite *TestFPTestSuite) TestEmptyFlatten() {
	slice := []chan string{
		make(chan string, 3),
		make(chan string, 3),
		make(chan string, 3),
	}

	slice[1] <- "b"
	slice[1] <- "c"
	for _, ch := range slice {
		close(ch)
	}
	out := StreamOf(slice).
		Flatten().
		Strings()
	suite.Equal([]string{"b", "c"}, out)
}

func (suite *TestFPTestSuite) TestEmptyFlatten2() {
	slice := [][]string{
		{"a", "b"},
		nil,
		{},
		{"c"},
		nil,
		{},
		{""},
		{},
		nil,
	}

	out := StreamOf(slice).
		Flatten().
		Strings()
	suite.Equal([]string{"a", "b", "c", ""}, out)
}

func (suite *TestFPTestSuite) TestHybridComplexFlatten() {
	slice := []chan []byte{
		make(chan []byte, 3),
		make(chan []byte, 3),
		make(chan []byte, 3),
	}
	slice[0] <- []byte("a")
	slice[1] <- []byte("b")
	slice[1] <- []byte("c")
	slice[2] <- []byte("d")
	slice[2] <- []byte("e")
	for _, ch := range slice {
		close(ch)
	}
	out := StreamOf(slice).
		Flatten().
		Flatten().
		Bytes()
	suite.Equal("abcde", string(out))
}

func (suite *TestFPTestSuite) TestGetSize() {
	slice := []chan []byte{
		make(chan []byte, 3),
		make(chan []byte, 3),
		make(chan []byte, 3),
	}
	slice[0] <- []byte("a")
	slice[1] <- []byte("b")
	slice[1] <- []byte("c")
	slice[2] <- []byte("d")
	slice[2] <- []byte("e")
	for _, ch := range slice {
		close(ch)
	}
	q := StreamOf(slice).
		Flatten().
		Flatten()
	out := q.Size()
	suite.Equal(len("abcde"), out)
	// check again
	out = q.Size()
	suite.Equal(len("abcde"), out)
}

func (suite *TestFPTestSuite) TestJoinStream() {
	slice1 := []string{"abc", "de", "f"}
	q1 := StreamOf(slice1).Map(strings.ToUpper)
	slice2 := []string{"g", "hi"}
	q2 := StreamOf(slice2).Map(strings.ToUpper)
	out := q2.Union(q1).Strings()

	suite.Equal([]string{"G", "HI", "ABC", "DE", "F"}, out)
}

func (suite *TestFPTestSuite) TestJoinAfterNilStream() {
	slice1 := make(chan string, 1)
	close(slice1)
	q1 := StreamOf(slice1).Map(strings.ToUpper)
	slice2 := []string{"a", "b"}
	q2 := StreamOf(slice2).Map(strings.ToUpper)
	out := q2.Union(q1).Strings()

	suite.Equal([]string{"A", "B"}, out)
}

func (suite *TestFPTestSuite) TestGroupBy() {
	slice1 := []string{"abc", "de", "f", "gh"}
	q := StreamOf(slice1).Map(strings.ToUpper).GroupBy(func(s string) int {
		return len(s)
	}).Result().(map[int][]string)
	suite.Equal(map[int][]string{
		1: {"F"},
		2: {"DE", "GH"},
		3: {"ABC"},
	}, q)
}

func (suite *TestFPTestSuite) TestPrepend() {
	slice := []string{"abc", "de"}
	out := StreamOf(slice).Prepend("A").Strings()
	suite.Equal([]string{"A", "abc", "de"}, out)
}

func (suite *TestFPTestSuite) TestPrepend2() {
	slice := []string{"abc", "de"}
	out := StreamOf(slice).Prepend("A", "B").Strings()
	suite.Equal([]string{"A", "B", "abc", "de"}, out)
}
func (suite *TestFPTestSuite) TestAppend() {
	slice := []string{"abc", "de"}
	out := StreamOf(slice).Append("A").Strings()
	suite.Equal([]string{"abc", "de", "A"}, out)
}

func (suite *TestFPTestSuite) TestAppend2() {
	slice := []string{"abc", "de"}
	out := StreamOf(slice).Append("A", "B").Strings()
	suite.Equal([]string{"abc", "de", "A", "B"}, out)
}

func (suite *TestFPTestSuite) TestNilStream() {
	var slice []string
	out := StreamOf(slice).Append("a").Strings()
	suite.Equal([]string{"a"}, out)
}

func (suite *TestFPTestSuite) TestTakeWhile() {
	slice := []string{"a", "b", "c"}
	out := StreamOf(slice).TakeWhile(func(v string) bool {
		return v < "c"
	}).Strings()
	suite.Equal([]string{"a", "b"}, out)

	out = StreamOf(slice).TakeWhile(func(v string) bool {
		return v < "a"
	}).Strings()
	suite.Nil(out)
}

func (suite *TestFPTestSuite) TestSkipWhile() {
	slice := []string{"a", "b", "c"}
	out := StreamOf(slice).SkipWhile(func(v string) bool {
		return v < "c"
	}).Strings()
	suite.Equal([]string{"c"}, out)

	out = StreamOf(slice).SkipWhile(func(v string) bool {
		return v <= "c"
	}).Strings()
	suite.Nil(out)
}

func (suite *TestFPTestSuite) TestTakeWhileDropLeft() {
	slice := []string{"a", "b", "c", "d", "e"}
	var before, after []string
	out := StreamOf(slice).Foreach(func(s string) {
		before = append(before, s)
	}).TakeWhile(func(v string) bool {
		return v < "c"
	}).Foreach(func(s string) {
		after = append(after, s)
	}).Strings()
	suite.Equal([]string{"a", "b"}, out)
	suite.Equal([]string{"a", "b", "c"}, before)
	suite.Equal([]string{"a", "b"}, after)
}

func (suite *TestFPTestSuite) TestPartitionByAndIncludeSplittor() {
	slice := []string{"a", "b", "c", "d", "e", "c", "c"}
	out := StreamOf(slice).PartitionBy(func(s string) bool {
		return s == "c"
	}, true).StringsList()
	suite.Equal([][]string{
		{"a", "b", "c"},
		{"d", "e", "c"},
		{"c"},
	}, out)
}

func (suite *TestFPTestSuite) TestPartitionByAndExcludeSplittor() {
	slice := []string{"a", "b", "c", "d", "e", "c", "c"}
	out := StreamOf(slice).PartitionBy(func(s string) bool {
		return s == "c"
	}, false).StringsList()
	suite.Equal([][]string{
		{"a", "b"},
		{"d", "e"},
	}, out)
}

func (suite *TestFPTestSuite) TestCounter() {
	source := NewCounter(3)
	out := StreamOf(source).Ints()
	suite.Equal([]int{0, 1, 2}, out)
}

func (suite *TestFPTestSuite) TestCounterRange() {
	source := NewCounterRange(1, 3)
	out := StreamOf(source).Ints()
	suite.Equal([]int{1, 2, 3}, out)
}

func (suite *TestFPTestSuite) TestTickerSource() {
	source := NewTickerSource(time.Millisecond)
	defer source.Stop()
	now := time.Now()
	out := StreamOf(source).Take(3).Size()
	cost := time.Since(now).Milliseconds()
	suite.Equal(3, out)
	suite.Equal(int64(3), cost)
}

func (suite *TestFPTestSuite) TestSub() {
	slice1 := []int{1, 2, 3, 4}
	slice2 := []int{2, 1}
	out := StreamOf(slice1).Sub(StreamOf(slice2)).Ints()
	suite.Equal([]int{3, 4}, out)

	out = StreamOf(slice2).Sub(StreamOf(slice1)).Ints()
	suite.Nil(out)
}

func (suite *TestFPTestSuite) TestInteract() {
	slice1 := []int{1, 2, 3, 4}
	slice2 := []int{2, 1}
	out := StreamOf(slice1).Interact(StreamOf(slice2)).Ints()
	suite.ElementsMatch([]int{1, 2}, out)

	out = StreamOf(slice2).Interact(StreamOf(slice1)).Ints()
	suite.ElementsMatch([]int{1, 2}, out)
}

func (suite *TestFPTestSuite) TestLazyCollectionOp() {
	slice1 := []int{1, 2, 3, 4}
	slice2 := []int{2, 1}
	var count int
	q := StreamOf(slice1).
		Interact(StreamOf(slice2)).
		Union(StreamOf([]int{5, 6, 7})).
		Sub(StreamOf([]int{5})).
		Prepend(10).
		Foreach(func(int) { count++ })

	suite.Zero(count)
	out := q.Ints()
	suite.NotZero(count)
	suite.ElementsMatch([]int{1, 2, 6, 7, 10}, out)
}

func (suite *TestFPTestSuite) TestToSet() {
	slice := []int{1, 2, 3, 2, 1}
	out := StreamOf(slice).ToSet().Keys().Ints()
	suite.ElementsMatch([]int{1, 2, 3}, out)

	out1 := StreamOf(slice).ToSetBy(func(i int) string {
		return strconv.FormatInt(int64(i), 10)
	}).Keys().Strings()
	suite.ElementsMatch([]string{"1", "2", "3"}, out1)

	out = StreamOf(slice).ToSetBy(func(i int) string {
		return strconv.FormatInt(int64(i), 10)
	}).Values().Ints()
	suite.ElementsMatch([]int{1, 2, 3}, out)

}

func (suite *TestFPTestSuite) TestZip() {
	slice1 := []int{1, 2, 3}
	slice2 := []int{4, 5, 6, 7}
	out := StreamOf(slice1).Zip(StreamOf(slice2), func(i, j int) string {
		return strconv.FormatInt(int64(i+j), 10)
	}).Strings()
	suite.ElementsMatch([]string{"5", "7", "9"}, out)

	slice2 = nil
	out = StreamOf(slice1).Zip(StreamOf(slice2), func(i, j int) string {
		return strconv.FormatInt(int64(i+j), 10)
	}).Strings()
	suite.Nil(out)
}

func (suite *TestFPTestSuite) TestFullLazy() {
	var count int
	q := StreamOf([]int{1, 2, 3, 4}).Map(func(i int) int {
		count++
		return i
	}).FlatMap(func(i int) (int, bool) {
		count++
		return i, true
	}).ToSetBy(func(i int) int {
		count++
		return i
	}).Foreach(func(i, j int) {
		count++
	}).Filter(func(i, j int) bool {
		count++
		return true
	}).Values().Filter(func(i int) bool {
		count++
		return true
	}).Foreach(func(i int) {
		count++
	}).GroupBy(func(i int) int {
		count++
		return i
	}).Values().Flatten()
	suite.Zero(count)
	q.Result()
	suite.NotZero(count)
}

func (suite *TestFPTestSuite) TestStreamOfFunction() {
	var i int
	fn := func() (int, bool) {
		i++
		return i, i < 5
	}
	out := StreamOf(fn).Ints()
	suite.Equal([]int{1, 2, 3, 4}, out)

	i = 0
	fn1 := func() (interface{}, bool) {
		i++
		return i, i < 5
	}
	out1 := StreamOf(fn1).Result().([]interface{})
	suite.Equal([]interface{}{1, 2, 3, 4}, out1)
}

func (suite *TestFPTestSuite) TestJoinStrings() {
	slice := []string{"a", "b", "c"}
	out := StreamOf(slice).Map(strings.ToUpper).JoinStrings("|")

	suite.Equal("A|B|C", out)
}

func (suite *TestFPTestSuite) TestIndexNumber() {
	slice := []string{"a", "b", "c"}
	out := StreamOf(slice).Zip(Index(), func(s string, i int) string {
		return fmt.Sprintf("%v-%v", s, i)
	}).Strings()

	suite.Equal([]string{"a-0", "b-1", "c-2"}, out)
}

func (suite *TestFPTestSuite) TestFirstError() {
	slice := []string{"a", "b", "c"}
	out := StreamOf(slice).Map(func(s string) error {
		return errors.New(s)
	}).First().Err()

	suite.Equal(errors.New("a"), out)

	out = StreamOf([]string{""}).Map(func(s string) error {
		return nil
	}).First().Err()

	suite.NoError(out)
}
