package taskmaster

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesCronJobToCronJob(t *testing.T) {

	expectedCronJob := CronJob{
		Name:              "test_123",
		Namespace:         "testnamespace_1234",
		Taskname:          "testtaskname_4444",
		Image:             "testimage_1231",
		Schedule:          "5 4 * * *",
		Args:              []string{"testargs1", "testargs2"},
		Env:               map[string]string{"ENV": "UAT", "TESTVAR": "TESTVAL"},
		ConcurrencyPolicy: CronConcurrencyPolicyForbid,
		RestartPolicy:     CronRestartPolicyNever,
	}

	meta := metav1.ObjectMeta{
		Name:      expectedCronJob.Name,
		Namespace: expectedCronJob.Namespace,
	}

	var envs []v1.EnvVar
	for n, v := range expectedCronJob.Env {
		envs = append(envs, v1.EnvVar{Name: n, Value: v})
	}

	container := v1.Container{
		Name:  expectedCronJob.Name,
		Image: expectedCronJob.Image,
		Env:   envs,
		Args:  expectedCronJob.Args,
	}

	var Conts []v1.Container
	Conts = append(Conts, container)
	PodSpec := v1.PodSpec{
		Containers:    Conts,
		RestartPolicy: v1.RestartPolicy(expectedCronJob.RestartPolicy),
	}

	PodTemplateSpec := v1.PodTemplateSpec{Spec: PodSpec}
	Jspec := batchv1.JobSpec{Template: PodTemplateSpec}
	JobTemplateSpec := v1beta1.JobTemplateSpec{Spec: Jspec}

	spec := v1beta1.CronJobSpec{
		Schedule:                   expectedCronJob.Schedule,
		ConcurrencyPolicy:          v1beta1.ConcurrencyPolicy(expectedCronJob.ConcurrencyPolicy),
		SuccessfulJobsHistoryLimit: pointer.ToInt32(3),
		FailedJobsHistoryLimit:     pointer.ToInt32(3),
		JobTemplate:                JobTemplateSpec,
	}

	kcj := v1beta1.CronJob{
		ObjectMeta: meta,
		Spec:       spec,
	}

	testCronJob := k8sCronJobtoCronJob(kcj)

	assert.Equal(t, testCronJob.Name, kcj.ObjectMeta.Name)
	assert.Equal(t, testCronJob.Namespace, kcj.ObjectMeta.Namespace)
	assert.Equal(t, testCronJob.Schedule, kcj.Spec.Schedule)
	assert.Equal(t, testCronJob.ConcurrencyPolicy, string(kcj.Spec.ConcurrencyPolicy))
	assert.Equal(t, testCronJob.RestartPolicy, string(kcj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy))
	assert.Equal(t, testCronJob.Image, kcj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, testCronJob.Args, kcj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args)
	assert.Equal(t, len(testCronJob.Env), len(kcj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env))

	for _, v := range kcj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env {
		assert.Equal(t, testCronJob.Env[v.Name], expectedCronJob.Env[v.Name])
	}

	testK8sCronJob := cronJobToK8sCronJob(&expectedCronJob)

	assert.Equal(t, testCronJob.Name, testK8sCronJob.ObjectMeta.Name)
	assert.Equal(t, testCronJob.Namespace, testK8sCronJob.ObjectMeta.Namespace)
	assert.Equal(t, testCronJob.Schedule, testK8sCronJob.Spec.Schedule)
	assert.Equal(t, testCronJob.ConcurrencyPolicy, string(testK8sCronJob.Spec.ConcurrencyPolicy))
	assert.Equal(t, testCronJob.RestartPolicy, string(testK8sCronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy))
	assert.Equal(t, testCronJob.Image, testK8sCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, testCronJob.Args, testK8sCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args)
	assert.Equal(t, len(testCronJob.Env), len(testK8sCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env))

}

func TestSyncCronJob(t *testing.T) {

	k8sClient, err := NewKubernetesMockClient()
	k8sClient.flushCronJobs()
	assert.Nil(t, err)

	tm := NewTaskmaster(
		&Options{
			Debug: false,
		},
		k8sClient,
	)

	tmcj := CronJob{
		Name:              "testname",
		Namespace:         "testnamespace",
		Taskname:          "testtask",
		Image:             "busybox:latest",
		Schedule:          "*/1 * * * *",
		Args:              []string{"ls"},
		Env:               map[string]string{"ENV": "UAT", "TESTVAR": "TESTVAL"},
		ConcurrencyPolicy: CronConcurrencyPolicyForbid,
		RestartPolicy:     CronRestartPolicyNever,
	}

	tmcjs := []CronJob{tmcj}

	assert.Len(t, k8sClient.getCronJobs(), 0)

	err = tm.Sync(
		tmcjs,
		"testtask",
	)

	assert.Nil(t, err)

	cjs := k8sClient.getCronJobs()

	assert.Len(t, cjs, 1)
	assert.Equal(t, cjs[k8sClient.cronJobKey(&tmcj)].Schedule, "*/1 * * * *")

	tmcj = CronJob{
		Name:              "testname",
		Namespace:         "testnamespace",
		Taskname:          "testtask",
		Image:             "busybox:latest",
		Schedule:          "*/10 * * * *",
		Args:              []string{"ls"},
		Env:               map[string]string{"ENV": "UAT", "TESTVAR": "TESTVAL"},
		ConcurrencyPolicy: CronConcurrencyPolicyForbid,
		RestartPolicy:     CronRestartPolicyNever,
	}

	tmcjs = []CronJob{tmcj}

	assert.Len(t, k8sClient.getCronJobs(), 1)

	err = tm.Sync(
		tmcjs,
		"testtask",
	)

	assert.Nil(t, err)

	cjs = k8sClient.getCronJobs()

	assert.Len(t, k8sClient.getCronJobs(), 1)
	assert.Equal(t, cjs[k8sClient.cronJobKey(&tmcj)].Schedule, "*/10 * * * *")

}
