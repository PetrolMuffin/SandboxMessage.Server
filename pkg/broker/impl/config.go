package broker

import (
	"time"
)

type Config struct {
	consume ConsumersConfig
	stream  StreamsConfig
}

type StreamsConfig struct {
	workerName    string
	duplicates    time.Duration
	dlqDuplicates time.Duration
	replicas      int
}

type ConsumersConfig struct {
	defaultRetries  []time.Duration
	defaultTimeout  time.Duration
	consumers       map[string]ConsumerConfig
	defaultPrefetch uint16
}

type ConsumerConfig struct {
	retries  []time.Duration
	stream   string
	timeout  time.Duration
	prefetch uint16
}

func (c *Config) Prefetch(subject string) uint16 {
	prefetch := c.consume.consumers[subject].prefetch
	if prefetch == 0 {
		prefetch = c.consume.defaultPrefetch
	}

	return prefetch
}

func (c *Config) Timeout(subject string) time.Duration {
	timeout := c.consume.consumers[subject].timeout
	if timeout == 0 {
		timeout = c.consume.defaultTimeout
	}

	return timeout
}

func (c *Config) BackOff(subject string) []time.Duration {
	retries := c.consume.consumers[subject].retries
	if len(retries) == 0 {
		retries = c.consume.defaultRetries
	}

	return retries
}
