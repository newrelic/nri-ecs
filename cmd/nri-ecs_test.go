package main_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	main "github.com/newrelic/nri-ecs/cmd"
	"github.com/newrelic/nri-ecs/cmd/testdata"
	"github.com/newrelic/nri-ecs/internal/ecs/metadata"
)

const v3 = `{
	"Cluster": "ecs-local-cluster", 
	"TaskARN": "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	}`

const v4 = `{
	"Cluster": "ecs-local-cluster", 
	"TaskARN": "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152",
	"LaunchType": "EC2"
	}`

func Test_integration_publishes_payload(t *testing.T) {
	args := main.ArgumentList{}

	cases := map[string]struct {
		response       string
		endpointEnvVar string
	}{
		"from_v3_metadata_endpoint": {
			response:       v3,
			endpointEnvVar: metadata.ContainerMetadataEnvVar,
		},
		"from_v4_metadata_endpoint": {
			response:       v4,
			endpointEnvVar: metadata.ContainerMetadataV4EnvVar,
		},
	}

	for testCaseName, testData := range cases {
		testData := testData

		t.Run(testCaseName, func(t *testing.T) {
			ts := metadataServer(t, testData.response)

			t.Setenv(testData.endpointEnvVar, ts.URL)

			integrationPayload := &bytes.Buffer{}

			ecsIntegration, _ := integration.New("com.newrelic.ecs", "0.0.0", integration.Writer(integrationPayload))

			assert.NoError(t, main.Run(ecsIntegration, args))

			assert.JSONEq(t, testdata.IntegrationOutput, integrationPayload.String())
		})
	}
}

func Test_integration_fails_to_run(t *testing.T) {
	args := main.ArgumentList{}

	t.Run("when_no_metadata_endpoint_present", func(t *testing.T) {
		ecsIntegration, _ := integration.New("com.newrelic.ecs", "0.0.0")

		assert.Error(t, main.Run(ecsIntegration, args))
	})

	t.Run("when_no_metadata_endpoint_fails_to_respond", func(t *testing.T) {
		ecsIntegration, _ := integration.New("com.newrelic.ecs", "0.0.0")

		t.Setenv(metadata.ContainerMetadataV4EnvVar, "http://willfail/v4")

		assert.Error(t, main.Run(ecsIntegration, args))
	})

	t.Run("when_no_metadata_endpoint_response_fails_to_unmarshall", func(t *testing.T) {
		ecsIntegration, _ := integration.New("com.newrelic.ecs", "0.0.0")

		ts := metadataServer(t, "will_fail")

		t.Setenv(metadata.ContainerMetadataV4EnvVar, ts.URL)

		assert.Error(t, main.Run(ecsIntegration, args))
	})
}

func metadataServer(t *testing.T, response string) *httptest.Server {
	t.Helper()

	testServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/task", r.RequestURI)

			_, err := w.Write([]byte(response))
			require.NoError(t, err)
		},
	))

	t.Cleanup(func() {
		testServer.Close()
	})

	return testServer
}
