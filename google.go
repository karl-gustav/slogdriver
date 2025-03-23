package slogdriver

import "os"

func OnGCP() bool {
	return os.Getenv("K_SERVICE") != "" || os.Getenv("GAE_SERVICE") != "" || os.Getenv("CLOUD_RUN_JOB") != ""
}
