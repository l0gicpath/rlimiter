package buckets

import (
	rl "github.com/juju/ratelimit"
	"rlimiter/conf"
	"time"
)

type Buckets map[string]*rl.Bucket

var PathBuckets Buckets = make(Buckets)

func (b Buckets) GetOrCreate(key string) (bucket *rl.Bucket) {
	bucket, ok := b[key]

	if !ok {
		bucket = rl.NewBucketWithQuantum(60*time.Second, conf.Cfg.RPM, conf.Cfg.RPM)
		b[key] = bucket
	}

	return
}
