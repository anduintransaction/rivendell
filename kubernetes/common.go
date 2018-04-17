package kubernetes

import (
	"os"
	"time"
)

const (
	defaultTerminateInterval   = 3 * time.Second
	defaultTerminateCheckLimit = 40
	defaultPendingInterval     = 3 * time.Second
	defaultPendingCheckLimit   = 40
)

func buildTestContext(namespace string) (*Context, error) {
	kubeContext := os.Getenv("KUBERNETES_TEST_CONTEXT")
	if kubeContext == "" {
		kubeContext = "minikube"
	}
	kubeConfig := os.Getenv("KUBERNETES_TEST_KUBE_CONFIG")
	return NewContext(namespace, kubeContext, kubeConfig)
}
