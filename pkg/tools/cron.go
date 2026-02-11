package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/sipeed/picoclaw/pkg/cron"
)

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// JobExecutor is the interface for executing cron jobs through the agent
type JobExecutor interface {
	ProcessDirectWithChannel(ctx context.Context, content, sessionKey, channel, chatID string) (string, error)
}

// CronTool provides scheduling capabilities for the agent
type CronTool struct {
	cronService *cron.CronService
	executor    JobExecutor
	channel     string
	chatID      string
	mu          sync.RWMutex
}

// NewCronTool creates a new CronTool
func NewCronTool(cronService *cron.CronService, executor JobExecutor) *CronTool {
	return &CronTool{
		cronService: cronService,
		executor:    executor,
	}
}

// Name returns the tool name
func (t *CronTool) Name() string {
	return "cron"
}

// Description returns the tool description
func (t *CronTool) Description() string {
	return "Schedule reminders and recurring tasks. Actions: add, list, remove, enable, disable."
}

// Parameters returns the tool parameters schema
func (t *CronTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add", "list", "remove", "enable", "disable"},
				"description": "Action to perform",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Reminder message (for add)",
			},
			"every_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Interval in seconds for recurring tasks",
			},
			"cron_expr": map[string]interface{}{
				"type":        "string",
				"description": "Cron expression like '0 9 * * *' for scheduled tasks",
			},
			"job_id": map[string]interface{}{
				"type":        "string",
				"description": "Job ID (for remove/enable/disable)",
			},
		},
		"required": []string{"action"},
	}
}

// SetContext sets the current session context for job creation
func (t *CronTool) SetContext(channel, chatID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.channel = channel
	t.chatID = chatID
}

// Execute runs the tool with given arguments
func (t *CronTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	switch action {
	case "add":
		return t.addJob(args)
	case "list":
		return t.listJobs()
	case "remove":
		return t.removeJob(args)
	case "enable":
		return t.enableJob(args, true)
	case "disable":
		return t.enableJob(args, false)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *CronTool) addJob(args map[string]interface{}) (string, error) {
	t.mu.RLock()
	channel := t.channel
	chatID := t.chatID
	t.mu.RUnlock()

	if channel == "" || chatID == "" {
		return "Error: no session context (channel/chat_id not set). Use this tool in an active conversation.", nil
	}

	message, ok := args["message"].(string)
	if !ok || message == "" {
		return "Error: message is required for add", nil
	}

	var schedule cron.CronSchedule

	// Check for every_seconds
	everySeconds, hasEvery := args["every_seconds"].(float64)
	cronExpr, hasCron := args["cron_expr"].(string)

	if !hasEvery && !hasCron {
		return "Error: either every_seconds or cron_expr is required", nil
	}

	if hasEvery {
		everyMS := int64(everySeconds) * 1000
		schedule = cron.CronSchedule{
			Kind:    "every",
			EveryMS: &everyMS,
		}
	} else {
		schedule = cron.CronSchedule{
			Kind: "cron",
			Expr: cronExpr,
		}
	}

	job, err := t.cronService.AddJob(
		truncateString(message, 30),
		schedule,
		message,
		true, // deliver
		channel,
		chatID,
	)
	if err != nil {
		return fmt.Sprintf("Error adding job: %v", err), nil
	}

	return fmt.Sprintf("Created job '%s' (id: %s)", job.Name, job.ID), nil
}

func (t *CronTool) listJobs() (string, error) {
	jobs := t.cronService.ListJobs(false)

	if len(jobs) == 0 {
		return "No scheduled jobs.", nil
	}

	result := "Scheduled jobs:\n"
	for _, j := range jobs {
		var scheduleInfo string
		if j.Schedule.Kind == "every" && j.Schedule.EveryMS != nil {
			scheduleInfo = fmt.Sprintf("every %ds", *j.Schedule.EveryMS/1000)
		} else if j.Schedule.Kind == "cron" {
			scheduleInfo = j.Schedule.Expr
		} else if j.Schedule.Kind == "at" {
			scheduleInfo = "one-time"
		} else {
			scheduleInfo = "unknown"
		}
		result += fmt.Sprintf("- %s (id: %s, %s)\n", j.Name, j.ID, scheduleInfo)
	}

	return result, nil
}

func (t *CronTool) removeJob(args map[string]interface{}) (string, error) {
	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return "Error: job_id is required for remove", nil
	}

	if t.cronService.RemoveJob(jobID) {
		return fmt.Sprintf("Removed job %s", jobID), nil
	}
	return fmt.Sprintf("Job %s not found", jobID), nil
}

func (t *CronTool) enableJob(args map[string]interface{}, enable bool) (string, error) {
	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return "Error: job_id is required for enable/disable", nil
	}

	job := t.cronService.EnableJob(jobID, enable)
	if job == nil {
		return fmt.Sprintf("Job %s not found", jobID), nil
	}

	status := "enabled"
	if !enable {
		status = "disabled"
	}
	return fmt.Sprintf("Job '%s' %s", job.Name, status), nil
}

// ExecuteJob executes a cron job through the agent
func (t *CronTool) ExecuteJob(ctx context.Context, job *cron.CronJob) string {
	// Get channel/chatID from job payload
	channel := job.Payload.Channel
	chatID := job.Payload.To

	// Default values if not set
	if channel == "" {
		channel = "cli"
	}
	if chatID == "" {
		chatID = "direct"
	}

	sessionKey := fmt.Sprintf("cron-%s", job.ID)

	// Call agent with the job's message
	response, err := t.executor.ProcessDirectWithChannel(
		ctx,
		job.Payload.Message,
		sessionKey,
		channel,
		chatID,
	)

	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Response is automatically sent via MessageBus by AgentLoop
	_ = response // Will be sent by AgentLoop
	return "ok"
}
