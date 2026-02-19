# Quota-Activator

API 额度自动激活工具 - 通过智能调度在预设时间点触发微量请求，确保在高峰时段拥有全新 API 额度。

## 特性

- **智能调度**: 根据目标时间和刷新周期自动计算触发时间点
- **多平台支持**: 模块化设计，轻松扩展支持新的 AI 平台
- **安全缓冲**: 可配置的触发延迟，确保旧额度完全过期
- **重试机制**: 支持自动重试（指数退避）
- **优雅关闭**: 支持 Ctrl+C 信号处理
- **流式请求**: 开启流式模式，收到首个数据包后断开，节省资源

## 支持

## 安装

### 从源码编译

```bash
git clone <repository-url>
cd QuotaActivator
make build
```

### 使用 Go 安装

```bash
go install github.com/yourusername/quota-activator@latest
```

## 配置

创建 `config.yaml` 文件：

```yaml
# 调度器配置
scheduler:
  # 刷新周期（小时）
  interval_hours: 5

  # 目标时间点（24小时制）
  target_time: "14:00"

  # 安全缓冲时间（秒）- 确保旧额度完全过期
  safety_buffer_seconds: 60

# 平台配置
platform:
  # 平台类型: anthropic
  type: "anthropic"

  # API 端点
  base_url: "https://api.anthropic.com/v1/messages"

  # 平台特定选项
  options:
    # API 密钥
    api_key: "your-api-key-here"

    # 模型名称
    model: "claude-3-5-sonnet-20241022"

    # 请求超时时间（秒），默认 30
    timeout_seconds: 30

    # 最大重试次数，默认 0
    max_retries: 1
```

### 调度逻辑说明

假设配置为 `interval_hours: 5`, `target_time: "14:00"`:

- 系统会计算所有间隔 5 小时的触发点: `00:00, 05:00, 10:00, 15:00, 20:00, 01:00+1d, 06:00+1d, ...`
- 找出所有匹配 `14:00` 时分的触发点
- 选择当前时间之后的下一个触发点
- 加上 60 秒安全缓冲后执行

这样能确保在每天的 14:01 准时获得全新额度。

## 使用

### 使用 Makefile

```bash
# 编译
make build

# 运行
make run

# 开发模式（直接运行，不编译）
make run-dev
```

### 直接运行

```bash
# 编译后运行
./build/quota-activator

# 或使用 go run
go run .
```

### 日志输出

```
2025/02/19 16:00:00 Scheduler started for platform: anthropic
2025/02/19 16:00:00 Target time: 14:00, Interval: 5h, Safety buffer: 60s
2025/02/19 16:00:00 First trigger scheduled at: 2025-02-19 19:01:00
2025/02/19 16:00:00 Waiting 3h0m0s until next trigger...
2025/02/19 19:01:00 [anthropic] Triggering quota refresh...
2025/02/19 19:01:01 [SUCCESS] Trigger completed
2025/02/19 19:01:01 Next trigger: 2025-02-20 00:01:00
```

## 项目结构

```
QuotaActivator/
├── main.go                 # 入口点
├── config/                 # 配置模块
│   ├── config.go          # 配置结构和加载
│   └── validator.go       # 配置验证
├── scheduler/             # 调度模块
│   ├── scheduler.go       # 调度器主逻辑
│   └── calculator.go      # 触发时间计算
├── platform/              # 平台抽象模块
│   ├── platform.go        # Platform 接口定义
│   └── anthropic.go       # Anthropic 平台实现
├── config.yaml            # 配置文件
├── Makefile               # 构建脚本
└── README.md
```

## 添加新平台

1. 在 `platform/` 目录下创建新文件，如 `openai.go`:

```go
package platform

import (
    "context"
    "fmt"
    // 其他必要的导入
)

type OpenAIPlatform struct {
    // 平台特定字段
    apiKey  string
    model   string
    timeout time.Duration
}

func NewOpenAIPlatform(input *PlatformInput) (Platform, error) {
    // 解析配置并返回实例
    return &OpenAIPlatform{...}, nil
}

func (p *OpenAIPlatform) Name() string {
    return "openai"
}

func (p *OpenAIPlatform) Trigger(ctx context.Context) error {
    // 实现触发逻辑
    return nil
}

func (p *OpenAIPlatform) ValidateConfig() error {
    // 验证配置
    return nil
}
```

2. 在 `platform/platform.go` 的 `NewPlatform` 函数中添加 case:

```go
func NewPlatform(input *PlatformInput) (Platform, error) {
    switch input.Type {
    case "anthropic":
        return NewAnthropicPlatform(input)
    case "openai":
        return NewOpenAIPlatform(input)
    default:
        return nil, &UnsupportedPlatformError{Type: input.Type}
    }
}
```

3. 更新 `config/validator.go` 中的支持平台列表。

## 开发

```bash
# 格式化代码
make fmt

# 代码检查
make vet

# 运行测试
make test

# 测试覆盖率
make test-coverage

# 更新依赖
make deps-update
```

## 许可证

MIT License
