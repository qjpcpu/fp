package fp

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type KVStreamTestSuite struct {
	suite.Suite
}

func (suite *KVStreamTestSuite) SetupTest() {
}

func TestKVStreamTestSuite(t *testing.T) {
	suite.Run(t, new(KVStreamTestSuite))
}

func (suite *KVStreamTestSuite) TestForeach() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	var keys []string
	var vals []int
	KVStreamOf(m).Foreach(func(key string, val int) {
		keys = append(keys, key)
		vals = append(vals, val)
	}).Run()
	suite.ElementsMatch([]string{"a", "b"}, keys)
	suite.ElementsMatch([]int{1, 2}, vals)
}

func (suite *KVStreamTestSuite) TestKeys() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	var keys []string
	var vals []int
	KVStreamOf(m).Keys().Filter(func(s string) bool { return s == "a" }).ToSlice(&keys)
	suite.ElementsMatch([]string{"a"}, keys)
	KVStreamOf(m).Values().Filter(func(v int) bool { return v == 2 }).ToSlice(&vals)
	suite.ElementsMatch([]int{2}, vals)
}

func (suite *KVStreamTestSuite) TestMap() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	var vk map[int]string
	KVStreamOf(m).Map(func(k string, v int) (int, string) {
		return v, k
	}).To(&vk)
	suite.Equal("a", vk[1])
	suite.Equal("b", vk[2])
}

func (suite *KVStreamTestSuite) TestFlatMap() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	vk := KVStreamOf(m).ZipMap(func(k string, v int) string {
		return fmt.Sprintf("%v-%v", k, v)
	}).Strings()
	suite.ElementsMatch([]string{"a-1", "b-2"}, vk)
}

func (suite *KVStreamTestSuite) TestFilter() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	var b []int
	KVStreamOf(m).Filter(func(k string, v int) bool {
		return v == 1
	}).Values().ToSlice(&b)
	suite.ElementsMatch(
		[]int{1},
		b,
	)
	KVStreamOf(m).Reject(func(k string, v int) bool {
		return v == 2
	}).Values().ToSlice(&b)
	suite.ElementsMatch(
		[]int{1},
		b,
	)
}

func (suite *KVStreamTestSuite) TestTo() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	var mp map[string]int
	KVStreamOf(m).Filter(func(k string, v int) bool {
		return v == 1
	}).To(&mp)
	suite.Equal(map[string]int{"a": 1}, mp)
}

func (suite *KVStreamTestSuite) TestToNil() {
	var m map[string]int
	var mp map[string]int
	KVStreamOf(m).To(&mp)
	suite.Equal(map[string]int{}, mp)
	suite.NotNil(mp)
}

func (suite *KVStreamTestSuite) TestToCantBeNil() {
	var mp map[string]int
	newNilStream().ToSet().To(&mp)
	suite.Equal(map[string]int{}, mp)
	suite.NotNil(mp)
}

func (suite *KVStreamTestSuite) TestToNilStream() {
	var out []string
	newNilKVStream().Values().Filter(func(s string) bool { return true }).ToSlice(&out)
	suite.Len(out, 0)
}

func (suite *KVStreamTestSuite) TestLazy() {
	src := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	var cnt int
	q := KVStreamOf(src).Reject(func(k string, v int) bool {
		return k == "c"
	}).Filter(func(k string, v int) bool {
		return k == "a"
	}).Foreach(func(k string, v int) {
		cnt++
	})
	suite.Zero(cnt)
	q.Run()
	suite.Equal(1, cnt)
}

type MyMap map[string]string

func (suite *KVStreamTestSuite) TestNilInputMap() {
	var mm MyMap
	var ret map[string]string
	KVStreamOf(mm).To(&ret)
	suite.NotNil(ret)

	var mm1 map[string]string
	var ret1 map[string]string
	KVStreamOf(mm1).To(&ret1)
	suite.NotNil(ret1)
}

func (suite *KVStreamTestSuite) TestTypedMap() {
	mm := MyMap{"a": "b"}
	var ret map[string]string
	KVStreamOf(mm).Map(func(k, v string) (string, string) {
		return strings.ToUpper(k), strings.ToUpper(v)
	}).To(&ret)
	suite.Equal(map[string]string{"A": "B"}, ret)

	mm1 := MyMap{"a": "b"}
	var ret1 MyMap
	KVStreamOf(mm1).Map(func(k, v string) (string, string) {
		return strings.ToUpper(k), strings.ToUpper(v)
	}).To(&ret1)
	suite.Equal(MyMap(map[string]string{"A": "B"}), ret1)

}

func (suite *KVStreamTestSuite) TestResultMapAlwayNonNil() {
	/* with error */
	ctx := newCtx(nil)
	ctx.SetErr(errors.New(""))
	s := newKvStream(ctx, boolType, boolType, func() reflect.Value { return reflect.Value{} })
	var mp map[bool]bool
	err := s.To(&mp)
	suite.Len(mp, 0)
	suite.NotNil(mp)
	suite.Error(err)

	/* with error 2 */
	var mp0 map[int]int
	err = StreamOf([]string{"1", "a"}).
		Map(strconv.Atoi).
		ToSetBy(func(i int) int { return i }).
		To(&mp0)
	suite.NotNil(mp0)
	suite.Error(err)

	/* nil stream */
	var mp1 map[string]string
	err = newNilKVStream().To(&mp1)
	suite.Nil(err)
	suite.NotNil(mp1)

	/* empty stream */
	var mp2 map[string]string
	err = KVStreamOf(map[string]string{}).To(&mp2)
	suite.Nil(err)
	suite.NotNil(mp2)
}

func (suite *KVStreamTestSuite) TestWithMapError() {
	var mp map[string]int64
	err := StreamOf([]string{"1", "2"}).ToSetBy(func(s string) (string, int64) {
		i, _ := strconv.ParseInt(s, 10, 64)
		return s, i
	}).Map(func(k string, v int64) (string, int64, error) {
		if v == 2 {
			return k, v, errors.New("error")
		}
		return k, v, nil
	}).To(&mp)
	suite.Error(err)
}

func (suite *KVStreamTestSuite) TestWithFlatMapError() {
	var list []int64
	err := StreamOf([]string{"1", "2"}).ToSetBy(func(s string) (string, int64) {
		i, _ := strconv.ParseInt(s, 10, 64)
		return s, i
	}).ZipMap(func(k string, v int64) (int64, error) {
		if v == 2 {
			return v, errors.New("error")
		}
		return v, nil
	}).ToSlice(&list)
	suite.Error(err)
}
