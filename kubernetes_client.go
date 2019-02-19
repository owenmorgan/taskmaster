package taskmaster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubernetesClient -
type KubernetesClient struct {
	kubernetes *kubernetes.Clientset
}

// NewKubernetesClient -
func NewKubernetesClient(config *rest.Config) (*KubernetesClient, error) {
	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesClient{kubernetes: k8s}, nil
}

// CreateCronJob -
func (k8s KubernetesClient) CreateCronJob(cj *CronJob) (err error) {
	kcj := cronJobToK8sCronJob(cj)
	_, err = k8s.kubernetes.BatchV1beta1().CronJobs(cj.Namespace).Create(kcj)

	return
}

// ListCronJobs -
func (k8s KubernetesClient) ListCronJobs(namespace string, lbls map[string]string) (cjs *[]CronJob, err error) {
	set := labels.Set(lbls)
	opts := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	cjl, err := k8s.kubernetes.BatchV1beta1().CronJobs(namespace).List(opts)
	if err != nil {
		return
	}
	for _, c := range cjl.Items {
		*cjs = append(*cjs, *k8sCronJobtoCronJob(c))
	}

	return
}

// GetCronJob -
func (k8s KubernetesClient) GetCronJob(namespace, name string) (cj *CronJob, err error) {
	kcj, err := k8s.kubernetes.BatchV1beta1().CronJobs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return
	}
	cj = k8sCronJobtoCronJob(*kcj)
	return cj, nil
}

// UpdateCronJob -
func (k8s KubernetesClient) UpdateCronJob(cj *CronJob) (err error) {
	kcj := cronJobToK8sCronJob(cj)
	_, err = k8s.kubernetes.BatchV1beta1().CronJobs(cj.Namespace).Update(kcj)

	return
}

// CreateOrUpdateCronJob -
func (k8s KubernetesClient) CreateOrUpdateCronJob(cj *CronJob) (err error) {
	cjc, err := k8s.GetCronJob(cj.Namespace, cj.Name)

	exists := cjc != nil && err == nil
	if exists {
		err = k8s.UpdateCronJob(cj)
	} else {
		err = k8s.CreateCronJob(cj)
	}
	return
}

// DeleteCronJob -
func (k8s KubernetesClient) DeleteCronJob(cj *CronJob) (err error) {
	opts := v1.DeleteOptions{}
	return k8s.kubernetes.BatchV1beta1().CronJobs(cj.Namespace).Delete(cj.Name, &opts)
}
