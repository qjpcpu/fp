package fp

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MonadTestSuite struct {
	suite.Suite
}

func (suite *MonadTestSuite) SetupTest() {

}

func (suite *MonadTestSuite) TearDownTest() {

}

func TestMonadTestSuite(t *testing.T) {
	suite.Run(t, new(MonadTestSuite))
}

func (suite *MonadTestSuite) TestErrorMonadMap() {
	var v int64
	err := M("a").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).To(&v)
	suite.Zero(v)
	suite.Error(err)

	err = M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).To(&v)
	suite.Equal(int64(2), v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestErrorMonadWithConstructError() {
	cons := func() (string, error) {
		return "11", errors.New("err")
	}
	var v int64
	err := M(cons()).Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).To(&v)
	suite.Zero(v)
	suite.Equal("err", err.Error())

}

func (suite *MonadTestSuite) TestErrorMonadFlatMap() {
	var out []int
	err := M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).FlatMap(func(i int64) []int {
		return StreamOf(NewCounter(int(i))).Ints()
	}).ToSlice(&out)
	suite.NoError(err)
	suite.Equal([]int{0, 1}, out)
}

func (suite *MonadTestSuite) TestErrorMonadFlatMapStream() {
	var out []int
	err := M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).FlatMap(func(i int64) Stream {
		return StreamOf(NewCounter(int(i)))
	}).ToSlice(&out)
	suite.NoError(err)
	suite.Equal([]int{0, 1}, out)
}

func (suite *MonadTestSuite) TestErrorMonadFlatMapStreamWithError() {
	var out []int
	err := M("2").Map(func(s string) (int64, error) {
		return 0, errors.New("err")
	}).FlatMap(func(i int64) Stream {
		return StreamOf(NewCounter(int(i)))
	}).ToSlice(&out)
	suite.Error(err)
	suite.Zero(out)
}

func (suite *MonadTestSuite) TestErrorMonadFlatMapStreamWithOptional() {
	var out []int
	err := M("2").Map(func(s string) (int64, bool) {
		return 0, false
	}).FlatMap(func(i int64) Stream {
		return StreamOf(NewCounter(int(i)))
	}).ToSlice(&out)
	suite.NoError(err)
	suite.Zero(out)
}

func (suite *MonadTestSuite) TestErrorMonadFMWithConstructError() {
	cons := func() (string, error) {
		return "11", errors.New("err")
	}
	var v []int64
	err := M(cons()).FlatMap(func(s string) ([]int64, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return []int64{i}, err
	}).ToSlice(&v)
	suite.Zero(v)
	suite.Equal("err", err.Error())
}

func (suite *MonadTestSuite) TestErrorMonadFMWithError() {
	var v []int64
	err := M("2a").FlatMap(func(s string) ([]int64, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return []int64{i}, err
	}).ToSlice(&v)
	suite.Zero(v)
	suite.Error(err)
}

func (suite *MonadTestSuite) TestMayBeMonad() {
	var v int64
	err := M("2", false).Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).To(&v)
	suite.Zero(v)
	suite.NoError(err)

	err = M("2", true).Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).To(&v)
	suite.Equal(int64(2), v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMayBeMonadMap() {
	var v int64
	err := M("2").Map(func(s string) (int64, bool) {
		i, err := strconv.ParseInt(s, 10, 64)
		return i, err == nil
	}).To(&v)
	suite.Equal(int64(2), v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMayBeMonadFlatMap() {
	var v []int64
	err := M("2").FlatMap(func(s string) ([]int64, bool) {
		i, err := strconv.ParseInt(s, 10, 64)
		return []int64{i}, err == nil
	}).ToSlice(&v)
	suite.Equal([]int64{2}, v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMayBeMonadFlatMap2() {
	var v []int64
	err := M("2", false).FlatMap(func(s string) ([]int64, bool) {
		i, _ := strconv.ParseInt(s, 10, 64)
		return []int64{i}, true
	}).ToSlice(&v)
	suite.Len(v, 0)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMayBeMonadFlatMap3() {
	var v []int64
	err := M("2", true).FlatMap(func(s string) ([]int64, bool) {
		i, _ := strconv.ParseInt(s, 10, 64)
		return []int64{i}, false
	}).ToSlice(&v)
	suite.Len(v, 0)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMonadExpect() {
	var v int64
	err := M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Expect(func(i int64) bool {
		return i > 0
	}).To(&v)
	suite.Equal(int64(2), v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMonadExpect2() {
	var v int64
	err := M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Expect(func(i int64) error {
		return errors.New("xerr")
	}).To(&v)
	suite.Equal(int64(0), v)
	suite.Error(err)
}

func (suite *MonadTestSuite) TestMonadExpect21() {
	var v int64
	err := M("21a").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Expect(func(i int64) error {
		return nil
	}).To(&v)
	suite.Equal(int64(0), v)
	suite.Error(err)
}
func (suite *MonadTestSuite) TestMonadExpect22() {
	var v int64
	err := M("21a").Map(func(s string) (int64, bool) {
		i, err := strconv.ParseInt(s, 10, 64)
		return i, err == nil
	}).Expect(func(i int64) error {
		return nil
	}).To(&v)
	suite.Equal(int64(0), v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMonadExpect3() {
	suite.Panics(func() {
		var v int
		M("2").Map(func(s string) (int64, error) {
			return strconv.ParseInt(s, 10, 64)
		}).Expect(func(i int64) int {
			return 0
		}).To(&v)

	})
}

func (suite *MonadTestSuite) TestMonadExpect31() {
	var v int64
	err := M("21a").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Expect(func(i int64) bool {
		return true
	}).To(&v)
	suite.Equal(int64(0), v)
	suite.Error(err)
}
func (suite *MonadTestSuite) TestMonadExpect32() {
	var v int64
	err := M("21a").Map(func(s string) (int64, bool) {
		i, err := strconv.ParseInt(s, 10, 64)
		return i, err == nil
	}).Expect(func(i int64) bool {
		return true
	}).To(&v)
	suite.Equal(int64(0), v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMonadReturnError() {
	err := M("21a").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Error()
	suite.Error(err)
}

func (suite *MonadTestSuite) TestMonadReturnError1() {
	err := M("21a").Map(func(s string) error {
		_, err := strconv.ParseInt(s, 10, 64)
		return err
	}).Error()
	suite.Error(err)

	err = M("21a").Map(func(s string) error {
		_, _ = strconv.ParseInt(s, 10, 64)
		return nil
	}).Error()
	suite.NoError(err)
}
