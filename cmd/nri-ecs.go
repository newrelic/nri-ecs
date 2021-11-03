package main

import (
	"encoding/json"
	"os"
	"time"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"

	"github.com/newrelic/nri-ecs/internal/ecs"
	"github.com/newrelic/nri-ecs/internal/ecs/metadata"
	"github.com/newrelic/nri-ecs/internal/infra"
)

var (
	integrationName    = "com.newrelic.ecs"
	integrationVersion = "1.3.1"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	DebugMode bool `default:"false" help:"Enable ECS Agent Metadata debug mode."`
	Fargate   bool `default:"false" help:"If running on fargate"`
}

func main() {
	args := &argumentList{}

	ecsIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(args))
	if err != nil {
		log.Error("failed to create integration: %w", err)
		os.Exit(1)
	}

	httpClient := ecs.ClientWithTimeout(5 * time.Second)

	taskMetadataEnpoint, found := metadata.TaskMetadataEndpoint()
	if !found {
		ecsIntegration.Logger().Errorf("unable to find task metadata endpoint")
		os.Exit(1)
	}

	body, err := metadata.MetadataResponse(httpClient, taskMetadataEnpoint)
	if err != nil {
		ecsIntegration.Logger().Errorf(
			"unable to get response from v3 task metadata endpoint (%s): %v",
			taskMetadataEnpoint,
			err,
		)
		os.Exit(1)
	}

	if args.DebugMode {
		log.Info("task metadata json response: %s", string(body))
		os.Exit(0)
	}

	var taskMetadata metadata.TaskResponse
	if err = json.Unmarshal(body, &taskMetadata); err != nil {
		ecsIntegration.Logger().Errorf("unable to parse response body: %v", err)
		os.Exit(1)
	}

	awsRegion := metadata.AWSRegionFromTask(taskMetadata.TaskARN)
	clusterName := metadata.ClusterToClusterName(taskMetadata.Cluster)
	clusterARN := metadata.ClusterARNFromTask(taskMetadata.TaskARN, clusterName)

	clusterEntity, err := infra.NewClusterEntity(clusterARN, ecsIntegration)
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to create cluster entity: %v", err)
		os.Exit(1)
	}

	err = infra.AddClusterInventory(clusterName, clusterARN, clusterEntity)
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to register cluster inventory: %v", err)
	}

	_, err = infra.NewClusterHeartbeatMetricSet(clusterName, clusterARN, clusterEntity)
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to create metrics for cluster: %v", err)
	}

	launchType := ecs.NewLaunchType(args.Fargate)
	err = infra.AddClusterInventoryToLocalEntity(clusterName, clusterARN, awsRegion, launchType, ecsIntegration)

	if err != nil {
		ecsIntegration.Logger().Errorf("unable to register cluster inventory to local entity: %v", err)
	}

	err = ecsIntegration.Publish()
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to publish metrics for cluster: %v", err)
	}
}
