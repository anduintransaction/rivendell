package project

import "github.com/anduintransaction/rivendell/kubernetes"

func Status(namespace, context, kubeConfig, kind, name string) (kubernetes.RsStatus, error) {
	kubeContext, err := kubernetes.NewContext(namespace, context, kubeConfig)
	if err != nil {
		return kubernetes.RsStatusUnknown, err
	}
	return kubeContext.Resource().GetStatus(name, kind)
}
