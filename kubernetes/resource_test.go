package kubernetes

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
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
	if utils.TestEnable() {
		exists, err := kubeContext.Namespace().Create()
		require.Nil(s.T(), err)
		require.False(s.T(), exists)
	}
}

func (s *ResourceTestSuite) TearDownSuite() {
	if utils.TestEnable() {
		exists, err := s.kubeContext.Namespace().Delete()
		require.Nil(s.T(), err)
		require.True(s.T(), exists)
	}
}

func (s *ResourceTestSuite) TestStaticResource() {
	if !utils.TestEnable() {
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

func (s *ResourceTestSuite) TestPodBasedResource() {
	if !utils.TestEnable() {
		fmt.Println("Skipping pod-based resource test")
		return
	}
	s.testPodBasedResource("cronjob", "cronjob", "cronjob.yml")
	s.testPodBasedResource("daemonset", "daemonset", "daemonset.yml")
	s.testPodBasedResource("deployment", "deployment", "deployment.yml")
	s.testPodBasedResource("statefulset", "statefulset", "statefulset.yml")
}

func (s *ResourceTestSuite) TestPodResource() {
	if !utils.TestEnable() {
		fmt.Println("Skipping pod-based resource test")
		return
	}
	s.testPodResource("happy", "happy.yml", 0)
	s.testPodResource("start-slow", "start-slow.yml", 0)
	s.testPodResource("stop-slow", "stop-slow.yml", 0)
	s.testPodResource("stop-slow", "stop-slow.yml", 0)
	s.testPodResource("completed", "completed.yml", 15)
	s.testPodResource("error", "error.yml", 15)
	s.testPodResource("timeout", "timeout.yml", 20)
}

func (s *ResourceTestSuite) TestPodWait() {
	if !utils.TestEnable() {
		fmt.Println("Skipping pod-based resource test")
		return
	}
	s.testPodWait("success", "success.yml", true)
	s.testPodWait("error", "error.yml", false)
	s.testPodWait("timeout", "timeout.yml", false)
	_, err := s.kubeContext.Resource().Wait("not-exists", "pod")
	require.NotNil(s.T(), err)
	_, ok := stacktrace.RootCause(err).(ErrNotExist)
	require.True(s.T(), ok)
	s.createPodResource("success", "success.yml")
	s.deleteResource("success", "pod")
	_, err = s.kubeContext.Resource().Wait("success", "pod")
	require.NotNil(s.T(), err)
	_, ok = stacktrace.RootCause(err).(ErrNotExist)
	require.True(s.T(), ok)
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

func (s *ResourceTestSuite) testPodBasedResource(name, kind, filename string) {
	s.createPodBasedResource(name, kind, filename)
	s.verifyExists(name, kind)
	s.recreatePodBasedResource(name, kind, filename)
	s.deleteResource(name, kind)
	s.verifyNotExists(name, kind)
	s.redeleteResource(name, kind)
}

func (s *ResourceTestSuite) testPodResource(name, filename string, wait int) {
	s.createPodResource(name, filename)
	s.verifyExists(name, "pod")
	if wait > 0 {
		time.Sleep(time.Duration(wait) * time.Second)
	}
	s.recreatePodResource(name, filename)
	s.deleteResource(name, "pod")
	s.verifyNotExists(name, "pod")
	s.redeleteResource(name, "pod")
}

func (s *ResourceTestSuite) testPodWait(name, filename string, expectedSuccess bool) {
	s.createPodResource(name, filename)
	s.verifyExists(name, "pod")
	success, err := s.kubeContext.Resource().Wait(name, "pod")
	require.Nil(s.T(), err)
	require.Equal(s.T(), expectedSuccess, success)
	s.deleteResource(name, "pod")
}

func (s *ResourceTestSuite) createClusterResource(name, kind, content string) {
	s.createResourceFromContent(name, kind, content, false)
}

func (s *ResourceTestSuite) recreateClusterResource(name, kind, content string) {
	s.createResourceFromContent(name, kind, content, true)
}

func (s *ResourceTestSuite) createResourceFromContent(name, kind, content string, expectedExists bool) {
	exists, err := s.kubeContext.Resource().Create(name, kind, string(content))
	create := "create"
	if expectedExists {
		create = "recreate"
	}
	require.Nil(s.T(), err, "should %s %s %q", create, kind, name)
	require.Equal(s.T(), expectedExists, exists, "should %s %s %q", create, kind, name)
}

func (s *ResourceTestSuite) createStaticResource(name, kind, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "static", filename)
	s.createResourceFromFile(name, kind, path, false)
}

func (s *ResourceTestSuite) recreateStaticResource(name, kind, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "static", filename)
	s.createResourceFromFile(name, kind, path, true)
}

func (s *ResourceTestSuite) createPodBasedResource(name, kind, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "pod-based", filename)
	s.createResourceFromFile(name, kind, path, false)
}

func (s *ResourceTestSuite) recreatePodBasedResource(name, kind, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "pod-based", filename)
	s.createResourceFromFile(name, kind, path, true)
}

func (s *ResourceTestSuite) createPodResource(name, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "pod", filename)
	s.createResourceFromFile(name, "pod", path, false)
}

func (s *ResourceTestSuite) recreatePodResource(name, filename string) {
	path := filepath.Join(s.resourceRoot, "resource-test", "pod", filename)
	s.createResourceFromFile(name, "pod", path, true)
}

func (s *ResourceTestSuite) createResourceFromFile(name, kind, path string, expectedExists bool) {
	content, err := ioutil.ReadFile(path)
	require.Nil(s.T(), err, "should read %q", path)
	exists, err := s.kubeContext.Resource().Create(name, kind, string(content))
	create := "create"
	if expectedExists {
		create = "recreate"
	}
	require.Nil(s.T(), err, "should %s %s %q", create, kind, name)
	require.Equal(s.T(), expectedExists, exists, "should %s %s %q", create, kind, name)
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