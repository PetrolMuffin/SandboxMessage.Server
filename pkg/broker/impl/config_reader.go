package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	retryDelaySeparator = "x"
)

func ReadConfig(ctx context.Context, data []byte, replicas int) (Config, error) {
	var config rawConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	consumers := make(map[string]ConsumerConfig)
	for subject, rawConsumerConfig := range config.Consume.Consumers {
		consumers[subject] = ConsumerConfig{
			retries:  flat(rawConsumerConfig.Retries),
			stream:   rawConsumerConfig.Stream,
			timeout:  time.Duration(rawConsumerConfig.Timeout),
			prefetch: rawConsumerConfig.Prefetch,
		}
	}

	consume := ConsumersConfig{
		defaultTimeout:  time.Duration(config.Consume.DefaultTimeout),
		defaultPrefetch: config.Consume.DefaultPrefetch,
		defaultRetries:  flat(config.Consume.DefaultRetries),
		consumers:       consumers,
	}

	stream := StreamsConfig{
		workerName:    config.Stream.WorkerName,
		duplicates:    time.Duration(config.Stream.Duplicates),
		dlqDuplicates: time.Duration(config.Stream.DlqDuplicates),
		replicas:      config.Stream.Replicas,
	}

	return Config{
		stream:  stream,
		consume: consume,
	}, nil
}

type rawConfig struct {
	Consume rawConsumersConfig `json:"consume" validate:"required"`
	Stream  rawStreamConfig    `json:"stream" validate:"required"`
}

type rawStreamConfig struct {
	WorkerName    string   `json:"worker_name" validate:"required,min=1"`
	Duplicates    duration `json:"duplicates" validate:"required,gt=0"`
	DlqDuplicates duration `json:"dlq_duplicates" validate:"required,gt=0"`
	Replicas      int      `json:"replicas" validate:"required,gt=0"`
}

type rawConsumersConfig struct {
	DefaultRetries  []retry                      `json:"default_retries" validate:"required,min=1"`
	DefaultTimeout  duration                     `json:"default_timeout" validate:"required,gt=0"`
	Consumers       map[string]rawConsumerConfig `json:"consumers"`
	DefaultPrefetch uint16                       `json:"default_prefetch" validate:"required,gt=0"`
}

type rawConsumerConfig struct {
	Retries  []retry  `json:"retries"`
	Stream   string   `json:"stream"`
	Timeout  duration `json:"timeout"`
	Prefetch uint16   `json:"prefetch"`
}

func flat(retries []retry) []time.Duration {
	flat := make([]time.Duration, 0, len(retries))
	for _, retry := range retries {
		for i := uint16(0); i < retry.Count; i++ {
			flat = append(flat, time.Duration(retry.Delay))
		}
	}
	return flat
}

type retry struct {
	Delay duration
	Count uint16
}

func (r *retry) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	retryParams := strings.Split(raw, retryDelaySeparator)
	if len(retryParams) > 2 {
		return fmt.Errorf("invalid retry format: %s", raw)
	}

	delay, err := time.ParseDuration(retryParams[0])
	if err != nil {
		return err
	}
	r.Delay = duration(delay)

	if len(retryParams) == 1 {
		r.Count = 1
		return nil
	}

	count, err := strconv.ParseUint(retryParams[1], 10, 16)
	if err != nil {
		return err
	}
	r.Count = uint16(count)

	return nil
}

type duration time.Duration

func (d *duration) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}

	*d = duration(parsed)
	return nil
}
