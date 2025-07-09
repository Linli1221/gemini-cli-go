# 🚀 Gemini CLI OpenAI Go

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

将 Google Gemini 模型转换为 OpenAI 兼容的 API 端点的 Go 语言实现。通过熟悉的 OpenAI API 模式访问 Google 最先进的 AI 模型，支持 OAuth2 认证。

## ✨ 功能特性

- 🔐 **OAuth2 认证** - 无需 API 密钥，使用您的 Google 账户
- 🎯 **OpenAI 兼容 API** - 直接替换 OpenAI 端点
- 📚 **OpenAI SDK 支持** - 与官方 OpenAI SDK 和库兼容
- 🖼️ **视觉支持** - 支持图像的多模态对话（base64 和 URL）
- 🌐 **第三方集成** - 兼容 Open WebUI、ChatGPT 客户端等
- ⚡ **高性能** - Go 语言实现，性能优异
- 🔄 **智能令牌缓存** - 使用内存缓存进行智能令牌管理
- 🆓 **免费层访问** - 通过 Code Assist API 利用 Google 的免费层
- 📡 **实时流式传输** - 服务器发送事件实现实时响应
- 🎭 **多模型支持** - 访问最新的 Gemini 模型，包括实验性模型

## 🤖 支持的模型

| 模型 ID | 上下文窗口 | 最大令牌 | 思维支持 | 描述 |
|---------|------------|----------|----------|------|
| `gemini-2.5-pro` | 1M | 65K | ✅ | 最新的 Gemini 2.5 Pro 模型，具有推理能力 |
| `gemini-2.5-flash` | 1M | 65K | ✅ | 快速的 Gemini 2.5 Flash 模型，具有推理能力 |

## 🛠️ 安装

### 使用 Go 编译

```bash
# 克隆仓库
git clone https://github.com/your-username/gemini-cli-go.git
cd gemini-cli-go

# 安装依赖
go mod download

# 编译
go build -o gemini-cli-go

# 运行
./gemini-cli-go
```

### 使用 Docker

```bash
# 构建镜像
docker build -t gemini-cli-go .

# 运行容器
docker run -p 8080:8080 --env-file .env gemini-cli-go
```

## ⚙️ 配置

### 环境变量

创建 `.env` 文件：

```env
# 必需：来自 Gemini CLI 认证的 OAuth2 凭据 JSON
GCP_SERVICE_ACCOUNT={"access_token":"ya29...","refresh_token":"1//...","scope":"...","token_type":"Bearer","id_token":"eyJ...","expiry_date":1750927763467}

# 可选：Google Cloud 项目 ID（如果未设置则自动发现）
# GEMINI_PROJECT_ID=your-project-id

# 可选：用于认证的 API 密钥（如果未设置，API 为公开访问）
# OPENAI_API_KEY=sk-your-secret-api-key-here

# 可选：启用假思维输出（设置为 "true" 启用）
# ENABLE_FAKE_THINKING=true

# 可选：启用真实 Gemini 思维输出（设置为 "true" 启用）
# ENABLE_REAL_THINKING=true

# 可选：将思维作为带有 <thinking> 标签的内容流式传输
# STREAM_THINKING_AS_CONTENT=true

# 服务器端口
PORT=8080
```

## 🚀 使用示例

### 使用 OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-your-secret-api-key-here"  # 如果启用了认证
)

response = client.chat.completions.create(
    model="gemini-2.5-flash",
    messages=[
        {"role": "user", "content": "解释一下机器学习"}
    ],
    stream=True
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

### 使用 cURL

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-secret-api-key-here" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [
      {"role": "user", "content": "你好！"}
    ]
  }'
```

## 📡 API 端点

### 基础 URL
```
http://localhost:8080
```

### 端点

- `GET /v1/models` - 列出可用模型
- `POST /v1/chat/completions` - 聊天完成
- `GET /v1/debug/cache` - 检查令牌缓存
- `POST /v1/token-test` - 测试认证
- `GET /health` - 健康检查

## 🔧 开发

### 运行测试

```bash
go test ./...
```

### 格式化代码

```bash
go fmt ./...
```

### 代码检查

```bash
go vet ./...
```

## 📄 许可证

本项目使用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 灵感来自官方 [Google Gemini CLI](https://github.com/google-gemini/gemini-cli)
- 使用 [Gin](https://gin-gonic.com/) Web 框架构建
- 基于原始 TypeScript 版本 [gemini-cli-openai](https://github.com/gewoonjaap/gemini-cli-openai)