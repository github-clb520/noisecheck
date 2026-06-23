<p align="center">
  <img src="imgs/logo-core.svg" width="120" alt="NoiseCheck">
</p>

<h1 align="center">NoiseCheck · 零噪音 AI 代码审查</h1>

<p align="center">
  <em>让 AI 审查只说你该听的</em>
  <br>
  <strong>基于 <code>alibaba/open-code-review</code> 深度改造</strong>
</p>

<p align="center">
  <a href="#安装"><strong>🏗️ 安装</strong></a> ·
  <a href="#快速开始"><strong>⚡ 快速开始</strong></a> ·
  <a href="#配置"><strong>⚙️ 配置</strong></a> ·
  <a href="#输出格式"><strong>📄 输出格式</strong></a> ·
  <a href="#安全规则包"><strong>🔒 安全规则包</strong></a> ·
  <a href="#cicd-集成"><strong>🤖 CI/CD</strong></a> ·
  <a href="https://github.com/github-clb520/noisecheck/releases"><strong>📦 下载</strong></a>
</p>

---

## 目录

- [Why NoiseCheck?](#why-noisecheck)
- [安装](#安装)
- [快速开始](#快速开始)
- [CLI 命令](#cli-命令)
- [配置](#配置)
- [输出格式](#输出格式)
- [安全规则包](#安全规则包)
- [CI/CD 集成](#cicd-集成)
- [架构](#架构)
- [FAQ](#faq)
- [许可证](#许可证)

---

## Why NoiseCheck?

每次 PR 审查都看到 AI 在说"应该提取常量"、"加个注释"、"方法太长了"……这些噪音挤掉了真正需要关注的安全漏洞和架构问题。

**NoiseCheck = 三层降噪过滤：**

```
代码变更 → 预过滤器（锁文件/生成代码）
            ↓
        确定性规则过滤（仅传有实际价值的变更）
            ↓
         LLM 验证 → 分级报告
```

| 特性 | NoiseCheck | 其他 AI 审查工具 |
|------|-----------|----------------|
| 三层降噪过滤 | ✅ 内置 | ❌ 大部分没有 |
| 中文/英文输出 | ✅ 双语言 | ❌ 仅英文 |
| OWASP/Secrets/Infra 规则 | ✅ 50+ 内置 | ❌ 需单独配置 |
| HTML 报告（暗色主题） | ✅ 带级别过滤 | ❌ 大部分仅 CLI |
| Markdown 报告（CI 友好） | ✅ PR 直接嵌入 | ❌ |
| 严格程度可调 | ✅ 标准/严格/轻量 | ❌ |
| nc init 向导 | ✅ 2 分钟完成 | ❌ |

## 安装

### 一行安装（Linux / macOS）

```bash
curl -fsSL https://raw.githubusercontent.com/github-clb520/noisecheck/main/install.sh | sh
```

### Go

```bash
go install github.com/github-clb520/noisecheck/cmd/noisecheck@latest
```

### 手动下载

从 [Releases](https://github.com/github-clb520/noisecheck/releases) 下载对应平台的二进制：

| 平台 | 架构 | 文件 |
|------|------|------|
| Linux | amd64 | `noisecheck-linux-amd64` |
| Linux | arm64 | `noisecheck-linux-arm64` |
| macOS Intel | amd64 | `noisecheck-darwin-amd64` |
| macOS Apple Silicon | arm64 | `noisecheck-darwin-arm64` |
| Windows | amd64 | `noisecheck-windows-amd64.exe` |

## 快速开始

```bash
# 1️⃣ 初始化（首次使用）
nc init

# 2️⃣ 审查当前分支（与 main 对比）
nc review

# 3️⃣ 审查指定分支
nc review --from origin/main --to feature/foo

# 4️⃣ 审查特定 commit 范围
nc review --from HEAD~5 --to HEAD

# 5️⃣ 输出 HTML 报告
nc review --format html --report ./review.html

# 6️⃣ Markdown 输出（CI 环境）
nc review --format markdown > review.md
```

## CLI 命令

```
NoiseCheck — AI 代码审查 CLI

命令：
  init          交互式初始化向导
  review/r      运行代码审查（核心命令）
  config        配置管理
  
全局选项：
  --format text|json|markdown|html   输出格式（默认 text）
  --report path                       输出到文件（html 格式）
  --audience agent|developer          输出详略级别
  --from ref                          起始 git ref
  --to ref                            目标 git ref
  --help                             查看帮助
```

### review 命令详解

```bash
nc review                     # 审查当前分支（自动检测 main/master）
nc review --from origin/main  # 从 main 开始对比
nc review --format json       # JSON 输出（机器解析）
nc review --format html       # HTML 报告（浏览器查看）
nc review --format markdown   # Markdown 输出（CI 评论）
nc review --report report.html --format html   # 输出到文件
```

## 配置

### LLM 提供商

| 提供商 | 推荐模型 | 推荐 |
|--------|---------|------|
| **Anthropic Claude** | claude-sonnet-4-6 | ⭐ 推荐 |
| **OpenAI** | gpt-4o | ✅ 可用 |
| **兼容 API** | 自定义 | ✅ 可用 |

### 配置优先级（从高到低）

1. **命令行参数**：`--model "claude-sonnet-4-6"`
2. **环境变量**：`NC_LLM_URL` / `NC_LLM_TOKEN` / `NC_LLM_MODEL`
3. **`~/.noisecheck/config.json`**
4. **Claude Code 配置**（自动读取）
5. **shell rc 文件**（`.bashrc` / `.zshrc`）

### 配置文件

```json
// ~/.noisecheck/config.json
{
  "provider": "anthropic",
  "llm": {
    "url": "https://api.anthropic.com/v1/",
    "token": "sk-ant-xxxxxxxxxx",
    "model": "claude-sonnet-4-6"
  },
  "language": "Chinese",
  "api_type": "anthropic"
}
```

## 输出格式

### Text（默认）— 终端彩色输出

```
╔═══════════════════════════════════════╗
║         NoiseCheck 审查结果            ║
╚═══════════════════════════════════════╝
总数: 3

🔴 [严重] api/handler.go:45-52
SQL 注入风险 — 用户输入直接拼接 SQL 查询

代码:
  db.Query("SELECT * FROM users WHERE id = " + req.ID)
建议:
  使用参数化查询: db.Query("SELECT * FROM users WHERE id = ?", req.ID)
```

### HTML — 暗色主题交互式报告

生成可过滤的 HTML 报告，支持按严重级别和文件路径筛选。

### Markdown — CI 友好

适合直接嵌入 PR 评论或 CI 构建日志。

### JSON — 机器解析

适合集成到自定义流水线中。

## 安全规则包

内置 **50+ 安全规则**，分为三大类：

### OWASP Top 10 (10 项)

越权访问、加密缺陷、注入攻击、不安全设计、安全配置错误、已知漏洞组件、认证缺陷、数据完整性、日志监控、SSRF

### 密钥泄露检测 (10+ 项)

| 文件模式 | 检测内容 |
|---------|---------|
| `.env*` | 环境变量中的密钥 |
| `*.pem`, `*.key` | 私钥文件 |
| `secrets*`, `credentials*` | 凭据文件 |
| `.aws/**` | AWS 配置泄露 |
| `service-account*.json` | GCP 服务账号 |

### 基础设施安全 (20+ 项)

| 文件模式 | 检测内容 |
|---------|---------|
| `Dockerfile*` | root 运行、latest 标签、端口暴露过多 |
| `docker-compose*.yml` | 容器安全配置 |
| `*.tf` | Terraform（S3 公开、IAM 过度宽松、安全组 0.0.0.0/0） |
| `kubernetes*.yml` | K8s（privileged 模式、hostPath、root 容器） |

### 自定义规则

规则文件位于 `~/.noisecheck/rules/rule_docs/`，支持自定义 markdown 检查清单。关联的文件模式在 `system_rules.json` 中配置。

## CI/CD 集成

### GitHub Actions

```yaml
# .github/workflows/noisecheck-review.yml
name: NoiseCheck Review
on: [pull_request]
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - run: curl -fsSL https://raw.githubusercontent.com/github-clb520/noisecheck/main/install.sh | sh
      - run: nc review --format markdown --audience agent > review.md
        env:
          NC_LLM_URL: ${{ secrets.NC_LLM_URL }}
          NC_LLM_TOKEN: ${{ secrets.NC_LLM_TOKEN }}
      - uses: actions/github-script@v7
        with:
          script: |
            const report = require('fs').readFileSync('review.md','utf8');
            await github.rest.issues.createComment({
              ...context.repo, issue_number: context.issue.number, body: report
            });
```

### GitLab CI

```yaml
# .gitlab-ci.yml
noisecheck-review:
  image: alpine:latest
  only: [merge_requests]
  before_script:
    - apk add --no-cache curl bash
    - curl -fsSL https://raw.githubusercontent.com/github-clb520/noisecheck/main/install.sh | sh
  script:
    - nc review --format markdown --audience agent > review-report.md
  artifacts:
    paths: [review-report.md]
```

## 架构

```
cmd/noisecheck/          CLI 入口
├── main.go              命令分发
├── review_cmd.go        review 命令
├── init_cmd.go          init 向导
├── output.go            输出格式化（text/json/markdown）
└── flags.go             命令行参数

internal/
├── model/               评论模型（Severity/Category）
├── diff/                Git diff
├── llm/                 LLM 适配层（Anthropic/OpenAI）
├── config/              配置管理 + 规则系统
├── report/              HTML 报告生成器
├── agent/               AI Agent 逻辑
├── tool/                代码搜索工具
└── telemetry/           遥测

config/
├── system_rules.json    文件模式 → 规则映射
└── rule_docs/           Markdown 检查清单
```

## FAQ

### NoiseCheck 和原版 open-code-review 有什么区别？

NoiseCheck 是基于 alibaba/open-code-review 的深度改造版：
- 三层降噪过滤：预过滤 + 确定性规则 + LLM 验证
- 中文优先输出，同时支持英文
- 严重级别系统（严重/高危/中危/低危）带彩色标签
- 内置 50+ OWASP/Secrets/Infra 安全规则
- HTML 报告（暗色主题，可过滤）
- CI/CD 集成模板（GitHub Actions + GitLab CI）
- 交互式初始化向导

### 需要什么 LLM？费用如何？

推荐 Anthropic Claude（claude-sonnet-4-6），也支持 OpenAI。每次审查的 token 消耗取决于变更大小，通常每次 PR 审查约 5-50K token。

### 能在离线环境使用吗？

可以，但需要能访问 LLM API（自托管或代理）。规则文件本地缓存。

### 支持私有部署吗？

支持。配置 `NC_LLM_URL` 指向私有 API 端点即可，所有代码变更仅发送到你自己配置的 LLM。

## 许可证

[Apache License 2.0](LICENSE)

---

<p align="center">
  <a href="https://github.com/github-clb520/noisecheck">GitHub</a> ·
  <a href="https://github.com/github-clb520/noisecheck/issues">Issues</a> ·
  <a href="https://github.com/github-clb520/noisecheck/releases">Releases</a>
</p>
