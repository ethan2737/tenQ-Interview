# TenQ Interview

一个基于 **Wails** 的桌面复习工具，用来把 Markdown 面试资料整理成适合背诵和复述的题卡。

当前版本已经打通这条主链路：

- 导入本地 Markdown 文件或目录
- 预览文档并检查疑似乱码
- 调用 Agent / LLM 生成标准答案
- 展示记忆提纲与原文依据
- 本地缓存导入结果并在重启后恢复

## 项目定位

TenQ Interview 面向“已经整理了一批面试资料，但缺少高效复习方式”的个人用户。  
它不是知识库，也不是通用笔记工具，而是一个偏复习场景的桌面工作台。

当前产品重点是：

- 把长文资料压缩成 **一篇文档对应一张主卡**
- 主输出是 **150-220 字左右、适合直接复述的标准答案**
- 辅助输出是 **记忆提纲** 和 **原文依据**
- 结果严格受原文约束，避免自由发挥式总结

## 当前能力

### 已完成

- Markdown 文件 / 文件夹导入
- 导入前预览与疑似乱码确认
- 本地缓存与重启恢复
- 清空导入结果
- Markdown 代码块和图片渲染
- Agent 接入方式选择
- DeepSeek 在线 API 联调通过
- 规则摘要降级：当 Agent 不可用时，回退到本地规则生成链路

### 当前 Agent 能力

- 前端可选择 Agent 接入方式
- 当前优先建议使用 **DeepSeek**
- 后端已抽象出多 Provider 结构
- **ModelScope** 适配层已经接入，但当前仓库只完成了代码层支持，未作为默认实网方案启用

## 技术栈

- 桌面框架：`Wails v2`
- 后端：`Go 1.25`
- 前端：原生 `HTML / CSS / JavaScript`
- LLM 接入：OpenAI-compatible Chat Completions 协议

## 快速开始

### 1. 环境要求

- Go `1.25+`
- Wails CLI `v2.12+`
- Windows 环境下建议已安装 WebView2

### 2. 配置 `.env`

在项目根目录创建 `.env`，至少填写 DeepSeek 配置：

```env
LLM_PROVIDER_DEFAULT=deepseek

DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-chat

MODELSCOPE_API_KEY=
MODELSCOPE_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
MODELSCOPE_MODEL=qwen-plus
```

说明：

- 如果默认 Provider 是 `deepseek`，则必须提供 `DEEPSEEK_API_KEY`
- 如果默认 Provider 是 `modelscope`，则必须提供 `MODELSCOPE_API_KEY`
- 当前实际推荐使用 `deepseek`

### 3. 开发运行

```bash
wails dev
```

### 4. 构建桌面程序

```bash
wails build
```

构建产物默认位于：

```text
build/bin/tenq-interview.exe
```

## 测试

后端测试：

```bash
go test ./...
```

前端关键测试：

```bash
node --test frontend/src/import-session.test.js frontend/src/markdown-render.test.js frontend/src/agent-options.test.js frontend/src/layout-contract.test.js frontend/src/workbench-contract.test.js
```

## 目录结构

```text
frontend/              前端工作台与展示逻辑
internal/agent/        LLM Provider、Prompt、Summarizer
internal/workbench/    导入、预览、处理、恢复等工作台服务
internal/cache/        本地缓存与索引
docs/                  联调记录与计划文档
```

## 工作方式

当前导入处理采用“规则链路做约束，Agent 做受限提纯”的方式：

1. 读取并归一化 Markdown
2. 解析标题和正文
3. 提取候选片段
4. 调用 Agent 生成：
   - 标准答案
   - 记忆提纲
   - 原文依据
5. 写入本地缓存并恢复到工作台

如果 Agent 调用失败，系统会回退到本地规则摘要，避免整条导入链路失效。

## 当前状态

- Markdown MVP 主链路已打通
- DeepSeek 已完成小规模真实 API 联调
- Wails 桌面导入、恢复、清空、卡片展示已可用
- UI 已完成一轮针对阅读区优先的收敛

## 后续规划

后续方向见 [TODOS.md](./TODOS.md)，当前主要包括：

- PDF / docs 导入与多题切分
- 专题刷 / 随机刷复习模式
- 模拟面试播客音频
- 资料搜索能力

## 说明

- 本仓库当前以中文使用场景为主
- README 只描述已实现和已验证能力，不把规划项写成现状能力
