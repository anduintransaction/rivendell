package kubernetes

import "os"

func testEnable() bool {
	kubernetesTestValue := os.Getenv("KUBERNETES_TEST_ENABLE")
	return kubernetesTestValue == "true" || kubernetesTestValue == "1"
}

func buildTestContext(namespace string) (*Context, error) {
	kubeContext := os.Getenv("KUBERNETES_TEST_CONTEXT")
	if kubeContext == "" {
		kubeContext = "minikube"
	}
	kubeConfig := os.Getenv("KUBERNETES_TEST_KUBE_CONFIG")
	return NewContext(namespace, kubeContext, kubeConfig)
}
