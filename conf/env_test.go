package conf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnvironmentEnums(t *testing.T) {
	assert.Equal(t, DevelopmentEnv.String(), "Development")
	assert.Equal(t, ProductionEnv.String(), "Production")
	assert.Equal(t, TestEnv.String(), "Test")
}
