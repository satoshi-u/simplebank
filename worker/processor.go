package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	db "github.com/web3dev6/simplebank/db/sqlc"
)

/*
   Pick up task from Redis queue & process them
   Note* Must register task@TaskSendVerifyEmail with asynq server,
        this tells asynq - the task has to be run by which handler function
        code for this task-registration written in Start()
*/

// Makes code more generic & easier to mock and test
type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor implements TaskProcessor
type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

// interface as return type - forcing RedisTaskProcessor to implement TaskProcessor
func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	// asynq.Config{} allows us to control many different parameters of the asynq server
	server := asynq.NewServer(redisOpt, asynq.Config{})
	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}

// ProcessTaskSendVerifyEmail - processes the read  SendVerifyEmail task
// Note* asynq has already taken care of the core part of pulling task from Redis & feed it to the bg-worker to process it via the task parameter of below handler func
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	// parse the task to get the payload - deserialize
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		// Note* if not unmarshalable, no point of re-trying, tell the same to asynq by wrapping the asynq.SkipRetry error
		return fmt.Errorf("failed to unmarshal task payload: %w", asynq.SkipRetry)
	}

	// process the task - Get user from db and send welcome email
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
			// user doesn't exist, no need to retry
			return fmt.Errorf("user with username %s doesn't exist: %w", payload.Username, asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user with username %s: %w", payload.Username, err)
	}
	// todo send email here

	// log processed task details
	log.Info().
		Str("type", task.Type()).
		RawJSON("payload", task.Payload()).
		Str("user_email", user.Email).
		Msg("processed task")

	return nil
}

// Start - we will register the task@TaskSendVerifyEmail  in this func before starting the asynq server
func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	// we can use this mux to register each task with its handler function, similar to http-mux
	// Register @TaskSendVerifyEmail
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	// start server
	return processor.server.Start(mux)
}
