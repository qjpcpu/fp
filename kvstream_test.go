package fp

import (
	"fmt"
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
	})
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
	vk := KVStreamOf(m).Map(func(k string, v int) (int, string) {
		return v, k
	}).Result().(map[int]string)
	suite.Equal("a", vk[1])
	suite.Equal("b", vk[2])
}

func (suite *KVStreamTestSuite) TestMapValue() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	vk := KVStreamOf(m).MapValue(func(k string, v int) string {
		return k
	}).Result().(map[string]string)
	suite.Equal("a", vk["a"])
	suite.Equal("b", vk["b"])

	vk = KVStreamOf(m).MapValue(func(v int) string {
		return fmt.Sprint(v)
	}).Result().(map[string]string)
	suite.Equal("1", vk["a"])
	suite.Equal("2", vk["b"])
}

func (suite *KVStreamTestSuite) TestMapKey() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	vk := KVStreamOf(m).MapKey(func(k string) string {
		return strings.ToUpper(k)
	}).Result().(map[string]int)
	suite.Equal(1, vk["A"])
	suite.Equal(2, vk["B"])

	vk = KVStreamOf(m).MapKey(func(k string, v int) string {
		return strings.ToUpper(k)
	}).Result().(map[string]int)
	suite.Equal(1, vk["A"])
	suite.Equal(2, vk["B"])
}

func (suite *KVStreamTestSuite) TestFilter() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	suite.ElementsMatch(
		[]int{1},
		KVStreamOf(m).Filter(func(k string, v int) bool {
			return v == 1
		}).Values().Result().([]int),
	)
	suite.ElementsMatch(
		[]int{1},
		KVStreamOf(m).Reject(func(k string, v int) bool {
			return v == 2
		}).Values().Result().([]int),
	)
}
