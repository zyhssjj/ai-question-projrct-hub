 3. `api文档.md`
```markdown
 API接口文档

 基础信息
- 基础URL：`http://localhost:8080/api`
- 请求格式：`application/json`
- 响应格式：`application/json`

 响应格式
```json
{
  "code": 200,      // 状态码：200成功，其他失败
  "msg": "success", // 提示信息
  "data": {}        // 响应数据
}
接口列表
1. 获取学习心得
URL: /note/get

方法: GET

描述: 读取学习心得.md文件内容

响应示例:

json
{
  "code": 200,
  "msg": "success",
  "data": " 学习心得\n\n这是我的学习心得..."
}
2. 获取题目列表
URL: /question/list

方法: GET

参数:

pageNum: 页码（默认1）

pageSize: 每页数量（默认10）

type: 题型（可选）

keyword: 搜索关键词（可选）

响应示例:

json
{
  "code": 200,
  "msg": "查询成功",
  "data": {
    "list": [
      {
        "id": 1,
        "type": "单选题",
        "content": "题目内容",
        "options": "[\"A. 选项1\", \"B. 选项2\"]",
        "answer": "A",
        "difficulty": "中等",
        "language": "Go",
        "createdAt": "2024-12-13T10:30:00Z"
      }
    ],
    "total": 100,
    "pageNum": 1,
    "pageSize": 10
  }
}
3. 添加单个题目
URL: /question/add

方法: POST

请求体:

json
{
  "type": "单选题",
  "content": "题目内容",
  "options": "[\"A. 选项1\", \"B. 选项2\"]",
  "answer": "A",
  "difficulty": "中等",
  "language": "Go"
}
4. 批量添加题目
URL: /question/add-batch

方法: POST

请求体: 题目数组

json
[
  {
    "type": "单选题",
    "content": "题目1",
    "difficulty": "中等",
    "language": "Go"
  },
  {
    "type": "多选题",
    "content": "题目2",
    "difficulty": "简单",
    "language": "JavaScript"
  }
]
5. 更新题目
URL: /question/update

方法: POST

请求体: 需要包含题目ID

json
{
  "id": 1,
  "type": "单选题",
  "content": "更新后的题目内容",
  "difficulty": "困难"
}
6. 删除题目
URL: /question/delete

方法: POST

请求体:

json
{
  "ids": [1, 2, 3]
}
7. AI生成题目
URL: /question/ai-generate

方法: POST

请求体:

json
{
  "type": "单选题",
  "count": 5,
  "difficulty": "中等",
  "language": "Go"
}
状态码说明
200: 请求成功

400: 请求参数错误

404: 资源不存在

500: 服务器内部错误

text

---

 第六步：运行说明文件

 1. `README.md`
```markdown
 题库生成管理系统

 项目简介
这是一个前后端分离的题库生成管理系统，支持AI自动出题、手工出题、题目管理等功能。

 功能特点
- ✅ AI智能出题（支持多种题型）
- ✅ 手工出题和题目编辑
- ✅ 题目搜索和筛选
- ✅ 批量操作（删除、选择）
- ✅ 学习心得展示（Markdown格式）
- ✅ 响应式布局（支持移动端）

 快速开始

 1. 环境要求
- Node.js 18+ (前端)
- Go 1.21+ (后端)
- SQLite3 (嵌入式，无需安装)

 2. 下载依赖
```bash
 前端依赖
cd client
npm install

 后端依赖
cd server
go mod tidy
 第七步：运行和测试

 启动步骤：

1. 启动后端：
```bash
cd week04/homework/server
go mod tidy
go run main.go

