package fp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FPIfTestSuite struct {
	suite.Suite
}

func (suite *FPIfTestSuite) SetupTest() {

}

func (suite *FPIfTestSuite) TearDownTest() {

}

func TestFPIfTestSuite(t *testing.T) {
	suite.Run(t, new(FPIfTestSuite))
}

func (suite *FPIfTestSuite) TestCondition() {
	fn := And(func(i int) bool {
		return 1 < i
	}, func(i int) bool {
		return i < 5
	}).Out().(func(int) bool)
	suite.True(fn(3))
	suite.False(fn(6))

	fn = And(func(i int) bool {
		return 1 < i
	}, func(i int) bool {
		return i < 5
	}).Or(func(i int) bool {
		return i == -1
	}).Out().(func(int) bool)

	suite.True(fn(3))
	suite.False(fn(6))
	suite.True(fn(-1))
}

func (suite *FPIfTestSuite) TestConditionCompose() {
	inner := Or(func(i int) bool {
		return i == -1
	}, func(i int) bool {
		return i == -2
	})
	fn := And(func(i int) bool {
		return 1 < i
	}, func(i int) bool {
		return i < 5
	}).Or(inner).Out().(func(int) bool)
	suite.True(fn(3))
	suite.False(fn(6))
	suite.True(fn(-1))
	suite.True(fn(-2))
}

func (suite *FPIfTestSuite) TestConditionCompose1() {
	inner := Or(func(i int) bool {
		return i == -1
	}, func(i int) bool {
		return i == -2
	})
	fn := And(func(i int) bool {
		return 1 < i
	}, func(i int) bool {
		return i < 5
	}).And(inner).Out().(func(int) bool)
	suite.False(fn(3))
	suite.False(fn(6))
	suite.False(fn(-1))
	suite.False(fn(-2))
}

func (suite *FPIfTestSuite) TestConditionCompose2() {
	inner := Or(func(i int) bool {
		return i == -1
	}, func(i int) bool {
		return i == -2
	})
	fn := And(inner, func(i int) bool {
		return i < 0
	}).And(inner).Out().(func(int) bool)
	suite.False(fn(3))
	suite.True(fn(-1))
	suite.True(fn(-2))

	fn = And(func(i int) bool {
		return i < 0
	}, inner).And(inner).Out().(func(int) bool)
	suite.False(fn(3))
	suite.True(fn(-1))
	suite.True(fn(-2))
}

func (suite *FPIfTestSuite) TestConditionCompose3() {
	inner := And(func(i int) bool {
		return i > 1
	}, func(i int) bool {
		return i < 5
	})
	var fn func(int) bool
	Or(inner, func(i int) bool {
		return i < 0
	}).To(&fn)
	suite.True(fn(3))
	suite.True(fn(-1))
	suite.True(fn(-2))
}

func (suite *FPIfTestSuite) TestIf() {
	fn := When(func(i, j int) bool {
		return i == 1
	}).Then(func(i, j int) string {
		return fmt.Sprint(i)
	}).When(func(i, j int) bool {
		return i == 100
	}).Then(func(i, j int) string {
		return "large"
	}).Else(func(i, j int) string {
		return "0"
	}).(func(int, int) string)

	suite.Equal("1", fn(1, 0))
	suite.Equal("0", fn(2, 0))
	suite.Equal("large", fn(100, 0))
}

func (suite *FPIfTestSuite) TestIfMany() {
	fn := When(func(i int) bool {
		return i == 1
	}).Then(func(i int) (string, int) {
		return fmt.Sprint(i), i
	}).When(func(i int) bool {
		return i == 100
	}).Then(func(i int) (string, int) {
		return "large", i
	}).Else(func(i int) (string, int) {
		return "0", 0
	}).(func(int) (string, int))

	s, i := fn(1)
	suite.Equal("1", s)
	suite.Equal(1, i)

	s, i = fn(2)
	suite.Equal("0", s)
	suite.Equal(0, i)

	s, i = fn(100)
	suite.Equal("large", s)
	suite.Equal(100, i)
}

func (suite *FPIfTestSuite) TestWithAndNot() {
	var fn func(i int) bool
	And(func(i int) bool {
		return i > 0
	}, func(i int) bool {
		return i < 100
	}).And(Not(func(i int) bool {
		return i == 70
	})).To(&fn)

	suite.True(fn(50))
	suite.False(fn(0))
	suite.False(fn(70))
}

func (suite *FPIfTestSuite) TestWithOrNot() {
	var fn func(i int) bool
	And(func(i int) bool {
		return i > 0
	}, func(i int) bool {
		return i < 100
	}).Or(Not(func(i int) bool {
		return i == 200
	})).To(&fn)

	suite.True(fn(50))
	suite.False(fn(200))
	suite.True(fn(500))
}
func (suite *FPIfTestSuite) TestComplex() {
	between1And10 := And(func(i int) bool {
		return 1 <= i
	}, func(i int) bool {
		return i <= 10
	})
	between11And20ButNot15 := And(func(i int) bool {
		return 11 <= i
	}, func(i int) bool {
		return i <= 20
	}).And(Not(func(i int) bool {
		return i == 15
	}))
	conditionMap := When(between1And10).Then(func(i int) string { return "between 1 and 10" }).
		When(between11And20ButNot15).Then(func(i int) string { return "between 11 and 20 but not equal 15" }).
		Else(func(i int) string { return "other" }).(func(int) string)

	suite.Equal("between 1 and 10", conditionMap(5))
	suite.Equal("between 11 and 20 but not equal 15", conditionMap(11))
	suite.Equal("other", conditionMap(15))
	suite.Equal("other", conditionMap(100))
}

func (suite *FPIfTestSuite) TestConditionWithDefault() {
	between1And10 := And(func(i int) bool {
		return 1 <= i
	}, func(i int) bool {
		return i <= 10
	})
	conditionMap := When(between1And10).Then(func(i int) int { return i * 2 }).Else(nil).(func(int) int)

	suite.Equal(10, conditionMap(5))
	suite.Equal(11, conditionMap(11))
}

func (suite *FPIfTestSuite) TestDefaultCondition() {
	between1And10 := And(func(i int) bool {
		return 1 <= i
	}, func(i int) bool {
		return i <= 10
	})

	suite.Panics(func() {
		When(between1And10).Then(func(i int) string { return "between 1 and 10" }).Else(nil)
	})

	suite.Panics(func() {
		When(between1And10).Then(func(i int) (int, string) { return i, "between 1 and 10" }).Else(nil)
	})
}
