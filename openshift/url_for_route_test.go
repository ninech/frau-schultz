package openshift_test

import (
	"errors"
	"testing"

	"github.com/ninech/actuator/openshift"
	"github.com/ninech/actuator/test"
	"github.com/stretchr/testify/assert"
)

func TestGetURLForRoute(t *testing.T) {
	sampleRouteExport := `apiVersion: v1
kind: Route
metadata:
  creationTimestamp: null
  name: actuator
spec:
  host: actuator.openshift.nine.ch
  port:
    targetPort: 8080-tcp
  to:
    kind: Service
    name: actuator
    weight: 100
  wildcardPolicy: None
status: {}`

	t.Run("when the command works", func(t *testing.T) {
		shell := &test.MockShell{OutputToReturn: sampleRouteExport}
		openshiftClient := openshift.CommandLineClient{CommandExecutor: shell}

		url, _ := openshiftClient.GetURLForRoute("actuator")
		assert.Equal(t, "http://actuator.openshift.nine.ch", url)
	})

	t.Run("when there is no such route", func(t *testing.T) {
		shell := &test.MockShell{ErrorToReturn: errors.New(`Error from server (NotFound): routes "actuator" not found`)}
		openshiftClient := openshift.CommandLineClient{CommandExecutor: shell}

		url, err := openshiftClient.GetURLForRoute("actuator")
		assert.Empty(t, url)
		assert.NotNil(t, err)
	})

	t.Run("when the yaml is not valid", func(t *testing.T) {
		shell := &test.MockShell{OutputToReturn: "12345"}
		openshiftClient := openshift.CommandLineClient{CommandExecutor: shell}

		url, err := openshiftClient.GetURLForRoute("actuator")
		assert.Empty(t, url)
		assert.NotNil(t, err)
	})
}
