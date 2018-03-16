package kubernetes

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ResourceTestSuite struct {
	suite.Suite
	resourceRoot string
	suffix       string
	kubeContext  *Context
}

func (s *ResourceTestSuite) SetupSuite() {
	s.resourceRoot = "../test-resources"
	s.suffix = fmt.Sprintf("%d", time.Now().Unix())
	namespace := fmt.Sprintf("test-ns-%s", s.suffix)
	kubeContext, err := NewContext(namespace, "", "")
	require.Nil(s.T(), err)
	s.kubeContext = kubeContext
	exists, err := kubeContext.Namespace().Create()
	require.Nil(s.T(), err)
	require.False(s.T(), exists)
}

func (s *ResourceTestSuite) TearDownSuite() {
	exists, err := s.kubeContext.Namespace().Delete()
	require.Nil(s.T(), err)
	require.True(s.T(), exists)
}

func (s *ResourceTestSuite) TestStaticResource() {
	if !testEnable() {
		fmt.Println("Skipping static resource test")
		return
	}
	s.testClusterResource(
		"cluster-role-"+s.suffix,
		"clusterrole",
		fmt.Sprintf(`kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cluster-role-%s
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "watch", "list"]
`, s.suffix),
	)

	s.testClusterResource(
		"cluster-role-binding-"+s.suffix,
		"clusterrolebinding",
		fmt.Sprintf(`kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cluster-role-binding-%s
subjects:
- kind: Group
  name: manager
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-role
  apiGroup: rbac.authorization.k8s.io
`, s.suffix),
	)

	s.testStaticResource("config-map", "configmap", "config-map.yml")
	s.testStaticResource("endpoints", "endpoints", "endpoints.yml")
	s.testStaticResource("ingress", "ingress", "ingress.yml")
	s.testStaticResource("role", "role", "role.yml")
	s.testStaticResource("role-binding", "rolebinding", "role-binding.yml")
	s.testStaticResource("secret", "secret", "secret.yml")
	s.testStaticResource("service-account", "serviceaccount", "service-account.yml")
	s.testStaticResource("service", "service", "service.yml")
}

func (s *ResourceTestSuite) testClusterResource(name, kind, content string) {
	s.createClusterResource(name, kind, content)
	s.verifyExists(name, kind)
	s.recreateClusterResource(name, kind, content)
	s.deleteResource(name, kind)
	s.verifyNotExists(name, kind)
	s.redeleteResource(name, kind)
}

func (s *ResourceTestSuite) testStaticResource(name, kind, filename string) {
	s.createStaticResource(name, kind, filename)
	s.verifyExists(name, kind)
	s.recreateStaticResource(name, kind, filename)
	s.deleteResource(name, kind)
	s.verifyNotExists(name, kind)
	s.redeleteResource(name, kind)
}

func (s *ResourceTestSuite) createClusterResource(name, kind, content string) {
	exists, err := s.kubeContext.Resource().Create(name, kind, string(content))
	require.Nil(s.T(), err, "should create %s %q", kind, name)
	require.False(s.T(), exists, "should create %s %q", kind, name)
}

func (s *ResourceTestSuite) recreateClusterResource(name, kind, content string) {
	exists, err := s.kubeContext.Resource().Create(name, kind, string(content))
	require.Nil(s.T(), err, "should recreate %s %q", kind, name)
	require.True(s.T(), exists, "should recreate %s %q", kind, name)
}

func (s *ResourceTestSuite) createStaticResource(name, kind, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "static", filename)
	content, err := ioutil.ReadFile(path)
	require.Nil(s.T(), err, "should read %q", path)
	exists, err := s.kubeContext.Resource().Create(name, kind, string(content))
	require.Nil(s.T(), err, "should create %s %q", kind, name)
	require.False(s.T(), exists, "should create %s %q", kind, name)
}

func (s *ResourceTestSuite) recreateStaticResource(name, kind, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "static", filename)
	content, err := ioutil.ReadFile(path)
	require.Nil(s.T(), err, "should read %q", path)
	exists, err := s.kubeContext.Resource().Create(name, kind, string(content))
	require.Nil(s.T(), err, "should recreate %s %q", kind, name)
	require.True(s.T(), exists, "should recreate %s %q", kind, name)
}

func (s *ResourceTestSuite) verifyExists(name, kind string) {
	exists, err := s.kubeContext.Resource().Exists(name, kind)
	require.Nil(s.T(), err, "should verify exists %s %q", kind, name)
	require.True(s.T(), exists, "%s %q should exists", kind, name)
}

func (s *ResourceTestSuite) verifyNotExists(name, kind string) {
	exists, err := s.kubeContext.Resource().Exists(name, kind)
	require.Nil(s.T(), err, "should not exists %s %q", kind, name)
	require.False(s.T(), exists, "%s %q should not exists", kind, name)
}

func (s *ResourceTestSuite) deleteResource(name, kind string) {
	exists, err := s.kubeContext.Resource().Delete(name, kind)
	require.Nil(s.T(), err, "should delete %s %q", kind, name)
	require.True(s.T(), exists, "should delete %s %q", kind, name)
}

func (s *ResourceTestSuite) redeleteResource(name, kind string) {
	exists, err := s.kubeContext.Resource().Delete(name, kind)
	require.Nil(s.T(), err, "should redelete %s %q", kind, name)
	require.False(s.T(), exists, "should redelete %s %q", kind, name)
}

func TestResource(t *testing.T) {
	suite.Run(t, new(ResourceTestSuite))
}
