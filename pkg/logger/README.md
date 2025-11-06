# Zap Logger 优化说明

## 优化内容

本次优化包含以下功能：

1. **线程安全优化** - 使用 `sync.Once` 和 `sync.RWMutex` 保证并发安全
2. **上下文字段提取优化** - 支持自定义提取器和更多默认字段
3. **性能优化** - 异步写入、可配置采样参数
4. **日志级别增强** - 添加 `Panic`/`Fatal` 辅助函数
5. **结构化日志增强** - 添加常用字段辅助函数
6. **优雅关闭机制** - 提供 `Sync()` 和 `Close()` 方法
7. **多环境配置优化** - 扩展配置模型，支持 YAML 配置
8. **日志采样优化** - 采样参数可配置化
9. **文件轮转增强** - 自动创建日志目录，精准匹配路径

## 配置示例

### 开发环境配置 (dev.yaml)

```yaml
logger:
  level: debug
  encoding: console
  development: true
  enableCaller: true
  enableStacktrace: false
  stacktraceLevel: error
  outputPaths:
    - stdout
  errorOutputPaths:
    - stderr
  serviceName: sensor-data-hub
  environment: dev
  async: false
  asyncBufferSize: 256
  sampling:
    enable: false
    tick: 1
    first: 100
    thereafter: 100
    level: info
  file:
    enable: false
    filename: ./logs/app.log
    maxSize: 128
    maxBackups: 7
    maxAge: 7
    compress: false
```

### 生产环境配置 (prod.yaml)

```yaml
logger:
  level: info
  encoding: json
  development: false
  enableCaller: true
  enableStacktrace: true
  stacktraceLevel: error
  outputPaths:
    - stdout
  errorOutputPaths:
    - stderr
  serviceName: sensor-data-hub
  environment: prod
  async: true
  asyncBufferSize: 512
  sampling:
    enable: true
    tick: 1
    first: 100
    thereafter: 100
    level: info
  file:
    enable: true
    filename: /var/log/sensor-data-hub/app.log  # 支持绝对路径
    maxSize: 256
    maxBackups: 10
    maxAge: 14
    compress: true
```

### 测试环境配置 (test.yaml)

```yaml
logger:
  level: debug
  encoding: console
  development: true
  enableCaller: true
  enableStacktrace: false
  stacktraceLevel: error
  outputPaths:
    - stdout
  errorOutputPaths:
    - stderr
  serviceName: sensor-data-hub
  environment: test
  async: false
  asyncBufferSize: 256
  sampling:
    enable: false
    tick: 1
    first: 100
    thereafter: 100
    level: info
  file:
    enable: true
    filename: ./logs/test.log  # 相对路径，会自动创建目录
    maxSize: 128
    maxBackups: 5
    maxAge: 3
    compress: false
```

## 使用示例

### 基本使用

```go
package main

import (
    "backend/pkg/logger"
    "context"
    "time"
)

func main() {
    // 初始化 logger
    cfg := logger.ProdConfig()
    if err := logger.InitGlobal(cfg); err != nil {
        panic(err)
    }
    defer logger.Close() // 优雅关闭

    // 基本日志记录
    logger.L().Info("应用启动成功")
    logger.S().Infof("服务名称: %s", "sensor-data-hub")

    // 使用结构化字段
    logger.L().Info("用户登录",
        logger.WithString("username", "admin"),
        logger.WithInt("user_id", 12345),
        logger.WithDuration(time.Second*2),
    )

    // 错误日志
    err := someFunction()
    if err != nil {
        logger.L().Error("操作失败",
            logger.WithError(err),
            logger.WithString("operation", "create_user"),
        )
    }

    // Panic 和 Fatal
    // logger.Panic("发生严重错误")  // 会触发 panic
    // logger.Fatal("致命错误")      // 会退出程序
}
```

### 使用 Context

```go
func handleRequest(ctx context.Context) {
    // WithContext 会自动提取 context 中的字段
    logger := logger.WithContext(ctx)
    logger.Info("处理请求")

    // 添加自定义字段
    logger.With(
        logger.WithString("endpoint", "/api/users"),
        logger.WithInt("status_code", 200),
    ).Info("请求处理完成")
}

// 在中间件中设置 context
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), "trace_id", generateTraceID())
        ctx = context.WithValue(ctx, "request_id", generateRequestID())
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 注册自定义 Context 提取器

```go
// 注册自定义字段提取器
logger.RegisterContextExtractor(func(ctx context.Context) []zap.Field {
    var fields []zap.Field
    if sessionID, ok := ctx.Value("session_id").(string); ok && sessionID != "" {
        fields = append(fields, zap.String("session_id", sessionID))
    }
    if ip, ok := ctx.Value("client_ip").(string); ok && ip != "" {
        fields = append(fields, zap.String("client_ip", ip))
    }
    return fields
})
```

### 从配置文件加载

```go
package main

import (
    "backend/config"
    "backend/pkg/logger"
    "github.com/spf13/viper"
)

func initLogger() error {
    // 从配置文件加载 logger 配置
    var cfg logger.LoggerConfig
    if err := viper.UnmarshalKey("logger", &cfg); err != nil {
        return err
    }
    
    return logger.InitGlobal(cfg)
}
```

## 配置字段说明

### LoggerConfig

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `level` | string | 日志级别：debug\|info\|warn\|error\|panic\|fatal | info |
| `encoding` | string | 编码格式：json\|console | json |
| `development` | bool | 是否为开发模式 | false |
| `enableCaller` | bool | 是否显示调用者信息 | false |
| `enableStacktrace` | bool | 是否启用堆栈跟踪 | false |
| `stacktraceLevel` | string | 堆栈跟踪级别 | error |
| `outputPaths` | []string | 输出路径列表 | ["stdout"] |
| `errorOutputPaths` | []string | 错误输出路径列表 | ["stderr"] |
| `serviceName` | string | 服务名称 | - |
| `environment` | string | 环境：dev\|test\|prod | - |
| `async` | bool | 是否异步写入 | false |
| `asyncBufferSize` | int | 异步缓冲区大小 | 256 |
| `file` | FileRotate | 文件轮转配置 | - |
| `sampling` | SamplingConfig | 采样配置 | - |

### FileRotate

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `enable` | bool | 是否启用文件日志 | false |
| `filename` | string | 日志文件路径（支持绝对路径和相对路径） | - |
| `maxSize` | int | 单个文件最大大小（MB） | 128 |
| `maxBackups` | int | 保留的备份文件数量 | 7 |
| `maxAge` | int | 保留日志文件的最大天数 | 7 |
| `compress` | bool | 是否压缩旧日志文件 | false |

### SamplingConfig

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `enable` | bool | 是否启用采样 | false |
| `tick` | int | 采样时间间隔（秒） | 1 |
| `first` | int | 第一个时间窗口内允许的日志条数 | 100 |
| `thereafter` | int | 后续时间窗口内允许的日志条数 | 100 |
| `level` | string | 采样级别（保留字段，暂未使用） | info |

## 注意事项

1. **路径处理**：日志文件路径支持绝对路径和相对路径，会自动创建目录
2. **异步写入**：启用异步写入时，程序退出前务必调用 `logger.Close()` 确保日志刷新
3. **线程安全**：所有全局函数都是线程安全的，可以在并发环境中使用
4. **采样配置**：采样功能可以限制日志输出频率，适合高并发场景
5. **文件轮转**：使用 lumberjack 实现日志轮转，支持按大小和时间轮转

## 性能优化建议

1. **生产环境**：建议启用异步写入和采样功能
2. **开发环境**：建议使用 console 编码，便于调试
3. **文件路径**：生产环境建议使用绝对路径，避免路径问题
4. **缓冲区大小**：根据日志量调整 `asyncBufferSize`，默认 256 适合大多数场景

