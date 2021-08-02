package fp

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
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
	var out1 []string
	q.ToSlice(&out1)
	suite.Equal([]string{"abc", "de", "f"}, out1)
}

func (suite *TestFPTestSuite) TestHasSomething() {
	slice := []string{"abc", "de", "f"}
	q := StreamOf(slice)
	suite.True(q.HasSomething())
	suite.True(q.Exists())
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

func (suite *TestFPTestSuite) TestContains1() {
	suite.True(StreamOf([]int{1}).Contains(1))
	suite.True(StreamOf([]uint{1}).Contains(uint(1)))
	suite.True(StreamOf([]bool{false}).Contains(false))
	suite.True(StreamOf([]float64{1}).Contains(float64(1)))
}

func (suite *TestFPTestSuite) TestRun() {
	suite.NotPanics(func() {
		StreamOf([]int{1}).Run()
		_ = StreamOf([]int{1}).(Source).ElemType()
	})
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

func (suite *TestFPTestSuite) TestSubBy() {
	slice1 := []string{"a", "b", "c", "d"}
	slice2 := []string{"C", "D"}
	out := StreamOf(slice1).SubBy(StreamOf(slice2), func(elem string) string {
		return strings.ToLower(elem)
	}).Strings()
	suite.Equal([]string{"a", "b"}, out)
}

func (suite *TestFPTestSuite) TestInteract() {
	slice1 := []int{1, 2, 3, 4}
	slice2 := []int{2, 1}
	out := StreamOf(slice1).Interact(StreamOf(slice2)).Ints()
	suite.ElementsMatch([]int{1, 2}, out)

	out = StreamOf(slice2).Interact(StreamOf(slice1)).Ints()
	suite.ElementsMatch([]int{1, 2}, out)
}

func (suite *TestFPTestSuite) TestInteractBy() {
	slice1 := []string{"a", "b", "c", "d"}
	slice2 := []string{"C", "D"}
	out := StreamOf(slice1).InteractBy(StreamOf(slice2), strings.ToLower).Strings()
	suite.ElementsMatch([]string{"c", "d"}, out)
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

func (suite *TestFPTestSuite) TestZipN() {
	slice1 := []int{1, 2, 3}
	slice2 := []int{4, 5, 6, 7}
	slice3 := []int{2, 3}
	out := StreamOf(slice1).ZipN(func(i, j, k int) string {
		return strconv.FormatInt(int64(i+j+k), 10)
	}, StreamOf(slice2), StreamOf(slice3)).Strings()
	suite.ElementsMatch([]string{"7", "10"}, out)

	slice2 = nil
	out = StreamOf(slice1).ZipN(func(i, j, k int) string {
		return strconv.FormatInt(int64(i+j+k), 10)
	}, StreamOf(slice2), StreamOf(slice3)).Strings()
	suite.Nil(out)
}

func (suite *TestFPTestSuite) TestZipN1() {
	slice1 := []int{1, 2, 3}
	slice2 := []int{4, 5, 6, 7}
	slice3 := []int{2, 3}
	out := StreamOf(slice1).ZipN(func(i int) string {
		return strconv.FormatInt(int64(i), 10)
	}).Strings()
	suite.ElementsMatch([]string{"1", "2", "3"}, out)

	q := StreamOf(slice1).ZipN(func(i, j, k int) string {
		return strconv.FormatInt(int64(i+j+k), 10)
	}, StreamOf(slice2), StreamOf(slice3))
	suite.True(q.Contains("7"))
	suite.False(q.Contains("8"))
	suite.ElementsMatch([]string{"7", "10"}, q.Strings())
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

func (suite *TestFPTestSuite) TestFirstErrorPattern() {
	var count int
	slice := []string{"a", "b", "c"}
	err := StreamOf(slice).Map(func(s string) error {
		count++
		if s >= "b" {
			return errors.New(s)
		}
		return nil
	}).SkipWhile(NoError()).First().Err()
	suite.Equal(errors.New("b"), err)
	suite.Equal(2, count)
}

func (suite *TestFPTestSuite) TestInPlaceToSlice() {
	holder := struct {
		Slice []string
	}{
		Slice: []string{"a", "b"},
	}
	StreamOf(holder.Slice).Map(strings.ToUpper).ToSlice(&holder.Slice)
	suite.Equal([]string{"A", "B"}, holder.Slice)
}

func (suite *TestFPTestSuite) TestFirstToPtr() {
	strPtr := func(s string) *string { return &s }
	slice := []*string{strPtr("a"), strPtr("b"), strPtr("c")}
	var out *string
	StreamOf(slice).First().To(&out)
	suite.Equal(strPtr("a"), out)

	var out1 *string
	StreamOf(slice).Filter(func(s *string) bool { return false }).First().To(&out)
	suite.Nil(out1)

	var out2 string
	StreamOf([]string{}).First().To(&out2)
	suite.Equal("", out2)
}

func (suite *TestFPTestSuite) TestFirstToSuccess() {
	strPtr := func(s string) *string { return &s }
	slice := []*string{strPtr("a"), strPtr("b"), strPtr("c")}
	var out *string
	success := StreamOf(slice).First().To(&out)
	suite.Equal(strPtr("a"), out)
	suite.True(success)

	var out1 *string
	success = StreamOf(slice).Filter(func(s *string) bool { return false }).First().To(&out)
	suite.Nil(out1)
	suite.False(success)

	var out2 string
	success = StreamOf([]string{}).First().To(&out2)
	suite.Equal("", out2)
	suite.False(success)
}

func (suite *TestFPTestSuite) TestStream0() {
	get := func() ([]string, error) { return []string{"a", "b"}, nil }
	out := Stream0Of(get()).Strings()
	suite.Equal([]string{"a", "b"}, out)
}

func (suite *TestFPTestSuite) TestEqual() {
	slice := []int{1, 2, 3, 4}

	out := StreamOf(slice).Filter(Equal(3)).Ints()
	suite.Equal([]int{3}, out)

	slice1 := []string{"a", "b"}
	out1 := StreamOf(slice1).Filter(Equal("b")).Strings()
	suite.Equal([]string{"b"}, out1)
}

func (suite *TestFPTestSuite) TestEqualIgnoreCase() {
	slice1 := []string{"a", "b"}
	out1 := StreamOf(slice1).Filter(EqualIgnoreCase("B")).Strings()
	suite.Equal([]string{"b"}, out1)
}

func (suite *TestFPTestSuite) TestEmptyString() {
	slice1 := []string{"a", "b", " ", "", "\t"}
	out1 := StreamOf(slice1).Reject(EmptyString()).Strings()
	suite.Equal([]string{"a", "b"}, out1)
}

func (suite *TestFPTestSuite) TestMulti0() {
	slice := []string{"a", "b", "c", "d"}
	StreamOf(slice).Branch()
	StreamOf(slice).Branch(func(s Stream) {
		suite.Equal(slice, s.Strings())
	})
}

func (suite *TestFPTestSuite) TestMulti() {
	slice := []string{"a", "b", "c", "d"}
	var out1, out2 []string
	var out3 string
	StreamOf(slice).Reject(Equal("d")).Branch(func(stream Stream) {
		stream.Map(strings.ToUpper).ToSlice(&out1)
	}, func(stream Stream) {
		out3 = stream.First().String()
	}, func(stream Stream) {
		stream.Skip(1).Take(1).ToSlice(&out2)
	})

	suite.Equal([]string{"A", "B", "C"}, out1)
	suite.Equal("a", out3)
	suite.Equal([]string{"b"}, out2)
}

func (suite *TestFPTestSuite) TestMulti2() {
	slice := []string{"a", "b", "c", "d"}
	var out2 []string
	var out3 string
	StreamOf(slice).Reject(Equal("d")).Branch(func(stream Stream) {

	}, func(stream Stream) {
		out3 = stream.First().String()
	}, func(stream Stream) {
		stream.Skip(1).Take(1).ToSlice(&out2)
	})

	suite.Equal("a", out3)
	suite.Equal([]string{"b"}, out2)
}

func (suite *TestFPTestSuite) TestZipnFuncCheck() {
	slice := []string{"a", "b", "c", "d"}
	suite.Panics(func() {
		StreamOf(slice).ZipN(func(string) string { return "" }, StreamOf([]string{"a"}))
	})
}

func (suite *TestFPTestSuite) TestStreamMustHaveIterator() {
	s := newStream(reflect.TypeOf(1), nil)
	suite.NotNil(s.iter)
	_, v := s.iter()
	suite.False(v)
}

func (suite *TestFPTestSuite) TestMustFlattenSlice() {
	suite.Panics(func() {
		StreamOf([]string{}).Flatten()
	})
}

func (suite *TestFPTestSuite) TestPartitionSize() {
	suite.Panics(func() {
		StreamOf([]string{}).Partition(0)
	})
}

func (suite *TestFPTestSuite) TestToSliceMustBePtr() {
	suite.Panics(func() {
		StreamOf([]string{}).ToSlice([]string{})
	})
}

func (suite *TestFPTestSuite) TestMakeIter() {
	suite.Panics(func() {
		makeIter(reflect.ValueOf(1))
	})
}

func (suite *TestFPTestSuite) TestNaturalNumbers() {
	suite.Equal(uint64(0), NaturalNumbers().First().Uint64())
}

func (suite *TestFPTestSuite) TestMaxNaturalNumbers() {
	s := &naturalNumSource{i: math.MaxUint64}
	suite.Equal(0, StreamOf(s).Count())
}

type _kvdemo struct{ i int }

func (d *_kvdemo) ElemType() (reflect.Type, reflect.Type) {
	return reflect.TypeOf(0), reflect.TypeOf(0)
}
func (d *_kvdemo) Next() (reflect.Value, reflect.Value, bool) {
	if d.i > 0 {
		i := d.i
		d.i--
		return reflect.ValueOf(i), reflect.ValueOf(i), true
	}
	return reflect.Value{}, reflect.Value{}, false
}

func (suite *TestFPTestSuite) TestKVstreamSource() {
	suite.Equal([]int{1, 2}, KVStreamOf(&_kvdemo{i: 2}).Keys().Sort().Ints())
	suite.Equal(2, KVStreamOf(&_kvdemo{i: 2}).Size())
}

func (suite *TestFPTestSuite) TestKVstreamConvertiableContain() {
	suite.True(KVStreamOf(&_kvdemo{i: 2}).Contains(uint64(1)))
}
func (suite *TestFPTestSuite) TestKvStreamMustBeMap() {
	suite.Panics(func() {
		KVStreamOf([]string{})
	})
}

func (suite *TestFPTestSuite) TestCompare() {
	suite.Equal(-1, StreamOf([]string{}).(*stream).compare(reflect.ValueOf("a"), reflect.ValueOf("b")))
	suite.Equal(1, StreamOf([]string{}).(*stream).compare(reflect.ValueOf("b"), reflect.ValueOf("a")))
	suite.Equal(0, StreamOf([]string{}).(*stream).compare(reflect.ValueOf("b"), reflect.ValueOf("b")))

	suite.Equal(-1, StreamOf([]uint64{}).(*stream).compare(reflect.ValueOf(uint64(1)), reflect.ValueOf(uint64(2))))
	suite.Equal(1, StreamOf([]uint64{}).(*stream).compare(reflect.ValueOf(uint64(3)), reflect.ValueOf(uint64(2))))

	suite.Equal(-1, StreamOf([]bool{}).(*stream).compare(reflect.ValueOf(false), reflect.ValueOf(true)))
	suite.Equal(1, StreamOf([]bool{}).(*stream).compare(reflect.ValueOf(true), reflect.ValueOf(false)))

	suite.Equal(-1, StreamOf([]TupleStringInt{}).(*stream).compare(reflect.ValueOf(TupleStringInt{E2: 1}), reflect.ValueOf(TupleStringInt{E2: 2})))
	suite.Equal(1, StreamOf([]TupleStringInt{}).(*stream).compare(reflect.ValueOf(TupleStringInt{E2: 3}), reflect.ValueOf(TupleStringInt{E2: 2})))
}

func (suite *TestFPTestSuite) TestToXXX() {
	suite.Equal([]int64{1}, StreamOf([]int64{1}).Int64s())
	suite.Equal([]int32{1}, StreamOf([]int32{1}).Int32s())
	suite.Equal([]uint{1}, StreamOf([]uint{1}).Uints())
	suite.Equal([]uint32{1}, StreamOf([]uint32{1}).Uint32s())
	suite.Equal([]uint64{1}, StreamOf([]uint64{1}).Uint64s())
	suite.Equal([]float64{1}, StreamOf([]float64{1}).Float64s())

	suite.Equal(int64(1), StreamOf([]int64{1}).First().Int64())
	suite.Equal(int32(1), StreamOf([]int32{1}).First().Int32())
	suite.Equal(uint32(1), StreamOf([]uint32{1}).First().Uint32())
	suite.Equal(float64(1), StreamOf([]float64{1}).First().Float64())

	suite.Panics(func() {
		StreamOf([]int{1}).First().To(1)
	})
}

func (suite *TestFPTestSuite) TestInvalidValue() {
	suite.Nil(Value{}.Result())
	suite.Nil(Value{}.Err())
}

func (suite *TestFPTestSuite) TestRepeatableIter() {
	suite.Nil(repeatableIter(nil, nil))
}

func (suite *TestFPTestSuite) TestTickerSourceFinish() {
	ds := NewDelaySource(time.Microsecond)
	suite.Equal(reflect.TypeOf(time.Time{}), ds.ElemType())
	_, ok := ds.Next()
	suite.True(ok)
}

func (suite *TestFPTestSuite) TestTuples() {
	suite.NotNil(TupleOf(1, 2))
	suite.Equal("a", TupleStringOf("a", "b").E1)
	suite.Equal("a", TupleStringAnyOf("a", "b").E1)
	suite.Equal("a", TupleStringIntOf("a", 1).E1)
	suite.Equal("a", TupleStringStringsOf("a", []string{"b"}).E1)
	suite.Equal("a", TupleStringTypeOf("a", reflect.TypeOf([]string{"b"})).E1)
	suite.Equal(1, TupleIntTypeOf(1, reflect.TypeOf([]string{"b"})).E1)
	suite.Equal(errors.New("err"), TuppleWithError(1, errors.New("err")).E2)
}

func (suite *TestFPTestSuite) TestReduceHelper() {
	suite.Equal("a", ShorterString("a", "bb"))
	suite.Equal("b", ShorterString("aa", "b"))
	suite.Equal("bb", LongerString("a", "bb"))
	suite.Equal("aa", LongerString("aa", "b"))

	suite.Equal(2, MaxInt(1, 2))
	suite.Equal(2, MaxInt(2, 0))

	suite.Equal(int32(2), MaxInt32(1, 2))
	suite.Equal(int32(2), MaxInt32(2, 0))

	suite.Equal(int8(2), MaxInt8(1, 2))
	suite.Equal(int8(2), MaxInt8(2, 0))

	suite.Equal(int16(2), MaxInt16(1, 2))
	suite.Equal(int16(2), MaxInt16(2, 0))

	suite.Equal(int64(2), MaxInt64(1, 2))
	suite.Equal(int64(2), MaxInt64(2, 0))

	suite.Equal(uint32(2), MaxUint32(1, 2))
	suite.Equal(uint32(2), MaxUint32(2, 0))

	suite.Equal(uint8(2), MaxUint8(1, 2))
	suite.Equal(uint8(2), MaxUint8(2, 0))

	suite.Equal(uint64(2), MaxUint64(1, 2))
	suite.Equal(uint64(2), MaxUint64(2, 0))

	suite.Equal(uint16(2), MaxUint16(1, 2))
	suite.Equal(uint16(2), MaxUint16(2, 0))

	suite.Equal(1, MinInt(2, 1))
	suite.Equal(0, MinInt(0, 2))

	suite.Equal(int32(1), MinInt32(2, 1))
	suite.Equal(int32(0), MinInt32(0, 2))

	suite.Equal(int8(1), MinInt8(2, 1))
	suite.Equal(int8(0), MinInt8(0, 2))

	suite.Equal(int16(1), MinInt16(2, 1))
	suite.Equal(int16(0), MinInt16(0, 2))

	suite.Equal(int64(1), MinInt64(2, 1))
	suite.Equal(int64(0), MinInt64(0, 2))

	suite.Equal(uint32(1), MinUint32(2, 1))
	suite.Equal(uint32(0), MinUint32(0, 2))

	suite.Equal(uint8(1), MinUint8(2, 1))
	suite.Equal(uint8(0), MinUint8(0, 2))

	suite.Equal(uint64(1), MinUint64(2, 1))
	suite.Equal(uint64(0), MinUint64(0, 2))

	suite.Equal(uint16(1), MinUint16(2, 1))
	suite.Equal(uint16(0), MinUint16(0, 2))
}
