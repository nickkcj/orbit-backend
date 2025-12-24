package worker

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
	"github.com/nickkcj/orbit-backend/internal/service"
	"github.com/nickkcj/orbit-backend/internal/worker/handlers"
	"github.com/nickkcj/orbit-backend/internal/worker/tasks"
)

// Worker manages the Asynq server and task handlers
type Worker struct {
	server   *asynq.Server
	mux      *asynq.ServeMux
	services *service.Services
}

// workerLogger implements asynq.Logger interface
type workerLogger struct{}

func (l *workerLogger) Debug(args ...interface{}) {
	log.Println("[DEBUG]", args)
}

func (l *workerLogger) Info(args ...interface{}) {
	log.Println("[INFO]", args)
}

func (l *workerLogger) Warn(args ...interface{}) {
	log.Println("[WARN]", args)
}

func (l *workerLogger) Error(args ...interface{}) {
	log.Println("[ERROR]", args)
}

func (l *workerLogger) Fatal(args ...interface{}) {
	log.Fatal("[FATAL]", args)
}

// NewWorker creates a new worker with the given configuration
func NewWorker(redisOpt asynq.RedisClientOpt, concurrency int, services *service.Services) *Worker {
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			tasks.QueueCritical: 6,
			tasks.QueueDefault:  3,
			tasks.QueueLow:      1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)

			if retried >= maxRetry {
				log.Printf("[DEAD] Task %s exhausted retries: %v", task.Type(), err)
			} else {
				log.Printf("[RETRY] Task %s failed (attempt %d/%d): %v",
					task.Type(), retried+1, maxRetry, err)
			}
		}),
		Logger: &workerLogger{},
	})

	mux := asynq.NewServeMux()

	// Register handlers
	notificationHandler := handlers.NewNotificationHandler(services.Notification)
	mux.HandleFunc(tasks.TypeSendNotification, notificationHandler.Handle)

	return &Worker{
		server:   srv,
		mux:      mux,
		services: services,
	}
}

// Start begins processing tasks
func (w *Worker) Start() error {
	log.Println("Starting background worker...")
	return w.server.Start(w.mux)
}

// Shutdown gracefully stops the worker
func (w *Worker) Shutdown() {
	log.Println("Shutting down background worker...")
	w.server.Shutdown()
}
