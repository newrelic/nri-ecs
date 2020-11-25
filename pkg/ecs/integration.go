package ecs

import (
	"os"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

const (
	integrationName    = "com.newrelic.ecs"
	integrationVersion = "1.2.0"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	DebugMode bool `default:"false" help:"Enable ECS Agent Metadata debug mode."`
	Fargate   bool `default:"false" help:"If running on fargate"`
}

var (
	Args argumentList
)

func NewIntegration(args *argumentList) (*integration.Integration, error) {
	i, err := integration.New(integrationName, integrationVersion, integration.Args(args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	return i, nil
}
