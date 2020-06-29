// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package injector

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/stretchr/testify/assert"
)

func TestLogAsJSONEnabled(t *testing.T) {
	t.Run("dapr.io/log-as-json is true", func(t *testing.T) {
		var fakeAnnotation = map[string]string{
			daprLogAsJSON: "true",
		}

		assert.Equal(t, true, logAsJSONEnabled(fakeAnnotation))
	})

	t.Run("dapr.io/log-as-json is false", func(t *testing.T) {
		var fakeAnnotation = map[string]string{
			daprLogAsJSON: "false",
		}

		assert.Equal(t, false, logAsJSONEnabled(fakeAnnotation))
	})

	t.Run("dapr.io/log-as-json is not given", func(t *testing.T) {
		var fakeAnnotation = map[string]string{}

		assert.Equal(t, false, logAsJSONEnabled(fakeAnnotation))
	})
}

func TestFormatProbePath(t *testing.T) {
	var testCases = []struct {
		given    []string
		expected string
	}{
		{
			given:    []string{"api", "v1"},
			expected: "/api/v1",
		},
		{
			given:    []string{"//api", "v1"},
			expected: "/api/v1",
		},
		{
			given:    []string{"//api", "/v1/"},
			expected: "/api/v1",
		},
		{
			given:    []string{"//api", "/v1/", "healthz"},
			expected: "/api/v1/healthz",
		},
		{
			given:    []string{""},
			expected: "/",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, formatProbePath(tc.given...))
	}
}

func TestGetProbeHttpHandler(t *testing.T) {
	pathElements := []string{"api", "v1", "healthz"}
	expectedPath := "/api/v1/healthz"
	expectedHandler := corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: expectedPath,
			Port: intstr.IntOrString{IntVal: sidecarHTTPPort},
		},
	}

	assert.EqualValues(t, expectedHandler, getProbeHTTPHandler(sidecarHTTPPort, pathElements...))
}

func TestGetSideCarContainer(t *testing.T) {
	annotations := map[string]string{}
	annotations[daprConfigKey] = "config"
	annotations[daprPortKey] = "5000"
	annotations[daprLogAsJSON] = "true"
	annotations[daprAPITokenSecret] = "secret"

	container, _ := getSidecarContainer(annotations, "app_id", "darpio/dapr", "dapr-system", "controlplane:9000", "placement:50000", nil, "", "", "", "sentry:50000", true, "pod_identity")

	var expectedArgs = []string{
		"--mode", "kubernetes",
		"--dapr-http-port", "3500",
		"--dapr-grpc-port", "50001",
		"--dapr-internal-grpc-port", "50002",
		"--app-port", "5000",
		"--app-id", "app_id",
		"--control-plane-address", "controlplane:9000",
		"--protocol", "http",
		"--placement-address", "placement:50000",
		"--config", "config",
		"--log-level", "info",
		"--max-concurrency", "-1",
		"--sentry-address", "sentry:50000",
		"--metrics-port", "9090",
		"--log-as-json",
	}

	assert.Equal(t, "secret", container.Env[2].ValueFrom.SecretKeyRef.Name)
	assert.EqualValues(t, expectedArgs, container.Args)
}