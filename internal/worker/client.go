package worker

import (
	"github.com/hibiken/asynq"
)

// TaskClient wraps the Asynq client for task enqueueing
type TaskClient struct {
	client *asynq.Client
}

// NewTaskClient creates a new task client with Redis connection
func NewTaskClient(redisOpt asynq.RedisClientOpt) *TaskClient {
	return &TaskClient{
		client: asynq.NewClient(redisOpt),
	}
}

// Enqueue adds a task to the queue
func (c *TaskClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, opts...)
}

// EnqueueContext adds a task to the queue with context support
func (c *TaskClient) EnqueueContext(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, opts...)
}

// Close closes the client connection
func (c *TaskClient) Close() error {
	return c.client.Close()
}
