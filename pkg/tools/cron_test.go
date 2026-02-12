package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/cron"
)

// MockCronService is a minimal mock of cron.CronService for testing
type MockCronService struct {
	jobs []string
}

func (m *MockCronService) AddJob(name string, schedule cron.CronSchedule, message string, deliver bool, channel, to string) (*cron.CronJob, error) {
	job := &cron.CronJob{
		ID:       "test-id",
		Name:     name,
		Schedule: schedule,
		Payload: cron.CronPayload{
			Message:  message,
			Deliver:  deliver,
			Channel: channel,
			To:      to,
		},
		Enabled: true,
	}
	m.jobs = append(m.jobs, name)
	return job, nil
}

func (m *MockCronService) ListJobs(includeDisabled bool) []*cron.CronJob {
	var result []*cron.CronJob
	for _, name := range m.jobs {
		result = append(result, &cron.CronJob{
			ID:       "test-id-" + name,
			Name:     name,
			Schedule: cron.CronSchedule{Kind: "every", EveryMS: func() *int64 { return nil }},
			Payload:  cron.CronPayload{},
			Enabled: true,
		})
	}
	return result
}

func (m *MockCronService) RemoveJob(jobID string) bool {
	for i, job := range m.jobs {
		if job.ID == jobID {
			m.jobs = append(m.jobs[:i], m.jobs[i+1:]...)
			return true
		}
	}
	return false
}

func (m *MockCronService) EnableJob(jobID string, enable bool) *cron.CronJob {
	for _, job := range m.jobs {
		if job.ID == jobID {
			job.Enabled = enable
			return job
		}
	}
	return nil
}

// TestCronTool_BasicIntegration provides basic integration testing for CronTool
func TestCronTool_BasicIntegration(t *testing.T) {
	mockService := &MockCronService{jobs: []string{}}
	msgBus := bus.NewMessageBus()

	tool := NewCronTool(mockService, nil, msgBus)
	tool.SetContext("test-channel", "test-chat")

	ctx := context.Background()

	// Test 1: Add job with at_seconds (one-time) - should return SilentResult
	t.Run("AddJob_OneTime", func(t *testing.T) {
		args := map[string]interface{}{
			"action":     "add",
			"message":   "test message",
			"at_seconds": float64(600),
			"deliver":   false,
		}
		result := tool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("Expected success, got IsError=true: %s", result.ForLLM)
		}

		if !result.Silent {
			t.Errorf("Expected SilentResult, got silent=%v", result.Silent)
		}

		if !strings.Contains(result.ForLLM, "one-time") {
			t.Errorf("Expected ForLLM to contain 'one-time', got: %s", result.ForLLM)
		}
	})

	// Test 2: Add job with every_seconds (recurring) - should return SilentResult
	t.Run("AddJob_Recurring", func(t *testing.T) {
		args := map[string]interface{}{
			"action":     "add",
			"message":   "recurring test",
			"every_seconds": float64(3600),
			"deliver":   true,
		}
		result := tool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("Expected success, got IsError=true: %s", result.ForLLM)
		}

		if !result.Silent {
			t.Errorf("Expected SilentResult, got silent=%v", result.Silent)
		}

		if !strings.Contains(result.ForLLM, "recurring") {
			t.Errorf("Expected ForLLM to contain 'recurring', got: %s", result.ForLLM)
		}
	})

	// Test 3: Add job with cron_expr (complex recurring) - should return SilentResult
	t.Run("AddJob_CronExpr", func(t *testing.T) {
		args := map[string]interface{}{
			"action":     "add",
			"message":   "complex recurring task",
			"cron_expr": "0 9 * * * *",
			"deliver":   true,
		}
		result := tool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("Expected success, got IsError=true: %s", result.ForLLM)
		}

		if !result.Silent {
			t.Errorf("Expected SilentResult, got silent=%v", result.Silent)
		}

		if !strings.Contains(result.ForLLM, "complex") {
			t.Errorf("Expected ForLLM to contain 'complex', got: %s", result.ForLLM)
		}
	})

	// Test 4: List jobs - should return SilentResult with job list
	t.Run("ListJobs", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "list",
		}
		result := tool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("Expected success, got IsError=true: %s", result.ForLLM)
		}

		if !result.Silent {
			t.Errorf("Expected SilentResult, got silent=%v", result.Silent)
		}

		// Verify ForLLM contains job count and one job name
		if !strings.Contains(result.ForLLM, "1 jobs") {
			t.Errorf("Expected ForLLM to contain '1 jobs', got: %s", result.ForLLM)
		}
	})

	// Test 5: Remove job
	t.Run("RemoveJob", func(t *testing.T) {
		// First add a job to remove
		addArgs := map[string]interface{}{
			"action":     "add",
			"message":   "temp job",
			"at_seconds": float64(60),
			"deliver":   false,
		}
		tool.Execute(ctx, addArgs)

		// Now try to remove it
		args := map[string]interface{}{
			"action":   "remove",
			"job_id":  "test-id-temp job",
		}
		result := tool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("Expected success removing job, got IsError=true: %s", result.ForLLM)
		}

		if !result.Silent {
			t.Errorf("Expected SilentResult, got silent=%v", result.Silent)
		}
	})

	// Test 6: Enable/disable job
	t.Run("EnableJob", func(t *testing.T) {
		// First add a job
		addArgs := map[string]interface{}{
			"action":     "add",
			"message":   "test job",
			"at_seconds": float64(120),
			"deliver":   false,
		}
		tool.Execute(ctx, addArgs)

		// Now disable it
		args := map[string]interface{}{
			"action":   "enable",
			"job_id":  "test-id-test job",
		}
		result := tool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("Expected success enabling job, got IsError=true: %s", result.ForLLM)
		}

		if !result.Silent {
			t.Errorf("Expected SilentResult, got silent=%v", result.Silent)
		}
	})

	// Test 7: Missing action parameter
	t.Run("MissingAction", func(t *testing.T) {
		args := map[string]interface{}{
			"message": "test without action",
		}
		result := tool.Execute(ctx, args)

		if !result.IsError {
			t.Errorf("Expected error for missing action, got IsError=false", result.ForLLM)
		}

		if result.Silent {
			t.Errorf("Expected non-silent for error case, got silent=%v", result.Silent)
		}
	})

	// Test 8: Missing parameters
	t.Run("MissingParameters", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "add",
		}
		result := tool.Execute(ctx, args)

		if !result.IsError {
			t.Errorf("Expected error for missing parameters, got IsError=false", result.ForLLM)
		}

		if !strings.Contains(result.ForLLM, "message is required") {
			t.Errorf("Expected ForLLM to contain 'message is required', got: %s", result.ForLLM)
		}
	})

	// Test 9: Job not found
	t.Run("JobNotFound", func(t *testing.T) {
		args := map[string]interface{}{
			"action":   "remove",
			"job_id": "nonexistent",
		}
		result := tool.Execute(ctx, args)

		if !result.IsError {
			t.Errorf("Expected error removing nonexistent job, got IsError=false", result.ForLLM)
		}

		if !strings.Contains(result.ForLLM, "not found") {
			t.Errorf("Expected ForLLM to contain 'not found', got: %s", result.ForLLM)
		}
	})

	// Test 10: No session context
	t.Run("NoSessionContext", func(t *testing.T) {
		tool2 := NewCronTool(mockService, nil, nil)

		ctx := context.Background()
		args := map[string]interface{}{
			"action":   "add",
			"message": "test",
			"at_seconds": float64(600),
		}

		result := tool2.Execute(ctx, args)

		if !result.IsError {
			t.Errorf("Expected error when session not set, got IsError=false", result.ForLLM)
		}

		if !strings.Contains(result.ForLLM, "no session context") {
			t.Errorf("Expected ForLLM to contain 'no session context', got: %s", result.ForLLM)
		}
	})
}
