package fp

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SourceTestSuite struct {
	suite.Suite
}

func (suite *SourceTestSuite) SetupTest() {
}

func (suite *SourceTestSuite) TearDownTest() {
}

func TestSourceTestSuite(t *testing.T) {
	suite.Run(t, new(SourceTestSuite))
}

func (suite *SourceTestSuite) TestLineSource() {
	r := bytes.NewBuffer(nil)
	r.WriteString("first\n")
	r.WriteString("second\n")

	s := NewLineSource(r)
	out := StreamOfSource(s).Map(strings.ToUpper).Strings()
	suite.Equal([]string{"FIRST", "SECOND"}, out)

	r = bytes.NewBuffer(nil)
	r.WriteString("first\n")
	r.WriteString("second\n")
	s = NewLineSource(r)
	q := StreamOfSource(s)
	suite.Equal("first", q.First().String())
	suite.Equal("first", q.First().String())
}

func (suite *SourceTestSuite) TestFileSource() {
	file, err := ioutil.TempFile("/tmp", "s")
	suite.NoError(err)
	defer os.RemoveAll(file.Name())

	content := "first\nsecond"
	ioutil.WriteFile(file.Name(), []byte(content), 0644)

	f, err := os.Open(file.Name())
	suite.NoError(err)
	defer f.Close()

	s := NewLineSource(f)
	out := StreamOfSource(s).Map(strings.ToUpper).Strings()
	suite.Equal([]string{"FIRST", "SECOND"}, out)
}
