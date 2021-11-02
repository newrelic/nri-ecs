package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"

	v3 "github.com/newrelic/nri-ecs/internal/metadata/v3"
	"github.com/newrelic/nri-ecs/pkg/ecs"
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

	taskMetadataEnpoint, found := v3.TaskMetadataEndpoint()
	if !found {
		ecsIntegration.Logger().Errorf("unable to find task metadata endpoint")
		os.Exit(1)
	}

	body, err := v3.MetadataResponse(httpClient, taskMetadataEnpoint)
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

	var taskMetadata v3.TaskResponse
	if err = json.Unmarshal(body, &taskMetadata); err != nil {
		ecsIntegration.Logger().Errorf("unable to parse response body: %v", err)
		os.Exit(1)
	}

	awsRegion := v3.AWSRegionFromTask(taskMetadata.TaskARN)
	clusterName := clusterToClusterName(taskMetadata.Cluster)
	clusterARN := v3.ClusterARNFromTask(taskMetadata.TaskARN, clusterName)

	clusterEntity, err := ecs.NewClusterEntity(clusterARN, ecsIntegration)
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to create cluster entity: %v", err)
		os.Exit(1)
	}

	err = ecs.AddClusterInventory(clusterName, clusterARN, clusterEntity)
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to register cluster inventory: %v", err)
	}

	_, err = ecs.NewClusterHeartbeatMetricSet(clusterName, clusterARN, clusterEntity)
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to create metrics for cluster: %v", err)
	}

	launchType := ecs.NewLaunchType(args.Fargate)
	err = ecs.AddClusterInventoryToLocalEntity(clusterName, clusterARN, awsRegion, launchType, ecsIntegration)

	if err != nil {
		ecsIntegration.Logger().Errorf("unable to register cluster inventory to local entity: %v", err)
	}

	err = ecsIntegration.Publish()
	if err != nil {
		ecsIntegration.Logger().Errorf("unable to publish metrics for cluster: %v", err)
	}
}

// clusterToClusterName will convert the given cluster string returned by the V3 metadata endpoint to the cluster name.
// This is needed, because the Task v3 metadata endpoint returns different Cluster strings for Fargate and EC2:
// Fargate: Cluster is the ClusterARN
// EC2: Cluster is the ClusterName
func clusterToClusterName(cluster string) string {
	if !isECSARN(cluster) {
		return cluster
	}
	clusterName, _ := v3.ResourceNameAndARNBase(cluster)
	if clusterName == "" {
		return cluster
	}

	return clusterName
}

// isARN returns whether the given string is an ECS ARN by looking for
// whether the string starts with "arn:aws:ecs" and contains the correct number
// of sections delimited by colons(:).
// Copied from: https://github.com/aws/aws-sdk-go/blob/81abf80dec07700b14a91ece14b8eca6c5e6b4f8/aws/arn/arn.go#L81
func isECSARN(arn string) bool {
	const arnPrefix = "arn:aws:ecs"
	const arnSections = 6

	return strings.HasPrefix(arn, arnPrefix) && strings.Count(arn, ":") >= arnSections-1
}
