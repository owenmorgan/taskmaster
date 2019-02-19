package taskmaster

import (
	"github.com/AlekSi/pointer"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Restart Policies - https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy

const (
	// CronRestartPolicyNever -
	CronRestartPolicyNever = "Never"
	// CronRestartPolicyOnFailure -
	CronRestartPolicyOnFailure = "OnFailure"
	// CronRestartPolicyAlways -
	CronRestartPolicyAlways = "Always"
	// CronConcurrencyPolicyAllow -
	CronConcurrencyPolicyAllow = "Allow"
	// CronConcurrencyPolicyForbid -
	CronConcurrencyPolicyForbid = "Forbid"
	// CronConcurrencyPolicyReplce -
	CronConcurrencyPolicyReplce = "Replace"
)

// Interface -
type Interface interface {
	injectLabels(cj v1beta1.CronJob, taskname string) v1beta1.CronJob
	getStandardLabels(taskname string) map[string]string
	Sync() error
}

// KubernetesClientInterface -
type KubernetesClientInterface interface {
	CreateCronJob(cj *CronJob) (err error)
	ListCronJobs(namespace string, lbls map[string]string) (cjs *[]CronJob, err error)
	GetCronJob(namespace, name string) (cj *CronJob, err error)
	UpdateCronJob(cj *CronJob) (err error)
	CreateOrUpdateCronJob(cj *CronJob) (err error)
	DeleteCronJob(cj *CronJob) (err error)
}

// Options -
type Options struct {
	Debug        bool
	IgnoreErrors bool
}

// CronJob -
type CronJob struct {
	Name              string
	Namespace         string
	Taskname          string
	Image             string
	Schedule          string
	Args              []string
	Env               map[string]string
	RestartPolicy     string
	ConcurrencyPolicy string
	Suspend           *bool
}

func k8sCronJobtoCronJob(k8sCronJob v1beta1.CronJob) (cj *CronJob) {
	cj = &CronJob{}
	cj.Name = k8sCronJob.ObjectMeta.Name
	cj.Suspend = k8sCronJob.Spec.Suspend
	cj.Namespace = k8sCronJob.ObjectMeta.Namespace
	cj.Args = k8sCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args
	cj.Image = k8sCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
	cj.ConcurrencyPolicy = string(k8sCronJob.Spec.ConcurrencyPolicy)
	cj.RestartPolicy = string(k8sCronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy)
	cj.Schedule = k8sCronJob.Spec.Schedule
	cj.Env = map[string]string{}
	for _, e := range k8sCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env {
		cj.Env[e.Name] = e.Value
	}
	return
}

// cronJobToK8sCronJob - This is to convert a taskmaster.CronJob to a Kubernetes cronjob
func cronJobToK8sCronJob(cronJob *CronJob) (kcj *v1beta1.CronJob) {
	//job name cannot exceed (52 characters)

	kcj = &v1beta1.CronJob{}

	stdLabels := make(map[string]string)
	stdLabels["created-by"] = "taskmaster"
	stdLabels["task-name"] = cronJob.Taskname
	stdLabels["name"] = cronJob.Name
	stdLabels["namespace"] = cronJob.Namespace

	meta := metav1.ObjectMeta{
		Name:      cronJob.Name,
		Namespace: cronJob.Namespace,
		Labels:    stdLabels,
	}

	var envs []v1.EnvVar
	for n, v := range cronJob.Env {
		envs = append(envs, v1.EnvVar{Name: n, Value: v})
	}

	container := v1.Container{
		Name:  cronJob.Name,
		Image: cronJob.Image,
		Env:   envs,
		Args:  cronJob.Args,
	}

	var Conts []v1.Container
	Conts = append(Conts, container)
	PodSpec := v1.PodSpec{
		Containers:    Conts,
		RestartPolicy: v1.RestartPolicy(cronJob.RestartPolicy),
	}

	PodTemplateSpec := v1.PodTemplateSpec{Spec: PodSpec}
	Jspec := batchv1.JobSpec{Template: PodTemplateSpec}
	JobTemplateSpec := v1beta1.JobTemplateSpec{Spec: Jspec}

	spec := v1beta1.CronJobSpec{
		Schedule:                   cronJob.Schedule,
		Suspend:                    cronJob.Suspend,
		ConcurrencyPolicy:          v1beta1.ConcurrencyPolicy(cronJob.ConcurrencyPolicy),
		SuccessfulJobsHistoryLimit: pointer.ToInt32(3),
		FailedJobsHistoryLimit:     pointer.ToInt32(3),
		JobTemplate:                JobTemplateSpec,
	}

	kcj.ObjectMeta = meta
	kcj.Spec = spec

	return
}
