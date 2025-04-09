package google

import "os"

func OnGCP() bool {
	return GetServiceName() != ""
}

func GetServiceName() string {
	for _, env := range []string{"K_SERVICE", "CLOUD_RUN_JOB", "GAE_SERVICE"} {
		if val := os.Getenv(env); val != "" {
			return val
		}
	}
	return ""
}
