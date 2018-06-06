package buckets

import (
	"github.com/l0gicpath/rlimiter/conf"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOrCreate(t *testing.T) {
	conf.Cfg = &conf.Config{RPM: 50}

	_, ok := PathBuckets["bucket-a"]
	assert.Equal(t, ok, false)

	PathBuckets.GetOrCreate("bucket-a")
	_, ok = PathBuckets["bucket-a"]
	assert.Equal(t, ok, true)

	bucket := PathBuckets.GetOrCreate("bucket-a")
	assert.NotNil(t, bucket)
}
