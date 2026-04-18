package agent

import "os"

const DeepSeekIntegrationEnv = "TENQ_RUN_DEEPSEEK_INTEGRATION"

func ShouldRunDeepSeekIntegration() bool {
	return os.Getenv(DeepSeekIntegrationEnv) == "1"
}
