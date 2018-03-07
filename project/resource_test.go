package project

import (
	"testing"

	"github.com/palantir/stacktrace"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ResourceTestSuite struct {
	suite.Suite
	resourceRoot string
}

func (s *ResourceTestSuite) SetupTest() {
	s.resourceRoot = "../test-resources"
}

func (s *ResourceTestSuite) TestResolveChildren() {
	rg1 := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name:   "a",
				Depend: []string{},
			},
			"b": &ResourceGroup{
				Name:   "b",
				Depend: []string{},
			},
			"c": &ResourceGroup{
				Name:   "c",
				Depend: []string{"a"},
			},
			"d": &ResourceGroup{
				Name:   "d",
				Depend: []string{"a", "b"},
			},
			"e": &ResourceGroup{
				Name:   "e",
				Depend: []string{"b"},
			},
			"f": &ResourceGroup{
				Name:   "f",
				Depend: []string{"a", "c", "d"},
			},
			"g": &ResourceGroup{
				Name:   "g",
				Depend: []string{"d", "e"},
			},
		},
	}
	err := rg1.resolveChildren()
	require.Nil(s.T(), err, "children should be resolved successfully")
	expectedRg1 := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name:     "a",
				Depend:   []string{},
				Children: []string{"c", "d", "f"},
			},
			"b": &ResourceGroup{
				Name:     "b",
				Depend:   []string{},
				Children: []string{"d", "e"},
			},
			"c": &ResourceGroup{
				Name:     "c",
				Depend:   []string{"a"},
				Children: []string{"f"},
			},
			"d": &ResourceGroup{
				Name:     "d",
				Depend:   []string{"a", "b"},
				Children: []string{"f", "g"},
			},
			"e": &ResourceGroup{
				Name:     "e",
				Depend:   []string{"b"},
				Children: []string{"g"},
			},
			"f": &ResourceGroup{
				Name:   "f",
				Depend: []string{"a", "c", "d"},
			},
			"g": &ResourceGroup{
				Name:   "g",
				Depend: []string{"d", "e"},
			},
		},
	}
	require.Equal(s.T(), expectedRg1, rg1)

	rg2 := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name:   "a",
				Depend: []string{"b"},
			},
		},
	}
	err = rg2.resolveChildren()
	expectedErr := ErrMissingDependency{
		Child:  "a",
		Parent: "b",
	}
	require.Equal(s.T(), stacktrace.RootCause(err), expectedErr)
}

func (s *ResourceTestSuite) TestCyclic() {
	rg1 := &ResourceGraph{}
	err := rg1.cyclicCheck()
	require.Nil(s.T(), err)
	rg2 := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name:   "a",
				Depend: []string{"a"},
			},
		},
	}
	rg2.resolveChildren()
	err = rg2.cyclicCheck()
	require.Equal(s.T(), ErrCyclicDependency{"a"}, stacktrace.RootCause(err))
	rg3 := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name:     "a",
				Depend:   []string{},
				Children: []string{"c", "d", "f"},
			},
			"b": &ResourceGroup{
				Name:     "b",
				Depend:   []string{},
				Children: []string{"d", "e"},
			},
			"c": &ResourceGroup{
				Name:     "c",
				Depend:   []string{"a"},
				Children: []string{"f"},
			},
			"d": &ResourceGroup{
				Name:     "d",
				Depend:   []string{"a", "b"},
				Children: []string{"f"},
			},
			"e": &ResourceGroup{
				Name:     "e",
				Depend:   []string{"b"},
				Children: []string{"g"},
			},
			"f": &ResourceGroup{
				Name:     "f",
				Depend:   []string{"a", "c", "d"},
				Children: []string{"h"},
			},
			"g": &ResourceGroup{
				Name:     "g",
				Depend:   []string{"d", "e"},
				Children: []string{"h"},
			},
			"h": &ResourceGroup{
				Name:     "h",
				Depend:   []string{"f", "g"},
				Children: []string{"e"},
			},
		},
	}
	err = rg3.cyclicCheck()
	require.NotNil(s.T(), err)
	rg4 := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name:   "a",
				Depend: []string{},
			},
			"b": &ResourceGroup{
				Name:   "b",
				Depend: []string{"b"},
			},
			"c": &ResourceGroup{
				Name:   "c",
				Depend: []string{},
			},
			"d": &ResourceGroup{
				Name:   "d",
				Depend: []string{},
			},
			"e": &ResourceGroup{
				Name:   "e",
				Depend: []string{},
			},
		},
	}
	rg4.resolveChildren()
	err = rg4.cyclicCheck()
	require.Equal(s.T(), ErrCyclicDependency{"b"}, stacktrace.RootCause(err))
}

func (s *ResourceTestSuite) TestWalk() {
	rg := &ResourceGraph{
		RootNodes: []string{"a", "b"},
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name:   "a",
				Depend: []string{},
			},
			"b": &ResourceGroup{
				Name:   "b",
				Depend: []string{},
			},
			"c": &ResourceGroup{
				Name:   "c",
				Depend: []string{"a"},
			},
			"f": &ResourceGroup{
				Name:   "f",
				Depend: []string{"a", "b"},
			},
			"e": &ResourceGroup{
				Name:   "e",
				Depend: []string{"b"},
			},
			"d": &ResourceGroup{
				Name:   "d",
				Depend: []string{"c", "a", "f"},
			},
			"g": &ResourceGroup{
				Name:   "g",
				Depend: []string{"f", "e"},
			},
			"h": &ResourceGroup{
				Name:   "h",
				Depend: []string{"d", "g"},
			},
		},
	}
	rg.resolveChildren()
	trail := ""
	rg.Walk(func(g *ResourceGroup) error {
		trail += g.Name
		return nil
	})
	require.Equal(s.T(), "abcfedgh", trail)
}

func TestResource(t *testing.T) {
	suite.Run(t, new(ResourceTestSuite))
}
