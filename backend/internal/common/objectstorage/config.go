package objectstorage

import (
	"fmt"
	"strings"
)

const DefaultRegion = "eu-central-1"

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

func (c Config) Validate() (bool, error) {
	var missing []string

	if strings.TrimSpace(c.Endpoint) == "" {
		missing = append(missing, "endpoint")
	}
	if strings.TrimSpace(c.AccessKeyID) == "" {
		missing = append(missing, "access key ID")
	}
	if strings.TrimSpace(c.SecretAccessKey) == "" {
		missing = append(missing, "secret access key")
	}
	if len(missing) > 0 {
		return false, fmt.Errorf("%w: missing %s", ErrInvalidConfig, strings.Join(missing, ", "))
	}

	return true, nil
}
