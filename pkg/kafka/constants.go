package kafka

import "time"

const (
	minBytes               = 10e3 // 10KB
	maxBytes               = 10e6 // 10MB
	queueCapacity          = 100
	heartbeatInterval      = 1 * time.Second
	commitInterval         = 0
	partitionWatchInterval = 500 * time.Millisecond
	maxAttempts            = 3
	dialTimeout            = 3 * time.Minute
	maxWait                = 3 * time.Second

	writerReadTimeout  = 1 * time.Second
	writerWriteTimeout = 1 * time.Second
	batchTimeout       = 100 * time.Millisecond
	batchSize          = 100

	writerRequireNoneReadTimeout  = 5 * time.Second
	writerRequireNoneWriteTimeout = 5 * time.Second

	writerRequiredAcks = 1
	writerMaxAttempts  = 5
)
