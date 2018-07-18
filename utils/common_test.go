package utils

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	yaml "gopkg.in/yaml.v2"
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

func (s *CommonTestSuite) TestReplaceEnv() {
	str := "$(RIVENDELL_USER) $RIVENDELL_USER ${RIVENDELL_USER} $(EMPTY)"
	os.Setenv("RIVENDELL_USER", "rivendell")
	expected := "rivendell $RIVENDELL_USER ${RIVENDELL_USER} "
	require.Equal(s.T(), expected, ExpandEnv(str))
}

func (s *CommonTestSuite) TestStringStack() {
	stack := NewStringStack()
	_, err := stack.Pop()
	require.Equal(s.T(), ErrEmptyStack, err)
	_, err = stack.Head()
	require.Equal(s.T(), ErrEmptyStack, err)
	stack.Push("1")
	stack.Push("2")
	value, err := stack.Head()
	require.Equal(s.T(), "2", value)
	require.Nil(s.T(), err)
	value, err = stack.Pop()
	require.Equal(s.T(), "2", value)
	require.Nil(s.T(), err)
	value, err = stack.Pop()
	require.Equal(s.T(), "1", value)
	require.Nil(s.T(), err)
	_, err = stack.Pop()
	require.Equal(s.T(), ErrEmptyStack, err)
	_, err = stack.Head()
	require.Equal(s.T(), ErrEmptyStack, err)
}

func (s *CommonTestSuite) TestTemplater() {
	variables := map[string]string{
		"value1": "value1",
		"value2": "value2",
		"value3": "value3",
	}
	templateFile := filepath.Join(s.resourceRoot, "utils-test", "templater", "parent.yml")
	content, err := ExecuteTemplate(templateFile, variables)
	require.Nil(s.T(), err)
	parsedYAML := &struct {
		Key1 string `yaml:"key1"`
		Sub  *struct {
			Key2 string `yaml:"key2"`
			Sub  *struct {
				Key3 string `yaml:"key3"`
			} `yaml:"sub"`
		} `yaml:"sub"`
	}{}
	expectedYAML := &struct {
		Key1 string `yaml:"key1"`
		Sub  *struct {
			Key2 string `yaml:"key2"`
			Sub  *struct {
				Key3 string `yaml:"key3"`
			} `yaml:"sub"`
		} `yaml:"sub"`
	}{
		Key1: "value1",
		Sub: &struct {
			Key2 string `yaml:"key2"`
			Sub  *struct {
				Key3 string `yaml:"key3"`
			} `yaml:"sub"`
		}{
			Key2: "value2",
			Sub: &struct {
				Key3 string `yaml:"key3"`
			}{
				Key3: "value3",
			},
		},
	}

	err = yaml.Unmarshal(content, parsedYAML)
	require.Nil(s.T(), err)
	require.Equal(s.T(), expectedYAML, parsedYAML)
}

func TestCommon(t *testing.T) {
	suite.Run(t, new(CommonTestSuite))
}
