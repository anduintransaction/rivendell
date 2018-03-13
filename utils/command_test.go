package utils

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CommandTestSuite struct {
	suite.Suite
}

func (s *CommandTestSuite) TestCommandSuccess() {
	cmd := NewCommand("echo", "42")
	status, err := cmd.Run()
	require.Nil(s.T(), err, "command should run successfully")
	require.Equal(s.T(), 0, status.ExitCode)
	output, _ := ioutil.ReadAll(status.Stdout)
	require.Equal(s.T(), "42\n", string(output))
}

func (s *CommandTestSuite) TestCommandExitCode() {
	cmd := NewCommand("ls", "should-not-exist.txt")
	status, err := cmd.Run()
	require.Nil(s.T(), err, "command should run successfully")
	require.Equal(s.T(), 1, status.ExitCode)
	output, _ := ioutil.ReadAll(status.Stderr)
	require.True(s.T(), strings.Contains(string(output), "ls"), "stderr should contain ls")
	require.True(s.T(), strings.Contains(string(output), "should-not-exist.txt"), "stderr should contain file name")
}

func (s *CommandTestSuite) TestCommandFailure() {
	cmd := NewCommand("command-not-exist")
	_, err := cmd.Run()
	require.NotNil(s.T(), err, "command should not run successfully")
}

func TestCommand(t *testing.T) {
	suite.Run(t, new(CommandTestSuite))
}
