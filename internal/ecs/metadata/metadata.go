package metadata

// Copyright 2017-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	containerMetadataEnvVar = "ECS_CONTAINER_METADATA_URI"
	maxRetries              = 4
	durationBetweenRetries  = time.Second
)

// TaskResponse defines the schema for the task response JSON object
type TaskResponse struct {
	Cluster            string              `json:"Cluster"`
	TaskARN            string              `json:"TaskARN"`
	Family             string              `json:"Family"`
	Revision           string              `json:"Revision"`
	DesiredStatus      string              `json:"DesiredStatus,omitempty"`
	KnownStatus        string              `json:"KnownStatus"`
	AvailabilityZone   string              `json:"AvailabilityZone"`
	Containers         []ContainerResponse `json:"Containers,omitempty"`
	Limits             *LimitsResponse     `json:"Limits,omitempty"`
	PullStartedAt      *time.Time          `json:"PullStartedAt,omitempty"`
	PullStoppedAt      *time.Time          `json:"PullStoppedAt,omitempty"`
	ExecutionStoppedAt *time.Time          `json:"ExecutionStoppedAt,omitempty"`
}

// ContainerResponse defines the schema for the container response
// JSON object
type ContainerResponse struct {
	ID            string            `json:"DockerId"`
	Name          string            `json:"Name"`
	DockerName    string            `json:"DockerName"`
	Image         string            `json:"Image"`
	ImageID       string            `json:"ImageID"`
	Ports         []PortResponse    `json:"Ports,omitempty"`
	Labels        map[string]string `json:"Labels,omitempty"`
	DesiredStatus string            `json:"DesiredStatus"`
	KnownStatus   string            `json:"KnownStatus"`
	ExitCode      *int              `json:"ExitCode,omitempty"`
	Limits        LimitsResponse    `json:"Limits"`
	CreatedAt     *time.Time        `json:"CreatedAt,omitempty"`
	StartedAt     *time.Time        `json:"StartedAt,omitempty"`
	FinishedAt    *time.Time        `json:"FinishedAt,omitempty"`
	Type          string            `json:"Type"`
	Networks      []Network         `json:"Networks,omitempty"`
	Health        HealthStatus      `json:"Health,omitempty"`
}

// LimitsResponse defines the schema for task/cpu limits response
// JSON object
type LimitsResponse struct {
	CPU    *float64 `json:"CPU,omitempty"`
	Memory *int64   `json:"Memory,omitempty"`
}

type HealthStatus struct {
	Status   string     `json:"status,omitempty"`
	Since    *time.Time `json:"statusSince,omitempty"`
	ExitCode int        `json:"exitCode,omitempty"`
	Output   string     `json:"output,omitempty"`
}

// PortResponse defines the schema for portmapping response JSON
// object.
type PortResponse struct {
	ContainerPort uint16 `json:"ContainerPort,omitempty"`
	Protocol      string `json:"Protocol,omitempty"`
	HostPort      uint16 `json:"HostPort,omitempty"`
}

// Network is a struct that keeps track of metadata of a network interface
type Network struct {
	NetworkMode   string   `json:"NetworkMode,omitempty"`
	IPv4Addresses []string `json:"IPv4Addresses,omitempty"`
	IPv6Addresses []string `json:"IPv6Addresses,omitempty"`
}

// MetadataResponse gets the response from the given endpoint using the given HTTP client.
func MetadataResponse(client *http.Client, endpoint string) ([]byte, error) {
	var resp []byte
	var err error
	for i := 0; i < maxRetries; i++ {
		resp, err = metadataResponse(client, endpoint)
		if err == nil {
			return resp, nil
		}
		fmt.Fprintf(os.Stderr, "Attempt [%d/%d]: unable to get metadata response from '%s': %v",
			i, maxRetries, endpoint, err)
		time.Sleep(durationBetweenRetries)
	}

	return nil, err
}

func metadataResponse(client *http.Client, endpoint string) ([]byte, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %v", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("incorrect status code  %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	return body, nil
}

// TaskMetadataEndpoint returns the V3 endpoint to fetch task metadata.
func TaskMetadataEndpoint() (string, bool) {
	baseEndpoint, found := metadataV3Endpoint()
	if !found {
		return "", found
	}
	return baseEndpoint + "/task", found
}

// metadataV3Endpoint returns the v3 metadata endpoint configured via the ECS_CONTAINER_METADATA_URI environment
// variable.
func metadataV3Endpoint() (string, bool) {
	return os.LookupEnv(containerMetadataEnvVar)
}

// ClusterARNFromTask builds a cluster ARN from a task ARN and a given cluster name.
// Example of task ARN (EC2): arn:aws:ecs:us-west-2:xxxxxxxx:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152
// Example of task ARN (Fargate): arn:aws:ecs:us-west-2:xxxxxxxx:task/37e873f6-37b4-42a7-af47-eac7275c6152
// Example of a cluster ARN: arn:aws:ecs:us-west-2:xxxxxxxx:cluster/test
func ClusterARNFromTask(taskARN string, clusterName string) string {
	_, arnPrefix := ResourceNameAndARNBase(taskARN)
	return fmt.Sprintf("%s:cluster/%s", arnPrefix, clusterName)
}

// AWSRegionFromTask returns the aws region from the task ARN.
// Example of task ARN: arn:aws:ecs:us-west-2:xxxxxxxx:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152
func AWSRegionFromTask(taskARN string) string {
	_, arnPrefix := ResourceNameAndARNBase(taskARN)
	return strings.Split(arnPrefix, ":")[3]
}

// ResourceNameAndARNBase explodes a resource ARN and returns the resource's name and the base ARN prefix for the
// account and region of the original resource.
func ResourceNameAndARNBase(resourceARN string) (resourceName string, arnPrefix string) {
	arnPrefixAndType := resourceARN[:strings.Index(resourceARN, "/")-1]
	arnPrefix = arnPrefixAndType[:strings.LastIndex(arnPrefixAndType, ":")]
	resourceName = resourceARN[strings.Index(resourceARN, "/")+1:]
	return resourceName, arnPrefix
}
