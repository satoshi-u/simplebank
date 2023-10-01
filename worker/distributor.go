package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

/*
   Create task and distribute them to workers via Redis queue
*/

const TaskSendVerifyEmail = "task:send_verify_email"

// Makes code more generic & easier to mock and test
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error
}

// RedisTaskDistributor implements TaskDistributor
type RedisTaskDistributor struct {
	client *asynq.Client
}

// interface as return type - forcing RedisTaskDistributor to implement TaskDistributor, or else compiler complains
func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}

// DistributeTaskSendVerifyEmail - create SendVerifyEmail task and send to redis queue
func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	// serialize payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// create task
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	// send task to queue
	tInfo, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// log enqueued task details
	log.Info().
		Str("type", tInfo.Type).
		RawJSON("payload", tInfo.Payload).
		Str("queue", tInfo.Queue).
		Int("max_retry", tInfo.MaxRetry).
		Msg("enqueued task")

	return nil
}
