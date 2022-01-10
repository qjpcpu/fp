package fp

import (
	"errors"
	"net"
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

func (suite *MonadTestSuite) TestErrorMonadStreamOf() {
	var out []int
	err := M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).StreamOf(func(i int64) []int {
		return StreamOf(NewCounter(int(i))).Ints()
	}).ToSlice(&out)
	suite.NoError(err)
	suite.Equal([]int{0, 1}, out)
}

func (suite *MonadTestSuite) TestErrorMonadStreamOfStream() {
	var out []int
	err := M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).StreamOf(func(i int64) Stream {
		return StreamOf(NewCounter(int(i)))
	}).ToSlice(&out)
	suite.NoError(err)
	suite.Equal([]int{0, 1}, out)
}

func (suite *MonadTestSuite) TestErrorMonadStreamOfStreamWithError() {
	var out []int
	err := M("2").Map(func(s string) (int64, error) {
		return 0, errors.New("err")
	}).StreamOf(func(i int64) Stream {
		return StreamOf(NewCounter(int(i)))
	}).ToSlice(&out)
	suite.Error(err)
	suite.Zero(out)
}

func (suite *MonadTestSuite) TestErrorMonadStreamOfStreamWithOptional() {
	var out []int
	err := M("2").Map(func(s string) (int64, bool) {
		return 0, false
	}).StreamOf(func(i int64) Stream {
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
	err := M(cons()).StreamOf(func(s string) ([]int64, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return []int64{i}, err
	}).ToSlice(&v)
	suite.Zero(v)
	suite.Equal("err", err.Error())
}

func (suite *MonadTestSuite) TestErrorMonadFMWithError() {
	var v []int64
	err := M("2a").StreamOf(func(s string) ([]int64, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return []int64{i}, err
	}).ToSlice(&v)
	suite.Zero(v)
	suite.Error(err)
}

func (suite *MonadTestSuite) TestErrorMonadFMWithError1() {
	var v []int64
	err := M("2a").StreamOf(func(s string) (Stream, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return StreamOf([]int64{i}), err
	}).ToSlice(&v)
	suite.Zero(v)
	suite.Error(err)

	err = M("2").StreamOf(func(s string) (Stream, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return StreamOf([]int64{i}), err
	}).ToSlice(&v)
	suite.ElementsMatch([]int64{2}, v)
	suite.Nil(err)
}

func (suite *MonadTestSuite) TestErrorMonadFMWithError2() {
	var v []int64
	err := M("2a").StreamOf(func(s string) Stream {
		i, err0 := strconv.ParseInt(s, 10, 64)
		return Stream0Of([]int64{i}, err0)
	}).ToSlice(&v)
	suite.Zero(v)
	suite.Error(err)

	err = M("2").StreamOf(func(s string) Stream {
		i, err0 := strconv.ParseInt(s, 10, 64)
		return Stream0Of([]int64{i}, err0)
	}).ToSlice(&v)
	suite.ElementsMatch([]int64{2}, v)
	suite.Nil(err)
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

func (suite *MonadTestSuite) TestMayBeMonadStreamOf() {
	var v []int64
	err := M("2").StreamOf(func(s string) ([]int64, bool) {
		i, err := strconv.ParseInt(s, 10, 64)
		return []int64{i}, err == nil
	}).ToSlice(&v)
	suite.Equal([]int64{2}, v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMayBeMonadStreamOf2() {
	var v []int64
	err := M("2", false).StreamOf(func(s string) ([]int64, bool) {
		i, _ := strconv.ParseInt(s, 10, 64)
		return []int64{i}, true
	}).ToSlice(&v)
	suite.Len(v, 0)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMayBeMonadStreamOf3() {
	var v []int64
	err := M("2", true).StreamOf(func(s string) ([]int64, bool) {
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
	}).ExpectPass(func(i int64) bool {
		return i > 0
	}).To(&v)
	suite.Equal(int64(2), v)
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMonadExpect2() {
	var v int64
	err := M("2").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).ExpectNoError(func(i int64) error {
		return errors.New("xerr")
	}).To(&v)
	suite.Equal(int64(0), v)
	suite.Error(err)
}

func (suite *MonadTestSuite) TestMonadExpect21() {
	var v int64
	err := M("21a").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).ExpectNoError(func(i int64) error {
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
	}).ExpectNoError(func(i int64) error {
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
		}).ExpectPass(func(i int64) int {
			return 0
		}).To(&v)

	})
}

func (suite *MonadTestSuite) TestMonadExpect31() {
	var v int64
	err := M("21a").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).ExpectPass(func(i int64) bool {
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
	}).ExpectPass(func(i int64) bool {
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

func (suite *MonadTestSuite) TestMonadCombine() {
	m1 := M("20").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})

	var score int64
	err := M("10").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Zip(func(a, b int64) int64 {
		return a + b
	}, m1).To(&score)
	suite.NoError(err)
	suite.Equal(int64(30), score)
}

func (suite *MonadTestSuite) TestMonadCombineFailedByAnyFailed() {
	m1 := M("20a").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})

	var score int64
	err := M("10").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Zip(func(a, b int64) int64 {
		return a + b
	}, m1).To(&score)
	suite.Error(err)
	suite.Equal(int64(0), score)
}

func (suite *MonadTestSuite) TestMonadCombineFailedByAnyFalse() {
	m1 := M("20a").Map(func(s string) (int64, bool) {
		i, err := strconv.ParseInt(s, 10, 64)
		return i, err == nil
	})

	var score int64
	err := M("10").Map(func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	}).Zip(func(a, b int64) int64 {
		return a + b
	}, m1).To(&score)
	suite.NoError(err)
	suite.Equal(int64(0), score)
}

func (suite *MonadTestSuite) TestMonadInvokeOnce() {
	var cnt int
	var score int64
	m := M("10").Map(func(s string) (int64, error) {
		cnt++
		return strconv.ParseInt(s, 10, 64)
	})
	err := m.To(&score)
	suite.NoError(err)
	suite.Equal(int64(10), score)
	suite.Equal(1, cnt)
	err = m.To(&score)
	suite.NoError(err)
	suite.Equal(int64(10), score)
	suite.Equal(2, cnt)

	cnt = 0
	score = 0
	m = M("10").Map(func(s string) (int64, error) {
		cnt++
		return strconv.ParseInt(s, 10, 64)
	}).Once()
	err = m.To(&score)
	suite.NoError(err)
	suite.Equal(int64(10), score)
	suite.Equal(1, cnt)
	err = m.To(&score)
	suite.NoError(err)
	suite.Equal(int64(10), score)
	suite.Equal(1, cnt)
}

func (suite *MonadTestSuite) TestMonadNilType() {
	f := func() net.Conn { return nil }

	err := M(f()).ExpectPass(func(c net.Conn) bool {
		return c != nil
	}).Error()
	suite.NoError(err)
}

func (suite *MonadTestSuite) TestMonadNilType2() {
	f := func() (net.Conn, error) { return nil, errors.New("error") }

	err := M(f()).ExpectPass(func(c net.Conn) bool {
		return c != nil
	}).Error()
	suite.Error(err)
}

func (suite *MonadTestSuite) TestMonadNilCover() {
	f := func() net.Conn { return nil }
	M(f()).Map(nil).ExpectPass(nil).ExpectNoError(nil).StreamOf(nil)
	M(f()).Zip(nil).Once().fnContainer()()
	M(f()).To(1)
}

func (suite *MonadTestSuite) TestCrossNil() {
	f := func() (net.Conn, error) { return nil, errors.New("error") }
	var num int
	err := M(12).Zip(func(int, net.Conn) int { return 100 }, M(f())).To(&num)
	suite.Error(err)
	suite.Zero(num)

	err = M(f()).Zip(func(net.Conn, int) int { return 100 }, M(12)).To(&num)
	suite.Error(err)
	suite.Zero(num)
}

type E1 struct{ S string }

func (suite *MonadTestSuite) TestNilDataCantContinue() {
	var q *E1
	M(q).ExpectPass(func(s *E1) bool {
		panic("should not occur")
	}).Error()
}

func (suite *MonadTestSuite) TestNilReturnDataCantContinue() {
	f := func() *E1 { return nil }
	M(f()).ExpectPass(func(s *E1) bool {
		panic("should not occur")
	}).Error()
}

func (suite *MonadTestSuite) TestMapBool() {
	var b bool
	M(1).Map(func(int) bool {
		return true
	}).To(&b)
	suite.True(b)

	var b1 bool
	M(1).Map(func(int) bool {
		return false
	}).To(&b1)
	suite.False(b1)
}

func (suite *MonadTestSuite) TestOnceWithErr() {
	f := func() (int, error) { return 0, errors.New("error") }
	m1 := M(f()).Once()
	err := M(10).Zip(func(a, b int) int { return a + b }, m1).Error()
	suite.Error(err)

	err = m1.Zip(func(a, b int) int { return a + b }, M(10)).Error()
	suite.Error(err)

	err = M(12).Zip(func(a, b int) int { return a + b }, m1).Error()
	suite.Error(err)

}
