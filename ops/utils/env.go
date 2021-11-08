package utils

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// GetEnvVars fetches environment variables for a given prefix
func GetEnvVars(ctx *pulumi.Context, p string) (out []string, err error) {
	var envVars []string
	if err = config.GetObject(ctx, p+"-ENV_VARS", &envVars); err != nil {
		return
	}
	for _, env := range envVars {
		out = append(out, fmt.Sprintf("%s=%s", env, config.Get(ctx, p+"-"+env)))
	}
	return
}
