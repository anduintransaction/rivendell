package project

import (
	"testing"
	"time"

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
		LeafNodes: []string{"f", "g"},
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

/*
     _  a           b
    /  /   \      /    \
   |  c        f         e
   \  /       / \       /
    d  ------    ---- g
     \               /
      ------ h -----
*/
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
	trailForward := ""
	rg.WalkForward(func(g *ResourceGroup) error {
		trailForward += g.Name
		return nil
	})
	require.Equal(s.T(), "abcfedgh", trailForward)
	trailBackward := ""
	rg.WalkBackward(func(g *ResourceGroup) error {
		trailBackward += g.Name
		return nil
	})
	require.Equal(s.T(), "hdgcfeab", trailBackward)
}

func (s *ResourceTestSuite) TestWalkWithWait() {
	rg := &ResourceGraph{
		RootNodes: []string{"a"},
		ResourceGroups: map[string]*ResourceGroup{
			"a": &ResourceGroup{
				Name: "a",
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						Resources: []*Resource{
							&Resource{
								Name: "Test",
							},
						},
					},
				},
				Wait: []*WaitConfig{
					&WaitConfig{
						Name: "test",
						Kind: "test",
					},
				},
			},
			"b": &ResourceGroup{
				Name:   "b",
				Depend: []string{"a"},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						Resources: []*Resource{
							&Resource{
								Name: "Test",
							},
						},
					},
				},
			},
		},
	}
	err := rg.resolveChildren()
	require.Nil(s.T(), err)
	work := &struct {
		value int
		done  bool
	}{
		value: 0,
		done:  false,
	}
	err = rg.WalkForwardWithWait(func(g *ResourceGroup) error {
		if g.Name == "a" {
			go func() {
				require.Equal(s.T(), 21, work.value)
				time.Sleep(2 * time.Second)
				work.value = 42
				work.done = true
			}()
		} else {
			require.Equal(s.T(), 42, work.value)
		}
		return nil
	}, func(r *Resource, g *ResourceGroup) error {
		if g.Name == "a" {
			for !work.done {
				time.Sleep(time.Second)
			}
		}
		return nil
	}, func(name, kind string) error {
		work.value = 21
		time.Sleep(2 * time.Second)
		return nil
	})
	require.Nil(s.T(), err)
	work.value = 0
	work.done = false
	err = rg.WalkBackwardWithWait(func(g *ResourceGroup) error {
		if g.Name == "b" {
			go func() {
				time.Sleep(2 * time.Second)
				work.value = 42
				work.done = true
			}()
		} else {
			require.Equal(s.T(), 42, work.value)
		}
		return nil
	}, func(r *Resource, g *ResourceGroup) error {
		if g.Name == "b" {
			for !work.done {
				time.Sleep(time.Second)
			}
		}
		return nil
	})
	require.Nil(s.T(), err)
}

func TestResource(t *testing.T) {
	suite.Run(t, new(ResourceTestSuite))
}
