package agent

import (
	"fmt"
	"strings"
)

func BuildSystemPrompt() string {
	return strings.TrimSpace(`
你是一个面试答案整理助手。
你的职责是基于用户提供的原文，整理出适合背诵和复述的标准答案。
你不是知识专家，不允许引入原文之外的新知识。
你可以润色语气，让表达更像真实面试回答，但不能新增事实、概念、例子或结论。
如果原文不足以支撑完整回答，必须保守表达，并在 notes 中说明信息不足。
你必须输出结构化 JSON。`)
}

func BuildUserPrompt(req SummarizeRequest) string {
	candidates := "无"
	if len(req.CandidateText) > 0 {
		candidates = strings.Join(req.CandidateText, "\n---\n")
	}

	return fmt.Sprintf(strings.TrimSpace(`
任务：
1. 基于原文生成 150 到 220 字的标准答案，适合直接背诵。
2. 生成 3 到 5 条记忆提纲。
3. 提取 2 到 4 条原文依据。
4. 输出 JSON，字段必须为 standard_answer, memory_outline, source_quotes, notes。

边界：
- 不允许引入原文没有的信息。
- 不允许使用外部知识补足定义或结论。
- 每个核心判断都必须能在原文中找到依据。

标题：
%s

候选片段：
%s

原文全文：
%s
`), req.Title, candidates, req.Body)
}
