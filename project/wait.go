package project

import (
	"time"

	"github.com/anduintransaction/rivendell/kubernetes"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

// Wait for pod or job to complete
func Wait(namespace, context, kubeConfig, kind, name string, timeout int) error {
	utils.Info("Waiting for %s %q", kind, name)
	kubeContext, err := kubernetes.NewContext(namespace, context, kubeConfig)
	if err != nil {
		return err
	}
	waitChannel := make(chan error, 1)
	go func() {
		success, err := kubeContext.Resource().Wait(name, kind)
		if err != nil {
			waitChannel <- err
			return
		}
		if !success {
			waitChannel <- stacktrace.Propagate(ErrWaitFailed{name, kind}, "wait failed")
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
			err = stacktrace.Propagate(ErrWaitTimeout{name, kind}, "wait timeout")
		}
	}
	if err == nil {
		utils.Success("====> Done")
	}
	return err
}
