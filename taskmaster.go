package taskmaster

import (
	"fmt"
)

// Taskmaster -
type Taskmaster struct {
	options    *Options
	kubernetes KubernetesClientInterface
}

// NewTaskmaster -
func NewTaskmaster(op *Options, k8s KubernetesClientInterface) *Taskmaster {
	return &Taskmaster{
		options:    op,
		kubernetes: k8s,
	}
}

// printDebug - Pring a debug message
func printDebug(message string, isDebug bool) {
	if isDebug {
		fmt.Printf("Debug:  %s\n", message)
	}
}

// Sync - Sync given CronJobs into Kubernetes
func (tm Taskmaster) Sync(cjs []CronJob, task string) (err error) {

	printDebug(fmt.Sprintf("Syncing %d CronJobs", len(cjs)), tm.options.Debug)

	for _, c := range cjs {

		name := c.Name
		namespace := c.Namespace

		printDebug(fmt.Sprintf("Syncing CronJob %s into Namespace %s", name, namespace), tm.options.Debug)

		err := tm.kubernetes.CreateOrUpdateCronJob(&c)
		if err != nil {
			printDebug(
				fmt.Sprintf("Error Syncing CronJob %s into Namespace %s, Error Message: %s", name, namespace, err.Error()),
				tm.options.Debug,
			)
			if !tm.options.IgnoreErrors {
				printDebug("Ending Sync..", tm.options.Debug)
				return err
			}
		}

		printDebug(fmt.Sprintf("Successfully Synced CronJob %s into Namespace %s", name, namespace), tm.options.Debug)

	}
	return err
}
