# 雨云讲义测验 Agent

> 面向课程讲义、考研资料和考证资料的资料驱动型 AI 在线测验平台

**Go · React · Gin · GORM · PostgreSQL · pgvector · Supabase Storage · DeepSeek · Poppler · Tesseract**

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](week04/homework/server) [![React](https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=20232A)](week04/homework/client) [![Gin](https://img.shields.io/badge/Gin-API-00ADD8)](https://gin-gonic.com/) [![PostgreSQL](https://img.shields.io/badge/PostgreSQL-pgvector-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/) [![AI](https://img.shields.io/badge/AI-DeepSeek-111827)](https://www.deepseek.com/)

用户上传资料后，系统异步完成文本提取、扫描件 OCR 和知识切片；出卷 Agent 通过多轮对话收集题量、题型、难度、范围与时限，生成可直接在线作答的单选/多选试卷，并提供自动判分、讲义引用、错题归档与重练。

**面试关键词：** RAG · Agent 工作流 · 结构化输出校验 · 异步任务 Worker · HttpOnly 会话隔离 · 服务端判分 · 对象存储 · OCR

> 仓库中的主要代码位于 `week04/homework`。项目采用匿名浏览器会话隔离学习数据，定位为个人作品和技术验证，不宣称具备生产级账号认证能力。

## 核心流程

```text
上传 TXT / Markdown / DOCX / PDF
        |
        v
Supabase Storage 保存原文件
        |
        v
Worker 文本提取 / Poppler + Tesseract OCR
        |
        v
切片 + pgvector 相似度检索
        |
        v
Agent 收集出卷计划 -> DeepSeek 生成并校验题目
        |
        v
限时作答 -> 自动保存 -> 判分 -> 解析 / 引用 / 错题重练
```

## 已实现能力

- 支持 TXT、Markdown、DOCX、文字 PDF 与扫描 PDF，单文件最大 50 MB，PDF 最多 200 页。
- 原文件存入私有 Supabase Storage，PostgreSQL 只保存业务数据、解析结果和对象 Key。
- 资料解析与试卷生成统一使用持久化任务表，支持失败重试和服务重启后的超时任务恢复。
- 通过对话补全试卷名称、题量、单选/多选比例、难度、知识范围和时限。
- 基于 384 维本地特征哈希向量和 pgvector 进行资料片段召回，避免将整份长文直接发送给模型。
- AI 输出采用严格 JSON；服务端校验题型、选项、答案、解析、原文引用和重复题，并自动修复一次。
- 未交卷接口不返回正确答案、解析和引用；交卷后才通过结果接口公开。
- 服务端倒计时、自动保存、题目标记、超时自动交卷、完整集合匹配判分。
- 我的试卷保留多次作答记录；错题本按原始试卷聚合，并支持错题重练。
- 页面具备请求加载状态和本地缓存，减轻云数据库冷连接造成的空白等待。

## 技术栈

| 层次 | 技术 | 用途 |
| --- | --- | --- |
| Web | React 18、Vite 5、Ant Design 5 | 资料、Agent、试卷、答题、结果与错题本 |
| API | Go、Gin、GORM | 会话隔离、业务规则、任务 Worker 与模型接入 |
| 数据库 | Supabase PostgreSQL、pgvector | 结构化数据与片段向量检索 |
| 文件 | Supabase Storage | 私有保存 PDF、Word 等原始讲义 |
| AI | OpenAI 兼容 Chat Completions API | 出卷计划抽取、题目生成和自动修复 |
| 解析 | Poppler、Tesseract (`chi_sim+eng`) | PDF 文本提取与扫描件 OCR |

## 关键设计

### 为什么使用 Supabase Storage

PDF、DOCX 等非结构化大文件不适合直接写入 PostgreSQL，否则会扩大主库、增加备份和迁移成本。系统将原文件放入私有对象存储，数据库只保存对象 Key、状态和解析文本。Storage 请求只在 Go 服务端携带 service role key，浏览器无法接触该密钥。

### 数据隔离与答案防泄漏

首次访问由后端签发 30 天 HttpOnly Cookie。资料、任务、对话、试卷、答卷和错题查询都附加 `anonymous_session_id` 条件。作答接口使用专门 DTO，只返回题干、选项、已保存答案、标记状态和服务端截止时间。

### 异步任务与幂等领取

上传和生成接口快速创建 `generation_jobs`，后台 Worker 使用条件更新将 `queued` 任务切换为 `running`，避免重复领取。失败任务自动重试一次；启动时会把长时间未完成的 `running` 任务重新入队。

### RAG 的工程取舍

当前版本使用确定性的本地特征哈希生成 384 维向量，不依赖额外 Embedding API，便于演示与控制成本；PostgreSQL 环境通过 pgvector 余弦距离召回 12 个片段。它适合关键词相关的中文资料检索，但语义能力弱于专用 Embedding 模型。生产化可在不改变业务表和检索流程的前提下替换向量 Provider。

更完整的架构、数据模型和取舍见 [`week04/homework/技术文档.md`](week04/homework/技术文档.md)，接口清单见 [`week04/homework/api文档.md`](week04/homework/api文档.md)。

## 本地运行

### 1. 环境要求

- Go 1.25+
- Node.js 18+
- PostgreSQL 15+ 与 pgvector，或仅用于开发的 SQLite 回退
- Supabase 私有 Storage Bucket（默认 `materials`）
- Poppler 与 Tesseract（扫描 PDF 需要 `chi_sim`、`eng`）

### 2. 配置

```powershell
cd week04/homework
Copy-Item server/.env.example server/.env
Copy-Item client/.env.example client/.env
```

按注释填写 `server/.env`。不要提交 `.env`；前端不得保存数据库、Storage 或 AI 密钥。

DeepSeek 的实际模型标识以账号控制台当前支持列表为准，示例中的模型名只是部署配置占位值。

### 3. 启动

Windows 可直接执行：

```powershell
.\start.ps1
```

或分别启动：

```powershell
cd server
go run .

cd ../client
npm install
npm run dev
```

访问 `http://localhost:5173`。首次建表可设置 `AUTO_MIGRATE=true`，完成后建议改为 `false`。

## 测试与构建

```powershell
cd week04/homework/server
go test ./... -count=1

cd ../client
npm run build
```

当前自动化测试覆盖答案集合判定、题目结构校验、对象 Key 安全、向量生成和运行时配置写入。真实 Supabase、OCR 与模型调用属于外部依赖集成测试，需要使用部署环境的测试账号执行。

## 项目边界

- 当前是匿名会话学习空间，不是完整的注册登录系统；个人资料和密码功能不等同于生产级身份认证。
- SQLite 仅用于无 PostgreSQL 配置时的本地开发，pgvector 检索只在 PostgreSQL 下启用。
- Worker 目前与 API 同进程，适合单实例演示；生产多实例应拆分独立 Worker，并使用数据库锁或消息队列。
- 模型生成仍有不确定性，因此后端必须保留结构化校验、引用校验和失败重试。

## 面试讲解重点

1. 为什么把原始文件放对象存储，而把状态、切片和作答数据放 PostgreSQL。
2. 如何通过 HttpOnly 会话、查询条件和响应 DTO 防止越权与答案提前泄漏。
3. 为什么上传与生成要异步化，以及任务重试、恢复和幂等领取如何实现。
4. AI 输出为什么不能直接入库，服务端如何做格式、答案、引用和重复度校验。
5. 当前本地向量方案的成本优势与语义局限，以及如何演进为可插拔 Embedding Provider。

## License

本项目用于个人学习、作品展示与面试交流。
