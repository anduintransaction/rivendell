package project

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CommandTestSuite struct {
	suite.Suite
	resourceRoot string
}

func (s *CommandTestSuite) SetupSuite() {
	s.resourceRoot = "../test-resources"
}

func (s *CommandTestSuite) TestUpAndDown() {
	if !utils.TestEnable() {
		fmt.Println("Skipping up and down test")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "up-down", "project.yml")
	namespace := fmt.Sprintf("test-ns-%d", time.Now().Unix())
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	project, err := ReadProject(projectFile, namespace, context, kubeConfig, variables)
	require.Nil(s.T(), err)
	err = project.Up()
	require.Nil(s.T(), err)
	err = project.Down()
	require.Nil(s.T(), err)
}

func TestCommand(t *testing.T) {
	suite.Run(t, new(CommandTestSuite))
}
