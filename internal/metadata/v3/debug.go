package v3

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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

func verifyTaskMetadataResponse(taskMetadataRawMsg json.RawMessage) error {
	var err error
	taskMetadataResponseMap := make(map[string]json.RawMessage)
	json.Unmarshal(taskMetadataRawMsg, &taskMetadataResponseMap)

	taskExpectedFieldEqualMap := map[string]interface{}{
		"DesiredStatus": "RUNNING",
		"KnownStatus":   "RUNNING",
	}

	taskExpectedFieldNotEmptyArray := []string{"Cluster", "TaskARN", "Family", "Revision", "Containers"}
	if checkContainerInstanceTags {
		taskExpectedFieldNotEmptyArray = append(taskExpectedFieldNotEmptyArray, "ContainerInstanceTags")
	}
	taskWarningFieldNotEmptyArray := []string{"PullStartedAt", "PullStoppedAt", "AvailabilityZone"}

	for fieldName, fieldVal := range taskExpectedFieldEqualMap {
		if err = fieldEqual(taskMetadataResponseMap, fieldName, fieldVal); err != nil {
			return err
		}
	}

	for _, fieldName := range taskExpectedFieldNotEmptyArray {
		if err = fieldNotEmpty(taskMetadataResponseMap, fieldName); err != nil {
			return err
		}
	}

	for _, fieldName := range taskWarningFieldNotEmptyArray {
		if err = fieldNotEmpty(taskMetadataResponseMap, fieldName); err != nil {
			fmt.Fprintf(os.Stderr, "unable to get optional container metadata: field %s not found\n", fieldName)
			return nil
		}
	}

	var containersMetadataResponseArray []json.RawMessage
	json.Unmarshal(taskMetadataResponseMap["Containers"], &containersMetadataResponseArray)

	if isAWSVPCNetworkMode {
		if len(containersMetadataResponseArray) != 2 {
			return fmt.Errorf("incorrect number of containers, expected 2, received %d",
				len(containersMetadataResponseArray))
		}

		ok, err := isPauseContainer(containersMetadataResponseArray[0])
		if err != nil {
			return err
		}
		if ok {
			return verifyContainerMetadataResponse(containersMetadataResponseArray[1])
		} else {
			return verifyContainerMetadataResponse(containersMetadataResponseArray[0])
		}
	} else {
		if len(containersMetadataResponseArray) != 1 {
			return fmt.Errorf("incorrect number of containers, expected 1, received %d",
				len(containersMetadataResponseArray))
		}

		return verifyContainerMetadataResponse(containersMetadataResponseArray[0])
	}

	return nil
}

func isPauseContainer(containerMetadataRawMsg json.RawMessage) (bool, error) {
	var err error
	containerMetadataResponseMap := make(map[string]json.RawMessage)
	json.Unmarshal(containerMetadataRawMsg, &containerMetadataResponseMap)

	if err = fieldNotEmpty(containerMetadataResponseMap, "Name"); err != nil {
		return false, err
	}

	var actualContainerName string
	json.Unmarshal(containerMetadataResponseMap["Name"], &actualContainerName)

	if actualContainerName == "~internal~ecs~pause" {
		return true, nil
	}

	return false, nil
}

func verifyContainerMetadataResponse(containerMetadataRawMsg json.RawMessage) error {
	var err error
	containerMetadataResponseMap := make(map[string]json.RawMessage)
	json.Unmarshal(containerMetadataRawMsg, &containerMetadataResponseMap)

	containerExpectedFieldEqualMap := map[string]interface{}{
		// Asserting on the container and image names is wya too strong. This can change a lot.
		// "Name":          "v3-task-endpoint-validator",
		// "Image":         "127.0.0.1:51670/amazon/amazon-ecs-v3-task-endpoint-validator:latest",
		"DesiredStatus": "RUNNING",
		"KnownStatus":   "RUNNING",
		"Type":          "NORMAL",
	}

	taskExpectedFieldNotEmptyArray := []string{"DockerId", "DockerName", "ImageID", "Limits", "CreatedAt", "StartedAt", "Networks"}
	taskWarningFieldNotEmptyArray := []string{"Health", "StartedAt", "Networks"}

	for fieldName, fieldVal := range containerExpectedFieldEqualMap {
		if err = fieldEqual(containerMetadataResponseMap, fieldName, fieldVal); err != nil {
			return err
		}
	}

	for _, fieldName := range taskExpectedFieldNotEmptyArray {
		if err = fieldNotEmpty(containerMetadataResponseMap, fieldName); err != nil {
			return err
		}
	}

	for _, fieldName := range taskWarningFieldNotEmptyArray {
		if err = fieldNotEmpty(containerMetadataResponseMap, fieldName); err != nil {
			fmt.Fprintf(os.Stderr, "unable to get optional container metadata: field %s not found\n", fieldName)
			return nil
		}
	}

	if err = verifyLimitResponse(containerMetadataResponseMap["Limits"]); err != nil {
		return err
	}
	if err = verifyNetworksResponse(containerMetadataResponseMap["Networks"]); err != nil {
		return err
	}

	return nil
}

func verifyLimitResponse(limitRawMsg json.RawMessage) error {
	var err error
	limitResponseMap := make(map[string]json.RawMessage)
	json.Unmarshal(limitRawMsg, &limitResponseMap)

	limitExpectedFieldEqualMap := map[string]interface{}{
		"CPU":    float64(0),
		"Memory": float64(50),
	}

	for fieldName, fieldVal := range limitExpectedFieldEqualMap {
		if err = fieldEqual(limitResponseMap, fieldName, fieldVal); err != nil {
			return err
		}
	}

	return nil
}

func verifyNetworksResponse(networksRawMsg json.RawMessage) error {
	var err error

	var networksResponseArray []json.RawMessage
	json.Unmarshal(networksRawMsg, &networksResponseArray)

	if len(networksResponseArray) == 1 {
		networkResponseMap := make(map[string]json.RawMessage)
		json.Unmarshal(networksResponseArray[0], &networkResponseMap)

		var actualFieldVal interface{}
		json.Unmarshal(networkResponseMap["NetworkMode"], &actualFieldVal)

		if _, ok := networkModes[actualFieldVal.(string)]; !ok {
			return errors.Errorf("network mode is incorrect: %s", actualFieldVal)
		}
		if actualFieldVal != "host" {
			if err = fieldNotEmpty(networkResponseMap, "IPv4Addresses"); err != nil {
				return err
			}

			var ipv4AddressesResponseArray []json.RawMessage
			json.Unmarshal(networkResponseMap["IPv4Addresses"], &ipv4AddressesResponseArray)

			if len(ipv4AddressesResponseArray) != 1 {
				return fmt.Errorf("incorrect number of IPv4Addresses, expected 1, received %d",
					len(ipv4AddressesResponseArray))
			}
		}
		if actualFieldVal == "awsvpc" {
			isAWSVPCNetworkMode = true
		} else if actualFieldVal == "bridge" {
			isBridgeNetworkMode = true
		}
	} else {
		return fmt.Errorf("incorrect number of networks, expected 1, received %d",
			len(networksResponseArray))
	}

	return nil
}

func fieldNotEmpty(rawMsgMap map[string]json.RawMessage, fieldName string) error {
	if rawMsgMap[fieldName] == nil {
		return fmt.Errorf("field %s should not be empty", fieldName)
	}
	return nil
}

func fieldEqual(rawMsgMap map[string]json.RawMessage, fieldName string, fieldVal interface{}) error {
	if err := fieldNotEmpty(rawMsgMap, fieldName); err != nil {
		return err
	}

	var actualFieldVal interface{}
	json.Unmarshal(rawMsgMap[fieldName], &actualFieldVal)

	if fieldVal != actualFieldVal {
		return fmt.Errorf("incorrect field value for field %s, expected %v, received %v",
			fieldName, fieldVal, actualFieldVal)
	}

	return nil
}

func DebugECSEndpoint() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	networkModes = map[string]bool{"awsvpc": true, "bridge": true, "host": true, "default": true}

	// If the image is built with option to check Tags
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		if argsWithoutProg[0] == "CheckTags" {
			checkContainerInstanceTags = true
		}
	}

	// Wait for the Health information to be ready
	time.Sleep(5 * time.Second)

	isAWSVPCNetworkMode = false
	isBridgeNetworkMode = false
	v3BaseEndpoint, found := MetadataV3Endpoint()
	if !found {
		fmt.Fprint(os.Stderr, "Unable to get URL for metadata v3 endpoint")
		os.Exit(1)
	}
	containerMetadataPath := v3BaseEndpoint
	taskMetadataPath := v3BaseEndpoint
	if checkContainerInstanceTags {
		taskMetadataPath += "/taskWithTags"
	} else {
		taskMetadataPath += "/task"
	}
	containerStatsPath := v3BaseEndpoint + "/stats"
	taskStatsPath := v3BaseEndpoint + "/task/stats"

	if err := verifyContainerMetadata(client, containerMetadataPath); err != nil {
		fmt.Fprintf(os.Stderr, "container metadata validation failed: %v\n", err)
		os.Exit(1)
	}

	if err := verifyTaskMetadata(client, taskMetadataPath); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get task metadata: %v\n", err)
		os.Exit(1)
	}

	if err := verifyContainerStats(client, containerStatsPath); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get container stats: %v\n", err)
		os.Exit(1)
	}

	if err := verifyTaskStats(client, taskStatsPath); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get task stats: %v\n", err)
		os.Exit(1)
	}

	os.Exit(42)
}

func verifyContainerMetadata(client *http.Client, containerMetadataEndpoint string) error {
	var err error
	body, err := MetadataResponse(client, containerMetadataEndpoint)
	if err != nil {
		return err
	}

	fmt.Printf("Received container metadata: %s \n", string(body))

	var containerMetadata ContainerResponse
	if err = json.Unmarshal(body, &containerMetadata); err != nil {
		return fmt.Errorf("unable to parse response body: %v", err)
	}

	if err = verifyContainerMetadataResponse(body); err != nil {
		return err
	}

	return nil
}

func verifyTaskMetadata(client *http.Client, taskMetadataEndpoint string) error {
	body, err := MetadataResponse(client, taskMetadataEndpoint)
	if err != nil {
		return err
	}

	fmt.Printf("Received task metadata: %s \n", string(body))

	var taskMetadata TaskResponse
	if err = json.Unmarshal(body, &taskMetadata); err != nil {
		return fmt.Errorf("unable to parse response body: %v", err)
	}

	if err = verifyTaskMetadataResponse(body); err != nil {
		return err
	}

	return nil
}

func verifyContainerStats(client *http.Client, containerStatsEndpoint string) error {
	body, err := MetadataResponse(client, containerStatsEndpoint)
	if err != nil {
		return err
	}

	fmt.Printf("Received container stats: %s \n", string(body))

	var containerStats types.StatsJSON
	err = json.Unmarshal(body, &containerStats)
	if err != nil {
		return fmt.Errorf("container stats: unable to parse response body: %v", err)
	}

	if isBridgeNetworkMode {
		// networks field should be populated in bridge mode
		if containerStats.Networks == nil {
			return errors.New("container stats: field networks should not be empty")
		}
	}

	return nil
}

func verifyTaskStats(client *http.Client, taskStatsEndpoint string) error {
	body, err := MetadataResponse(client, taskStatsEndpoint)
	if err != nil {
		return err
	}

	fmt.Printf("Received task stats: %s \n", string(body))

	var taskStats map[string]*types.StatsJSON
	err = json.Unmarshal(body, &taskStats)
	if err != nil {
		return fmt.Errorf("task stats: unable to parse response body: %v", err)
	}

	if isBridgeNetworkMode {
		for container, containerStats := range taskStats {
			// networks field should be populated in bridge mode
			if containerStats.Networks == nil {
				return fmt.Errorf("task stats: field networks for container %s should not be empty", container)
			}
		}
	}

	return nil
}

