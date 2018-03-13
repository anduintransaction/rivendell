package kubernetes

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type NamespaceTestSuite struct {
	suite.Suite
}

func (s *NamespaceTestSuite) TestExists() {
	if !testEnable() {
		fmt.Println("Skipping namespace test exists")
		return
	}
	context, err := buildTestContext("default")
	require.Nil(s.T(), err)
	exists, err := context.Namespace().Exists()
	require.Nil(s.T(), err)
	require.True(s.T(), exists)
	context.namespace = "kube-"
	exists, err = context.Namespace().Exists()
	require.Nil(s.T(), err)
	require.False(s.T(), exists)
}

func (s *NamespaceTestSuite) TestCreateAndDelete() {
	if !testEnable() {
		fmt.Println("Skipping namespace test create and delete")
		return
	}
	namespace := fmt.Sprintf("test-ns-%d", time.Now().Unix())
	context, err := buildTestContext(namespace)
	require.Nil(s.T(), err)
	exists, err := context.Namespace().Create()
	require.Nil(s.T(), err)
	require.False(s.T(), exists)
	exists, err = context.Namespace().Create()
	require.Nil(s.T(), err)
	require.True(s.T(), exists)
	exists, err = context.Namespace().Delete()
	require.Nil(s.T(), err)
	require.True(s.T(), exists)
	exists, err = context.Namespace().Exists()
	require.Nil(s.T(), err)
	require.False(s.T(), exists)
}

func TestCommand(t *testing.T) {
	suite.Run(t, new(NamespaceTestSuite))
}
