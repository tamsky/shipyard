package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNomadJobCreatesCorrectly(t *testing.T) {
	c, _, cleanup := setupTestConfig(t, nomadJobDefault)
	defer cleanup()

	cl, err := c.FindResource("nomad_job.test")
	assert.NoError(t, err)

	assert.Equal(t, "test", cl.Info().Name)
	assert.Equal(t, TypeNomadJob, cl.Info().Type)
	assert.Equal(t, PendingCreation, cl.Info().Status)
}

func TestNomadJobSetsDisabled(t *testing.T) {
	c, _, cleanup := setupTestConfig(t, nomadJobDisabled)
	defer cleanup()

	cl, err := c.FindResource("nomad_job.test")
	assert.NoError(t, err)

	assert.Equal(t, Disabled, cl.Info().Status)
}

const nomadJobDefault = `
network "test" {
	subnet = "10.0.0.0/24"
}

nomad_job "test" {
	cluster = "nomad_cluster.dc1"
	paths = []
}
`

const nomadJobDisabled = `
network "test" {
	subnet = "10.0.0.0/24"
}

nomad_job "test" {
	disabled = true
	cluster = "nomad_cluster.dc1"
	paths = []
}
`
