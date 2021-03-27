package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerIngressCreatesCorrectly(t *testing.T) {
	c, _, cleanup := setupTestConfig(t, containerIngressDefault)
	defer cleanup()

	co, err := c.FindResource("container_ingress.testing")
	assert.NoError(t, err)

	assert.Equal(t, "testing", co.Info().Name)
	assert.Equal(t, TypeContainerIngress, co.Info().Type)
	assert.Equal(t, PendingCreation, co.Info().Status)
}

func TestContainerIngressSetsDisabled(t *testing.T) {
	c, _, cleanup := setupTestConfig(t, containerIngressDisabled)
	defer cleanup()

	co, err := c.FindResource("container_ingress.testing")
	assert.NoError(t, err)
	assert.Equal(t, Disabled, co.Info().Status)
}

const containerIngressDefault = `
network "test" {
	subnet = "10.0.0.0/24"
}

container_ingress "testing" {
	network {
		name = "network.test"
	}

	target = "container.consul"

}
`

const containerIngressDisabled = `
network "test" {
	subnet = "10.0.0.0/24"
}

container_ingress "testing" {
	disabled = true

	network {
		name = "network.test"
	}
	
	target = "container.consul"
}
`
