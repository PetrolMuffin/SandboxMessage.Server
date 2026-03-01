package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/petrolmuffin/go-collections/set"
	"github.com/petrolmuffin/sandboxmessage-server/pkg/log"
)

const (
	dlqSuffix = ".dlq"
)

type Broker interface {
	Listen(ctx context.Context, consumers ...anyConsumer) error
}

type NatsBroker struct {
	js        jetstream.JetStream
	waitGroup sync.WaitGroup
	config    *Config
	subjects  set.SyncSet[string]
}

// js Storage: MemoryStorage; Replicas: 3
// примонтировать volume в docker compose. сделать 3 реплики nats
func New(ctx context.Context, js jetstream.JetStream, config *Config) (*NatsBroker, error) {
	if err := ensureStream(ctx, js, config); err != nil {
		return nil, err
	}

	return &NatsBroker{js: js, config: config}, nil
}

func (b *NatsBroker) Close(ctx context.Context) error {
	b.waitGroup.Wait()
	return nil
}

func ensureStream(ctx context.Context, js jetstream.JetStream, config *Config) error {
	tasksStreamName := "task." + config.stream.workerName
	tasksSubject := tasksStreamName + ".>"
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:       tasksStreamName,
		Subjects:   []string{tasksSubject},
		Storage:    jetstream.MemoryStorage,
		Duplicates: config.stream.duplicates,
		Replicas:   config.stream.replicas,
		Retention:  jetstream.WorkQueuePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to ensure tasks stream: %w", err)
	}

	eventsStreamName := "event." + config.stream.workerName
	eventsSubject := eventsStreamName + ".>"
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:       eventsStreamName,
		Subjects:   []string{eventsSubject},
		Storage:    jetstream.MemoryStorage,
		Duplicates: config.stream.duplicates,
		Replicas:   config.stream.replicas,
		Retention:  jetstream.InterestPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to ensure events stream: %w", err)
	}

	dlqStreamName := config.stream.workerName + "." + dlqSuffix
	tasksDlqSubject := tasksSubject + "." + dlqSuffix
	eventsDlqSubject := eventsSubject + "." + dlqSuffix
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:       dlqStreamName,
		Subjects:   []string{tasksDlqSubject, eventsDlqSubject},
		Storage:    jetstream.MemoryStorage,
		Duplicates: config.stream.dlqDuplicates,
		Replicas:   config.stream.replicas,
		MaxAge:     time.Hour * 24 * 14,
		Retention:  jetstream.LimitsPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to ensure dlq stream: %w", err)
	}

	log.CtxLogger(ctx).Info(ctx, "stream ensured", "stream", config.stream.workerName)
	return nil
}
