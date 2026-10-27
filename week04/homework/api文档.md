# 雨云讲义测验 Agent API

## 1. 基础约定

- 本地地址：`http://localhost:8080/api`
- 响应类型：`application/json`
- 文件上传：`multipart/form-data`
- 会话：HttpOnly Cookie `study_session`
- 除 `/session/bootstrap` 外，所有接口都要求有效会话 Cookie。

统一成功响应：

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

统一失败响应：

```json
{
  "code": 400,
  "msg": "错误原因"
}
```

## 2. 会话

### `POST /session/bootstrap`

创建或续用匿名会话，并设置 30 天 HttpOnly Cookie。

## 3. 资料与任务

### `POST /materials`

上传字段名为 `file` 的 TXT、Markdown、DOCX 或 PDF，最大 50 MB。接口返回后，解析任务在后台执行。

```json
{
  "code": 200,
  "data": {
    "material": { "id": 1, "name": "讲义.pdf", "status": "processing" },
    "jobId": 1
  }
}
```

### `GET /materials`

返回当前会话的资料列表。状态为 `processing`、`ready` 或 `failed`。

### `GET /materials/:id`

返回资料元数据、解析状态、错误原因和章节摘要，不返回完整正文。

### `GET /jobs/:id`

查询资料解析或试卷生成任务，重点字段包括 `kind`、`status`、`progress` 和 `retryCount`。

## 4. Agent 出卷

### `POST /agent/conversations`

```json
{ "materialIds": [1, 2] }
```

只能选择当前会话内状态为 `ready` 的资料。

### `GET /agent/conversations/:id`

返回会话、历史消息和当前计划。

### `POST /agent/conversations/:id/messages`

```json
{
  "content": "生成一份 Go 并发测试，10 题，6 道单选、4 道多选，中等难度，30 分钟"
}
```

计划完整时返回 `status: "ready"`；否则返回缺失字段和下一轮追问。

### `POST /agent/conversations/:id/confirm`

确认 `ready` 计划并创建异步生成任务，返回 `jobId` 和 `examId`。

## 5. 试卷

### `GET /exams`

返回当前会话全部试卷、作答次数和最近一次作答。

### `GET /exams/:id`

返回试卷元数据和历次作答记录，不返回题目答案。

### `POST /exams/:id/start`

创建一次答卷；如果存在未过期的进行中答卷，则返回该答卷继续作答。

## 6. 作答

### `GET /attempts/:id/questions`

只允许读取进行中的当前会话答卷。返回截止时间、题干、选项、已保存答案和标记状态，**不会返回正确答案、解析或引用**。

### `PUT /attempts/:id/answers/:questionId`

```json
{ "answer": ["A", "C"] }
```

后端验证选项 ID、重复值和题型允许的答案数量，再通过唯一键 upsert。

### `PUT /attempts/:id/marks/:questionId`

```json
{ "marked": true }
```

### `POST /attempts/:id/submit`

幂等交卷。若服务端时间已经超过 `expires_at`，提交原因记为 `timeout`。

### `GET /attempts/:id/result`

仅已结束答卷可访问，返回分数以及每题的用户答案、正确答案、正误、解析和讲义引用。

### `POST /attempts/:id/wrong-questions/retry`

根据该次答卷的错题创建一份快照试卷，返回新试卷元数据。

## 7. 错题本

### `GET /wrong-questions`

返回扁平错题记录，包含正确答案、解析和引用。

### `GET /wrong-books`

按原始试卷聚合错题与历次作答。错题重练产生的派生试卷通过 `source_exam_id` 归入原试卷。

## 8. 个人空间

### `GET /profile`

返回匿名学习空间的姓名、派生账号、密码状态和创建时间。

### `PUT /profile`

```json
{ "name": "张三" }
```

### `PUT /profile/password`

```json
{
  "currentPassword": "旧密码（首次设置可空）",
  "newPassword": "至少8位的新密码"
}
```

此能力属于匿名学习空间设置，不是完整登录认证系统。

## 9. 智能体配置

以下接口仅允许从后端服务器 loopback 地址调用，线上远程访问返回 `403`。

### `GET /settings/ai`

返回 Base URL、模型名、是否已配置及掩码后的 Key。

### `PUT /settings/ai`

```json
{
  "apiKey": "留空则保持原值",
  "baseUrl": "https://api.deepseek.com",
  "model": "账号实际可用的模型标识"
}
```

### `POST /settings/ai/test`

向当前模型发送最小 JSON 请求以验证连接。

## 10. 常用状态码

| 状态码 | 含义 |
| --- | --- |
| 200 | 请求成功 |
| 400 | 参数或当前业务状态不合法 |
| 401 | 会话不存在或已过期 |
| 403 | 不允许从当前位置操作 |
| 404 | 当前会话中不存在该资源 |
| 409 | 答卷已结束、计划状态冲突等 |
| 500 | 服务内部错误 |
| 502 | Storage 或 AI 等上游服务失败 |
| 503 | 数据库等基础设施暂时不可用 |
