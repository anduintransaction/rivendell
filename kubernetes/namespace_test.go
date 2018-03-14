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
	require.False(s.T(), exists, "first create should be successful")
	exists, err = context.Namespace().Create()
	require.Nil(s.T(), err)
	require.True(s.T(), exists, "should not create existed namespace")
	exists, err = context.Namespace().Delete()
	require.Nil(s.T(), err)
	require.True(s.T(), exists, "should delete successfully")
	exists, err = context.Namespace().Exists()
	require.Nil(s.T(), err)
	require.False(s.T(), exists, "should be removed")
	exists, err = context.Namespace().Create()
	require.Nil(s.T(), err)
	require.False(s.T(), exists, "can be recreated again")
	exists, err = context.Namespace().Delete()
	require.Nil(s.T(), err)
	require.True(s.T(), exists, "can be deleted again")
	exists, err = context.Namespace().Delete()
	require.Nil(s.T(), err)
	require.False(s.T(), exists, "delete again should be fine")
}

func TestCommand(t *testing.T) {
	suite.Run(t, new(NamespaceTestSuite))
}
