package agent_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"tenq-interview/internal/agent"
)

// TestDeepSeekRealAPICall 真实调用 DeepSeek API 的集成测试
// 需要 .env 文件中配置有效的 DEEPSEEK_API_KEY
// 运行：go test -run TestDeepSeekRealAPICall -v
func TestDeepSeekRealAPICall(t *testing.T) {
	// 跳过如果 API Key 未配置
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping real API test")
	}

	// 加载配置
	cfg, err := agent.LoadConfigFromEnv("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 创建 DeepSeek Provider
	provider, err := agent.NewProvider(agent.ProviderDeepSeek, cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// 创建 Summarizer
	summarizer := agent.NewSummarizer(provider, agent.PromptVersion)

	// 准备测试请求（使用一个简单的面试问题）
	testTitle := "Go 中的 defer 是什么？"
	testBody := `defer 是 Go 语言中的一个关键字，用于延迟函数执行直到周围函数返回。
defer 语句在以下场景非常有用：
1. 资源清理：关闭文件、释放锁、关闭数据库连接
2. 错误处理：记录日志、恢复 panic
3. 简化代码：将清理代码放在函数开头，而不是结尾

defer 的执行顺序是 LIFO（后进先出），多个 defer 语句按照相反顺序执行。

示例：
func readFile() {
    f, _ := os.Open("file.txt")
    defer f.Close()  // 函数返回前自动关闭文件

    data, _ := io.ReadAll(f)
    return data
}

defer 在返回语句之后但在函数实际返回之前执行，这意味着 defer 可以访问和修改命名返回值。`

	testSegments := []string{
		"defer 是 Go 语言中的一个关键字，用于延迟函数执行直到周围函数返回。",
		"defer 的执行顺序是 LIFO（后进先出），多个 defer 语句按照相反顺序执行。",
		"defer 在返回语句之后但在函数实际返回之前执行，这意味着 defer 可以访问和修改命名返回值。",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("正在调用 DeepSeek API...")
	fmt.Printf("标题：%s\n", testTitle)
	fmt.Printf("正文长度：%d 字符\n", len(testBody))

	// 调用 API
	response, err := summarizer.Summarize(ctx, agent.SummarizeRequest{
		Title:         testTitle,
		Body:          testBody,
		CandidateText: testSegments,
	})

	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}

	// 验证响应
	fmt.Println("\n=== API 响应 ===")
	fmt.Printf("Provider: %s\n", response.Provider)
	fmt.Printf("Model: %s\n", response.Model)
	fmt.Printf("\n标准答案 (%d 字):\n%s\n", len([]rune(response.StandardAnswer)), response.StandardAnswer)

	fmt.Printf("\n记忆提纲 (%d 条):\n", len(response.MemoryOutline))
	for i, outline := range response.MemoryOutline {
		fmt.Printf("  %d. %s\n", i+1, outline)
	}

	fmt.Printf("\n原文引用 (%d 条):\n", len(response.SourceQuotes))
	for i, quote := range response.SourceQuotes {
		fmt.Printf("  %d. %s\n", i+1, quote)
	}

	if response.Notes != "" {
		fmt.Printf("\n备注:\n%s\n", response.Notes)
	}

	// 验证字段非空
	if response.StandardAnswer == "" {
		t.Error("StandardAnswer is empty")
	}
	if len(response.MemoryOutline) == 0 {
		t.Error("MemoryOutline is empty")
	}
	if len(response.SourceQuotes) == 0 {
		t.Error("SourceQuotes is empty")
	}

	fmt.Println("\n✓ DeepSeek API 调用成功！")
}
