package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/v4/args"
	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	"github.com/newrelic/infra-integrations-sdk/v4/log"

	"github.com/newrelic/nri-ecs/internal/ecs"
	"github.com/newrelic/nri-ecs/internal/ecs/metadata"
	"github.com/newrelic/nri-ecs/internal/infra"
)

const (
	integrationName = "com.newrelic.ecs"
)

var (
	// Variables populated on compilation.
	integrationVersion = "0.0.0"
	gitCommit          = ""
)

type ArgumentList struct {
	sdkArgs.DefaultArgumentList
	Fargate     bool `default:"false" help:"If running on fargate"`
	ShowVersion bool `default:"false" help:"Print build information and exit"`
}

func main() {
	args := ArgumentList{}

	ecsIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create integration: %v", err))
	}

	if args.ShowVersion {
		fmt.Printf(
			"New Relic ECS integration Version: %s, Platform: %s, GoVersion: %s, GitCommit: %s\n",
			integrationVersion,
			fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), runtime.Version(), gitCommit)
		os.Exit(0)
	}

	if err := Run(ecsIntegration, args); err != nil {
		log.Fatal(fmt.Errorf("runing integration: %v", err))
	}
}

func Run(ecsIntegration *integration.Integration, args ArgumentList) error {
	httpClient := ecs.ClientWithTimeout(5 * time.Second)

	taskMetadataEnpoint, found := metadata.TaskMetadataEndpoint()
	if !found {
		return fmt.Errorf("unable to find task metadata endpoint")
	}

	body, err := metadata.MetadataResponse(httpClient, taskMetadataEnpoint)
	if err != nil {
		return fmt.Errorf("unable to get response from v3 task metadata endpoint (%s): %w", taskMetadataEnpoint, err)
	}

	log.Debug("task metadata json response: %s", string(body))

	taskMetadata := metadata.TaskResponse{}
	if err = json.Unmarshal(body, &taskMetadata); err != nil {
		return fmt.Errorf("unable to parse response body: %w", err)
	}

	if err = infra.PopulateIntegration(ecsIntegration, infra.NewClusterMetadata(taskMetadata, args.Fargate)); err != nil {
		return fmt.Errorf("populating integration metadata: %w", err)
	}

	if err = ecsIntegration.Publish(); err != nil {
		return fmt.Errorf("unable to publish metrics for cluster: %w", err)
	}

	return nil
}
