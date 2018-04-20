package project

import (
	"os"
	"time"

	"github.com/anduintransaction/rivendell/kubernetes"
	"github.com/palantir/stacktrace"
)

// Logs .
func Logs(namespace, context, kubeConfig, name, containerName string, timeout int) error {
	kubeContext, err := kubernetes.NewContext(namespace, context, kubeConfig)
	if err != nil {
		return err
	}
	waitChannel := make(chan error, 1)
	go func() {
		err := kubeContext.Resource().Logs(name, containerName, os.Stdout, os.Stderr)
		if err != nil {
			waitChannel <- err
			return
		}
		waitChannel <- nil
	}()
	if timeout <= 0 {
		err = <-waitChannel
	} else {
		timer := time.NewTimer(time.Duration(timeout) * time.Second)
		select {
		case err = <-waitChannel:
		case <-timer.C:
			err = stacktrace.Propagate(ErrWaitTimeout{name, "pod"}, "wait timeout")
		}
	}
	return nil
}
