# Taskmaster

A package that will allow the sync of Kubernetes CronJobs as tasks without having to deploy every variant of your parameterised task.

# Example
For example. A CronJob process which is intended to process the contents of an S3 Bucket, which can only be provided with 1 bucket as an argument, possibly for memory consumption reasons. When another bucket requires this process, Taskmaster can be used to sync another CronJob task using other means to enable or disable. This allows you to set up using parameters held in your configuration system.

This process could be set up as a CronJob to periodically sync your CronJob tasks.

    taskName := os.Getenv("TASK_NAME")
	env := os.Getenv("ENV")
	awsRegion := os.Getenv("AWS_REGION")
	imageVersion := os.Getenv("IMAGE_VERSION")

	fmt.Printf("Debug:  Syncing Task %s in Environment %s\n", taskName, env)
	fmt.Printf("Debug:  AWS Region Set to %s\n", awsRegion)
	fmt.Printf("Debug:  Syncing Version %s\n", imageVersion)

	// Create AWS Parameter Store Client
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(awsRegion)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		panic(err.Error())
	}
	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(awsRegion))

	// Create Kubernetes Client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	k8sClient, err := NewKubernetesClient(k8sConfig)
	if err != nil {
		panic(err.Error())
	}

	// Create Taskmaster Client 
	// with Verbose logging enabled
	taskmaster := NewTaskmaster(
		&Options{
			Debug: true,
		},
		k8sClient,
	)

	// Create CronJobs definitions and Sync
	taskKey := fmt.Sprintf("/%s/%s", env, taskName)

	namespace := getParamStoreValue(ssmsvc, fmt.Sprintf("%s/%s", taskKey, "namespace"))
	image := getParamStoreValue(ssmsvc, fmt.Sprintf("%s/%s", taskKey, "image"))
	buckets := getParamStoreValue(ssmsvc, fmt.Sprintf("%s/%s", taskKey, "buckets"))

	cjs := []CronJob{}

	for _, b := range strings.Split(strings.Trim(*buckets, " "), ",") {

		schedule := getParamStoreValue(ssmsvc, fmt.Sprintf("%s/%s/%s", taskKey, b, "schedule"))
		enabled := getParamStoreValue(ssmsvc, fmt.Sprintf("%s/%s/%s", taskKey, b, "enabled"))

		if enabled == "true" {
			cj := CronJob{
				Name:              fmt.Sprintf("%s-%s", taskName, b),
				Namespace:         *namespace,
				Taskname:          taskName,
				Image:             fmt.Sprintf("%s:%s", *image, imageVersion),
				Schedule:          *schedule,
				Args:              []string{"ls"},
				Env:               map[string]string{"ENV": env, "TESTVAR": "TESTVAL"},
				ConcurrencyPolicy: CronConcurrencyPolicyForbid,
				RestartPolicy:     CronRestartPolicyNever,
			}
		} else {
			fmt.Printf("Debug: Task %s for Bucket %s disabled.. skipping\n", taskName, b)
		}
		
		cjs = append(cjs, cj)
	}

	// Sync the CronJobs into Kubernetes
	err = taskmaster.Sync(cjs, taskName)
	if err != nil {
		panic(err.Error())
	}


# TODO
* Add ability to allow Taskmaster to manage the optimal sheduling for CronJobs
* Add ability to remove any prevously defined CronJobs