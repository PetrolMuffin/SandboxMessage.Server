package broker

import "errors"

var (
	ErrConsumerRunFailed = errors.New("consumer run failed")
	ErrFailedToMarshal   = errors.New("failed to marshal")
	ErrFailedToUnmarshal = errors.New("failed to unmarshal")
	ErrDuplicatePublish  = errors.New("duplicate publish")
	ErrPublishFailed     = errors.New("publish failed")
	ErrRequestFailed     = errors.New("request failed")
	ErrQueueRequired     = errors.New("queue required")
	ErrInvalidRequest    = errors.New("invalid request")
)
