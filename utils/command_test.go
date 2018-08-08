package utils

import (
	"io/ioutil"
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
	cmd := NewCommand("bash", "-c", "exit 1")
	status, err := cmd.Run()
	require.Nil(s.T(), err, "command should run successfully")
	require.Equal(s.T(), 1, status.ExitCode)
}

func (s *CommandTestSuite) TestCommandFailure() {
	cmd := NewCommand("command-not-exist")
	_, err := cmd.Run()
	require.NotNil(s.T(), err, "command should not run successfully")
}

func TestCommand(t *testing.T) {
	suite.Run(t, new(CommandTestSuite))
}
