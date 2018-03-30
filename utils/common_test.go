package utils

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CommonTestSuite struct {
	suite.Suite
	resourceRoot string
}

func (s *CommonTestSuite) SetupSuite() {
	s.resourceRoot = "../test-resources"
}

func (s *CommonTestSuite) TestGlob() {
	includePatterns := []string{
		filepath.Join(s.resourceRoot, "utils-test", "glob", "**", "file*"),
	}
	excludePatterns := []string{
		filepath.Join(s.resourceRoot, "utils-test", "glob", "**", "file3"),
	}
	files, err := GlobFiles(includePatterns, excludePatterns)
	require.Nil(s.T(), err)
	sort.Strings(files)
	expected := []string{
		filepath.Join(s.resourceRoot, "utils-test", "glob", "dir1", "file1"),
		filepath.Join(s.resourceRoot, "utils-test", "glob", "dir1", "file2"),
		filepath.Join(s.resourceRoot, "utils-test", "glob", "dir1", "subdir1", "file1"),
		filepath.Join(s.resourceRoot, "utils-test", "glob", "dir1", "subdir1", "file2"),
		filepath.Join(s.resourceRoot, "utils-test", "glob", "file1"),
		filepath.Join(s.resourceRoot, "utils-test", "glob", "file2"),
	}
	require.Equal(s.T(), expected, files)
}

func TestCommon(t *testing.T) {
	suite.Run(t, new(CommonTestSuite))
}
