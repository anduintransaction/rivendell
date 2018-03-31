package utils

import "os"

// TestEnable .
func TestEnable() bool {
	kubernetesTestValue := os.Getenv("KUBERNETES_TEST_ENABLE")
	return kubernetesTestValue == "true" || kubernetesTestValue == "1"
}
