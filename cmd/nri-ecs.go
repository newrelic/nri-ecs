package main

import (
	"encoding/json"
	"fmt"
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
	Fargate bool `default:"false" help:"If running on fargate"`
}

func main() {
	args := &argumentList{}

	ecsIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(args))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create integration: %v", err))
	}

	httpClient := ecs.ClientWithTimeout(5 * time.Second)

	taskMetadataEnpoint, found := metadata.TaskMetadataEndpoint()
	if !found {
		log.Fatal(fmt.Errorf("unable to find task metadata endpoint"))
	}

	body, err := metadata.MetadataResponse(httpClient, taskMetadataEnpoint)
	if err != nil {
		log.Fatal(
			fmt.Errorf("unable to get response from v3 task metadata endpoint (%s): %v", taskMetadataEnpoint, err),
		)
	}

	log.Debug("task metadata json response: %s", string(body))

	taskMetadata := metadata.TaskResponse{}
	if err = json.Unmarshal(body, &taskMetadata); err != nil {
		log.Fatal(fmt.Errorf("unable to parse response body: %v", err))
	}

	awsRegion := metadata.AWSRegionFromTask(taskMetadata.TaskARN)
	clusterName := metadata.ClusterToClusterName(taskMetadata.Cluster)
	clusterARN := metadata.ClusterARNFromTask(taskMetadata.TaskARN, clusterName)

	clusterEntity, err := infra.NewClusterEntity(clusterARN, ecsIntegration)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to create cluster entity: %v", err))
	}

	if err = infra.AddClusterInventory(clusterName, clusterARN, clusterEntity); err != nil {
		log.Error("unable to register cluster inventory: %v", err)
	}

	if _, err = infra.NewClusterHeartbeatMetricSet(clusterName, clusterARN, clusterEntity); err != nil {
		log.Error("unable to create metrics for cluster: %v", err)
	}

	launchType := ecs.NewLaunchType(args.Fargate)
	if err = infra.AddClusterInventoryToLocalEntity(clusterName, clusterARN, awsRegion, launchType, ecsIntegration); err != nil {
		log.Error("unable to register cluster inventory to local entity: %v", err)
	}

	if err = ecsIntegration.Publish(); err != nil {
		log.Error("unable to publish metrics for cluster: %v", err)
	}
}
