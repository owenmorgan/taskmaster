package taskmaster

import (
	"errors"
	"fmt"
)

// KubernetesMockClient -
type KubernetesMockClient struct {
	CronJobs map[string]CronJob
}

func (k8s KubernetesMockClient) cronJobKey(cj *CronJob) (key string) {
	key = fmt.Sprintf("%s_%s", cj.Name, cj.Namespace)
	return
}

func (k8s KubernetesMockClient) setCronJobs(cj *map[string]CronJob) {
	k8s.CronJobs = *cj
}

func (k8s KubernetesMockClient) getCronJobs() map[string]CronJob {
	return k8s.CronJobs
}

func (k8s KubernetesMockClient) flushCronJobs() {
	k8s.CronJobs = make(map[string]CronJob)
}

// NewKubernetesMockClient -
func NewKubernetesMockClient() (*KubernetesMockClient, error) {
	kmc := KubernetesMockClient{}
	kmc.CronJobs = make(map[string]CronJob)
	return &kmc, nil
}

// CreateCronJob -
func (k8s KubernetesMockClient) CreateCronJob(cj *CronJob) (err error) {
	key := k8s.cronJobKey(cj)
	if _, ok := k8s.CronJobs[key]; ok {
		err = errors.New("Cron Job already exists")
		return
	}
	k8s.CronJobs[key] = *cj
	return
}

// ListCronJobs -
func (k8s KubernetesMockClient) ListCronJobs(namespace string, lbls map[string]string) (cjs *[]CronJob, err error) {
	return
}

// GetCronJob -
func (k8s KubernetesMockClient) GetCronJob(namespace, name string) (cj *CronJob, err error) {
	key := fmt.Sprintf("%s_%s", name, namespace)
	cj = &CronJob{}
	if _, ok := k8s.CronJobs[key]; !ok {
		err = errors.New("Cron Job does not exist")
		return
	}
	ecj := k8s.CronJobs[key]
	return &ecj, nil
}

// UpdateCronJob -
func (k8s KubernetesMockClient) UpdateCronJob(cj *CronJob) (err error) {
	key := k8s.cronJobKey(cj)
	k8s.CronJobs[key] = *cj
	return
}

// CreateOrUpdateCronJob -
func (k8s KubernetesMockClient) CreateOrUpdateCronJob(cj *CronJob) (err error) {
	cjc, err := k8s.GetCronJob(cj.Namespace, cj.Name)
	exists := cjc.Name != "" && err == nil
	if exists {
		err = k8s.UpdateCronJob(cj)
	} else {
		err = k8s.CreateCronJob(cj)
	}
	return
}

// DeleteCronJob -
func (k8s KubernetesMockClient) DeleteCronJob(cj *CronJob) (err error) {
	return
}
