/*
Buckets contains a wrapping implementation of tokenbuckets.
*/
package buckets

import (
	rl "github.com/juju/ratelimit"
	"github.com/l0gicpath/rlimiter/conf"
	"time"
)

type Buckets map[string]*rl.Bucket

var PathBuckets Buckets = make(Buckets)

// GetOrCreate will create a new bucket with the given key if it doesn't exist
// or return an existing one.
//
// Configuration for new buckets are taken from the global configuration variable
// conf.Cfg and the only configuration required is RPM (Requests Per Minute)
//
// Returns a bucket pointer
func (b Buckets) GetOrCreate(key string) (bucket *rl.Bucket) {
	bucket, ok := b[key]

	if !ok {
		bucket = rl.NewBucketWithQuantum(60*time.Second, conf.Cfg.RPM, conf.Cfg.RPM)
		b[key] = bucket
	}

	return
}
