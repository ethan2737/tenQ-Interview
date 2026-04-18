# Agent LLM Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为单篇 Markdown 导入接入受约束的 Agent 总结能力，支持 DeepSeek / 魔塔在线 provider 选择，输出标准答案、记忆提纲和原文依据，并保留缓存与降级能力。

**Architecture:** 保留现有 importer、parser、segment 链路，把 card generator 升级为 agent summarizer。新增统一 provider 抽象、prompt 版本化、provider 选择配置和结构化结果解析，前端只增加本次导入级别的 provider 选择器与扩展卡片展示。

**Tech Stack:** Go, Wails, 原生前端 JS, `.env` 配置, 本地 JSON 缓存

---

### Task 1: Agent 配置与 Provider 抽象

**Files:**
- Create: `internal/agent/types.go`
- Create: `internal/agent/config.go`
- Create: `internal/agent/provider.go`
- Create: `internal/agent/providers/deepseek.go`
- Create: `internal/agent/providers/modelscope.go`
- Test: `internal/agent/config_test.go`
- Test: `internal/agent/providers/providers_test.go`

- [ ] **Step 1: 写配置层失败测试**

```go
func TestLoadConfigRequiresApiKeyForEnabledProvider(t *testing.T) {
	t.Setenv("LLM_PROVIDER_DEFAULT", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected missing api key error")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/agent/...`
Expected: FAIL，提示 `LoadConfigFromEnv` 未定义或行为不符合预期

- [ ] **Step 3: 实现最小配置与 provider 接口**

```go
type ProviderName string

const (
	ProviderDeepSeek  ProviderName = "deepseek"
	ProviderModelScope ProviderName = "modelscope"
)

type Config struct {
	DefaultProvider ProviderName
	DeepSeekAPIKey  string
	ModelScopeAPIKey string
}

func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		DefaultProvider: ProviderName(strings.TrimSpace(os.Getenv("LLM_PROVIDER_DEFAULT"))),
		DeepSeekAPIKey: strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		ModelScopeAPIKey: strings.TrimSpace(os.Getenv("MODELSCOPE_API_KEY")),
	}
	if cfg.DefaultProvider == ProviderDeepSeek && cfg.DeepSeekAPIKey == "" {
		return Config{}, errors.New("deepseek api key is required")
	}
	return cfg, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/agent/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/agent docs/superpowers/plans/2026-04-18-agent-llm-provider.md
git commit -m "feat(agent): add provider config skeleton"
```

### Task 2: Prompt、结构化输出与 Summarizer

**Files:**
- Create: `internal/agent/prompt.go`
- Create: `internal/agent/summarizer.go`
- Test: `internal/agent/summarizer_test.go`

- [ ] **Step 1: 写 summarizer 失败测试**

```go
func TestSummarizerBuildsStructuredResult(t *testing.T) {
	provider := stubProvider{
		response: SummarizeResponse{
			StandardAnswer: "这是一个 180 字左右的标准答案。",
			MemoryOutline: []string{"定义", "核心机制", "使用场景"},
			SourceQuotes: []string{"原文依据 1", "原文依据 2"},
		},
	}

	s := NewSummarizer(provider, PromptVersion)
	got, err := s.Summarize(context.Background(), SummarizeRequest{
		Title: "GMP 是什么？",
		Body:  "原文正文",
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if got.StandardAnswer == "" || len(got.MemoryOutline) == 0 {
		t.Fatalf("expected structured summary")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/agent/...`
Expected: FAIL，提示 `NewSummarizer` 未定义

- [ ] **Step 3: 实现 prompt 与 summarizer**

```go
const PromptVersion = "v1"

type Summarizer struct {
	provider Provider
	version  string
}

func NewSummarizer(provider Provider, version string) *Summarizer {
	return &Summarizer{provider: provider, version: version}
}

func (s *Summarizer) Summarize(ctx context.Context, req SummarizeRequest) (SummarizeResponse, error) {
	req.PromptVersion = s.version
	req.SystemPrompt = BuildSystemPrompt()
	req.UserPrompt = BuildUserPrompt(req)
	return s.provider.Summarize(ctx, req)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/agent/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/agent
git commit -m "feat(agent): add summarizer prompt pipeline"
```

### Task 3: 接入导入链路、缓存与降级

**Files:**
- Modify: `internal/cache/store.go`
- Modify: `internal/workbench/service.go`
- Modify: `internal/workbench/service_test.go`
- Modify: `app.go`

- [ ] **Step 1: 写 workbench 失败测试**

```go
func TestProcessDocumentStoresAgentFieldsAndProvider(t *testing.T) {
	service := newServiceWithStubSummarizer(t, stubSummarizeResponse())
	doc := writeMarkdownFixture(t, "# GMP\n\nGMP 是 Go 的调度模型。")

	got, err := service.ProcessDocument(doc, "gmp.md", "deepseek")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if got.Provider != "deepseek" {
		t.Fatalf("expected provider to be persisted")
	}
	if len(got.MemoryOutline) == 0 || len(got.SourceQuotes) == 0 {
		t.Fatalf("expected agent fields")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/workbench/...`
Expected: FAIL，提示 `ProcessDocument` 签名不匹配或字段不存在

- [ ] **Step 3: 实现 service 接入**

```go
type DocumentSummary struct {
	Provider      string   `json:"provider,omitempty"`
	Model         string   `json:"model,omitempty"`
	CardAnswer    string   `json:"cardAnswer,omitempty"`
	MemoryOutline []string `json:"memoryOutline,omitempty"`
	SourceQuotes  []string `json:"sourceQuotes,omitempty"`
	PromptVersion string   `json:"promptVersion,omitempty"`
}
```

处理策略：
- 先走 parser + segment
- 再调 summarizer
- provider 失败时降级到现有规则生成器
- 缓存键纳入 `provider + model + promptVersion`

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/workbench/... ./internal/cache/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add app.go internal/cache/store.go internal/workbench
git commit -m "feat(workbench): wire agent summarizer into import flow"
```

### Task 4: 前端 provider 选择器与卡片展示

**Files:**
- Modify: `frontend/index.html`
- Modify: `frontend/src/app.js`
- Modify: `frontend/src/style.css`
- Modify: `frontend/src/import-session.test.js`
- Create: `frontend/src/provider-options.test.js`

- [ ] **Step 1: 写前端失败测试**

```js
test("normalizeResult keeps provider and memory outline fields", () => {
  const result = normalizeResult({
    documents: [{
      provider: "deepseek",
      memoryOutline: ["定义", "机制"],
      sourceQuotes: ["依据 1"]
    }]
  });

  assert.equal(result.documents[0].provider, "deepseek");
  assert.deepEqual(result.documents[0].memoryOutline, ["定义", "机制"]);
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test frontend/src/*.test.js`
Expected: FAIL

- [ ] **Step 3: 实现 UI 接入**

要求：
- 新增 `Agent 接入方式` 下拉框
- 默认值来自后端配置或前端常量
- 导入时把 provider 一并传给 `ProcessDocument`
- 详情区展示：
  - 标准答案
  - 记忆提纲
  - 原文依据
  - provider / model 元信息

- [ ] **Step 4: 运行测试确认通过**

Run: `node --test frontend/src/*.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add frontend
git commit -m "feat(frontend): add provider selector and memory outline display"
```

### Task 5: 端到端验证与收尾

**Files:**
- Modify: `README.md` 或相关文档，仅在新增配置说明时修改

- [ ] **Step 1: 补 `.env` 配置说明**

```md
DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=
DEEPSEEK_MODEL=
MODELSCOPE_API_KEY=
MODELSCOPE_BASE_URL=
MODELSCOPE_MODEL=
```

- [ ] **Step 2: 跑完整验证**

Run: `go test ./...`
Expected: PASS

Run: `node --test frontend/src/*.test.js`
Expected: PASS

Run: `wails build`
Expected: build/bin/tenq-interview.exe 构建成功

- [ ] **Step 3: 最终提交**

```bash
git add .
git commit -m "feat: add agent-backed markdown summarization providers"
```
