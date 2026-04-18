package agent_test

import (
	"testing"

	"tenq-interview/internal/agent"
)

func TestShouldRunDeepSeekIntegrationRequiresOptIn(t *testing.T) {
	t.Setenv("TENQ_RUN_DEEPSEEK_INTEGRATION", "")
	t.Setenv("DEEPSEEK_API_KEY", "test-key")

	if agent.ShouldRunDeepSeekIntegration() {
		t.Fatalf("expected deepseek integration test to stay disabled without explicit opt-in")
	}
}
