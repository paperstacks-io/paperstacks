package tests

import "os"

func IsIntegrationTest() bool {
	return os.Getenv("INTEGRATION") != ""
}
