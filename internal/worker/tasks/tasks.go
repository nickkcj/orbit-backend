package tasks

// Task type constants
const (
	TypeSendNotification = "notification:send"
	TypeProcessWebhook   = "webhook:process"
	TypeSendEmail        = "email:send"
	TypeProcessVideo     = "video:process"
)

// Queue names with priorities
const (
	QueueCritical = "critical" // Webhooks, time-sensitive
	QueueDefault  = "default"  // Notifications
	QueueLow      = "low"      // Analytics, batch jobs
)
