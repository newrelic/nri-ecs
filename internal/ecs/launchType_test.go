package ecs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrelic/nri-ecs/internal/ecs"
)

func TestNewFargateLaunchType(t *testing.T) {
	testCases := []struct {
		isFargate  bool
		launchType string
	}{
		{
			isFargate:  true,
			launchType: "fargate",
		},
		{
			isFargate:  false,
			launchType: "ec2",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.launchType, func(t *testing.T) {
			assert.Equal(t, ecs.NewLaunchType(testCase.isFargate), ecs.LaunchType(testCase.launchType))
		})
	}
}
