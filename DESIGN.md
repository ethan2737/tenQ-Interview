# Design System — TenQ Interview

## Product Context
- **What this is:** 一个面向个人使用的桌面复习工作台，用来把长篇面试资料压缩成适合真实面试口头回答的题卡。
- **Who it's for:** 当前主要服务正在准备 Web3 后端开发岗位面试的用户，尤其是已经整理了大量资料、但需要更高复习效率的人。
- **Space/industry:** 学习工具 / 面试准备 / 知识压缩。
- **Project type:** 桌面应用，专注型工作台，而不是营销站点或通用后台。

## Aesthetic Direction
- **Direction:** Quiet Editorial Utility
- **Decoration level:** intentional
- **Mood:** 像一张安静、可信、可以长时间停留的晨间复习台。它应该有书房感和阅读感，但仍然保留清晰的工具理性，不做营销化、游戏化、SaaS 化表演。
- **Reference sites:** 研究结论来自学习工具、flashcard、知识管理产品的通用模式观察。方向上刻意避开泛 AI 学习工具和通用 dashboard 的模板气质。

## Typography
- **Display/Hero:** `Noto Serif SC`
  - 用于题目、空状态主标题、关键区块标题。
  - 理由：给题目和材料一点分量与记忆锚点，不让界面只有工具味。
- **Body:** `Source Han Sans SC`
  - 用于正文、按钮、树节点、状态文案。
  - 理由：中文长文阅读稳定、密度高时不乱，是系统里的主工作字体。
- **UI/Labels:** `Source Han Sans SC`
  - 与正文统一，避免视觉语言割裂。
- **Data/Tables:** `JetBrains Mono`
  - 用于状态标签、快捷键、路径、版本号、技术性小信息。
  - 理由：让工具感落在元信息层，而不是侵入主要阅读面。
- **Code:** `JetBrains Mono`
- **Loading:** 优先使用 Google Fonts 或项目内静态托管；若运行环境受限，中文回退到 `PingFang SC` / `Microsoft YaHei`。

### Type Scale
- Display XL: 56px
- Display L: 40px
- Section Title: 28px
- Question Title: 32px
- Body L: 18px
- Body M: 16px
- Body S: 14px
- Meta / Mono: 12px

## Color
- **Approach:** restrained
- **Primary:** `#285E61`
  - 用途：主按钮、当前选中、可信提示、主交互锚点。
  - 含义：稳定、克制、可信。
- **Secondary:** `#C08A3E`
  - 用途：轻微强调、进度点缀、少量二级高亮。
  - 含义：暖度、注意力提气，但绝不成为主色。
- **Neutrals:**
  - Canvas: `#F5F1E8`
  - Surface: `#FBF8F2`
  - Surface Alt: `#F1EBDF`
  - Border: `#D9D2C3`
  - Ink: `#1F2933`
  - Muted Text: `#667085`
- **Semantic:**
  - Success: `#3E7C59`
  - Warning: `#B7791F`
  - Error: `#B2544F`
  - Info: `#285E61`
- **Dark mode:** 使用低饱和深灰绿底，不直接反相。暗色模式下主色与强调色降低刺眼度，保证阅读稳定而不是霓虹感。

## Spacing
- **Base unit:** 8px
- **Density:** comfortable
- **Scale:**
  - 2xs: 4px
  - xs: 8px
  - sm: 16px
  - md: 24px
  - lg: 32px
  - xl: 48px
  - 2xl: 64px

## Layout
- **Approach:** grid-disciplined
- **Primary structure:** 左树右读，右侧阅读面优先。
- **Grid:** 桌面双栏，窄窗口下左侧树折叠为侧栏或抽屉，不做简单上下堆叠。
- **Max content width:** 右侧正文阅读宽度建议控制在 68-76 字符区间，避免长行阅读疲劳。
- **Border radius:**
  - sm: 6px
  - md: 10px
  - lg: 14px
  - full: 9999px
- **Container philosophy:** 右侧题卡区是阅读面，不是厚重卡片。层级主要依靠排版、留白、字号、权重和细边界建立。

## Motion
- **Approach:** minimal-functional
- **Easing:**
  - enter: ease-out
  - exit: ease-in
  - move: ease-in-out
- **Duration:**
  - micro: 50-100ms
  - short: 150-220ms
  - medium: 220-320ms
  - long: 320-500ms
- **Rules:**
  - 不做漂浮、弹跳、装饰性 motion。
  - 动效只服务于层级切换、折叠展开、焦点转移和状态反馈。

## Component Principles
- 第一屏默认是复习台，不是导入台。
- 左侧资料树只显示文件名，避免树本身成为噪音。
- 右侧详情顺序固定为：
  1. 题目
  2. 标准答案
  3. 可信提示
  4. 来源依据折叠区
  5. 原文片段折叠区
  6. 次级操作按钮
- 空状态必须像“晨间开始台”，不是系统空白。
- 批量处理中必须显示进度概览，并允许已完成项立刻可用。
- 单文件失败必须显示文件级失败面板，不允许静默缺失。

## Responsive & Accessibility
- 窄窗口下右侧阅读面优先保宽度，左树折叠为可展开侧栏。
- 第一版支持最小键盘操作集：
  - 上下切题
  - 展开/收起依据
  - 聚焦导入入口
  - Esc 关闭侧栏或弹层
- 交互元素应有清晰焦点态。
- 点击目标遵守 44px 最小触达标准。
- 颜色对比不得依赖浅灰弱对比伪精致。

## Anti-Patterns
- 不要使用紫色渐变、蓝紫科技风默认配色。
- 不要把右侧阅读面做成 SaaS 卡片墙。
- 不要让左树和右侧阅读区像两套不同产品。
- 不要在空状态里只写“暂无数据”。
- 不要用重阴影、大圆角、装饰图形来假装设计完成。

## Decisions Log
| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-17 | 初始设计系统创建 | 基于 office-hours、plan-eng-review、plan-design-review 的产品与交互约束综合生成 |
