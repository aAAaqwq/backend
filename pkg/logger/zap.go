package logger

import (
	"backend/config"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ContextFieldExtractor 上下文字段提取器函数类型
type ContextFieldExtractor func(ctx context.Context) []zap.Field

// ---- Global ----
var (
	globalLogger      *zap.Logger
	globalSugar       *zap.SugaredLogger
	globalLoggerOnce  sync.Once
	globalLoggerMutex sync.RWMutex
	contextExtractors []ContextFieldExtractor
	extractorsMutex   sync.RWMutex
	bufferedSyncers   []*bufferedWriteSyncer // 跟踪所有 bufferedWriteSyncer 实例
	syncersMutex      sync.Mutex             // 保护 bufferedSyncers
)

// RegisterContextExtractor 注册上下文字段提取器
func RegisterContextExtractor(extractor ContextFieldExtractor) {
	extractorsMutex.Lock()
	defer extractorsMutex.Unlock()
	contextExtractors = append(contextExtractors, extractor)
}

// L 获取全局 Logger（线程安全）
func L() *zap.Logger {
	globalLoggerMutex.RLock()
	if globalLogger != nil {
		defer globalLoggerMutex.RUnlock()
		return globalLogger
	}
	globalLoggerMutex.RUnlock()

	globalLoggerOnce.Do(func() {
		_ = InitGlobal(DevConfig()) // fallback
	})

	globalLoggerMutex.RLock()
	defer globalLoggerMutex.RUnlock()
	return globalLogger
}

// S 获取全局 SugaredLogger（线程安全）
func S() *zap.SugaredLogger {
	globalLoggerMutex.RLock()
	if globalSugar != nil {
		defer globalLoggerMutex.RUnlock()
		return globalSugar
	}
	globalLoggerMutex.RUnlock()

	globalLoggerOnce.Do(func() {
		_ = InitGlobal(DevConfig())
	})

	globalLoggerMutex.RLock()
	defer globalLoggerMutex.RUnlock()
	return globalSugar
}

// InitGlobal builds a logger and sets it as global (thread-safe).
func InitGlobal(cfg config.LoggerConfig, opts ...zap.Option) error {
	logger, err := New(cfg, opts...)
	if err != nil {
		return err
	}

	globalLoggerMutex.Lock()
	defer globalLoggerMutex.Unlock()

	// 关闭旧的 logger（不等待 Sync 完成，避免阻塞）
	if globalLogger != nil {
		// 异步同步，避免阻塞初始化
		go func() {
			_ = globalLogger.Sync()
		}()
	}

	globalLogger = logger
	globalSugar = logger.Sugar()
	zap.ReplaceGlobals(logger)
	return nil
}

// Sync 同步所有日志缓冲区（优雅关闭，不阻塞）
func Sync() error {
	globalLoggerMutex.RLock()
	defer globalLoggerMutex.RUnlock()
	if globalLogger != nil {
		// 使用非阻塞方式同步，避免阻塞
		_ = globalLogger.Sync()
	}
	return nil
}

// Close 关闭 logger（优雅关闭，停止所有异步写入 goroutine）
func Close() error {
	// 先同步所有日志
	Sync()

	// 停止所有 bufferedWriteSyncer 的 goroutine
	syncersMutex.Lock()
	syncers := make([]*bufferedWriteSyncer, len(bufferedSyncers))
	copy(syncers, bufferedSyncers)
	bufferedSyncers = nil // 清空列表
	syncersMutex.Unlock()

	// 停止所有 goroutine
	for _, bws := range syncers {
		bws.Stop()
	}

	return nil
}

// New builds a zap logger from Config (no side effects).
func New(cfg config.LoggerConfig, opts ...zap.Option) (*zap.Logger, error) {
	level := parseLevel(cfg.Level)

	encCfg := encoderConfig(cfg.Development)
	enc := func() zapcore.Encoder {
		if strings.ToLower(cfg.Encoding) == "console" {
			return zapcore.NewConsoleEncoder(encCfg)
		}
		return zapcore.NewJSONEncoder(encCfg)
	}()

	var cores []zapcore.Core

	// OutputPaths: 处理正常日志输出
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}
	for _, p := range cfg.OutputPaths {
		// 只对文件路径使用异步，stdout/stderr 始终同步
		isFile := !isStdOutput(p)
		useAsync := cfg.Async && isFile
		ws, err := writerFor(p, useAsync, cfg.AsyncBufferSize, cfg.AsyncFlushInterval)
		if err != nil {
			return nil, err
		}
		cores = append(cores, zapcore.NewCore(enc, ws, level))
	}

	// ErrorOutputPaths: 处理错误日志输出（zap 内部使用）
	// 注意：zap 的 ErrorOutputPaths 主要用于内部错误输出，不是日志级别过滤
	// 但为了完整性，我们也处理它
	if len(cfg.ErrorOutputPaths) == 0 {
		cfg.ErrorOutputPaths = []string{"stderr"}
	}
	var errorWriters []zapcore.WriteSyncer
	for _, p := range cfg.ErrorOutputPaths {
		// 只对文件路径使用异步，stdout/stderr 始终同步
		isFile := !isStdOutput(p)
		useAsync := cfg.Async && isFile
		ws, err := writerFor(p, useAsync, cfg.AsyncBufferSize, cfg.AsyncFlushInterval)
		if err != nil {
			return nil, err
		}
		errorWriters = append(errorWriters, ws)
	}

	// optional file rotate
	if cfg.File.Enable && cfg.File.Filename != "" {
		// 确保日志目录存在
		if err := ensureLogDir(cfg.File.Filename); err != nil {
			return nil, err
		}

		fw := &lumberjack.Logger{
			Filename:   cfg.File.Filename,
			MaxSize:    defaultInt(cfg.File.MaxSize, 128),
			MaxBackups: defaultInt(cfg.File.MaxBackups, 7),
			MaxAge:     defaultInt(cfg.File.MaxAge, 7),
			Compress:   cfg.File.Compress,
		}

		var ws zapcore.WriteSyncer
		fwSyncer := zapcore.AddSync(fw)
		if cfg.Async {
			// 文件写入使用异步
			ws = newBufferedWriteSyncer(fwSyncer, cfg.AsyncBufferSize, cfg.AsyncFlushInterval)
		} else {
			ws = fwSyncer
		}
		cores = append(cores, zapcore.NewCore(enc, ws, level))
	}

	core := zapcore.NewTee(cores...)

	// options
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller())
	}
	if cfg.EnableStacktrace {
		stackLevel := zapcore.ErrorLevel
		if cfg.StacktraceLevel != "" {
			stackLevel = parseLevel(cfg.StacktraceLevel)
		}
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	// 采样配置
	if cfg.Sampling.Enable {
		tick := time.Duration(defaultInt(cfg.Sampling.Tick, 1)) * time.Second
		first := defaultInt(cfg.Sampling.First, 100)
		thereafter := defaultInt(cfg.Sampling.Thereafter, 100)

		opts = append(opts, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(c, tick, first, thereafter)
		}))
	}

	fields := []zap.Field{
		zap.String("svc", cfg.ServiceName),
		zap.String("env", cfg.Environment),
	}
	opts = append(opts, zap.Fields(fields...))

	// 设置 ErrorOutput（zap 内部使用）
	if len(errorWriters) > 0 {
		errorOutput := zapcore.NewMultiWriteSyncer(errorWriters...)
		opts = append(opts, zap.ErrorOutput(errorOutput))
	}

	logger := zap.New(core, opts...)
	return logger, nil
}

// WithContext adds common IDs from context to logger (e.g., trace_id, user_id).
// Supports both default extra
// ctors and custom extractors registered via RegisterContextExtractor.
func WithContext(ctx context.Context) *zap.Logger {
	l := L()
	if ctx == nil {
		return l
	}

	var fields []zap.Field

	// 默认字段提取器
	type ctxKey string
	if traceID, ok := ctx.Value(ctxKey("trace_id")).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if userID, ok := ctx.Value(ctxKey("user_id")).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	if reqID, ok := ctx.Value(ctxKey("request_id")).(string); ok && reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}
	if spanID, ok := ctx.Value(ctxKey("span_id")).(string); ok && spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}
	if correlationID, ok := ctx.Value(ctxKey("correlation_id")).(string); ok && correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	// 自定义提取器
	extractorsMutex.RLock()
	extractors := contextExtractors
	extractorsMutex.RUnlock()

	for _, extractor := range extractors {
		if extracted := extractor(ctx); len(extracted) > 0 {
			fields = append(fields, extracted...)
		}
	}

	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

// WithContextFields 从 context 提取字段并添加自定义字段
func WithContextFields(ctx context.Context, fields ...zap.Field) *zap.Logger {
	logger := WithContext(ctx)
	if len(fields) > 0 {
		return logger.With(fields...)
	}
	return logger
}

// ---- 日志级别增强 ----

// Panic 记录 panic 级别日志并触发 panic
func Panic(msg string, fields ...zap.Field) {
	L().Panic(msg, fields...)
}

// Panicf 记录 panic 级别日志并触发 panic（格式化）
func Panicf(template string, args ...interface{}) {
	S().Panicf(template, args...)
}

// Fatal 记录 fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// Fatalf 记录 fatal 级别日志并退出程序（格式化）
func Fatalf(template string, args ...interface{}) {
	S().Fatalf(template, args...)
}

// ---- 结构化日志增强 ----

// WithError 添加 error 字段
func WithError(err error) zap.Field {
	return zap.Error(err)
}

// WithDuration 添加 duration 字段
func WithDuration(d time.Duration) zap.Field {
	return zap.Duration("duration", d)
}

// WithString 添加 string 字段
func WithString(key, val string) zap.Field {
	return zap.String(key, val)
}

// WithInt 添加 int 字段
func WithInt(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// WithInt64 添加 int64 字段
func WithInt64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// WithFloat64 添加 float64 字段
func WithFloat64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

// WithBool 添加 bool 字段
func WithBool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// WithTime 添加 time 字段
func WithTime(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

// WithAny 添加任意类型字段
func WithAny(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

// ---- helpers ----

func DevConfig() config.LoggerConfig {
	return config.LoggerConfig{
		Level:              "debug",
		Encoding:           "console",
		Development:        true,
		EnableCaller:       true,
		EnableStacktrace:   false,
		StacktraceLevel:    "error",
		OutputPaths:        []string{"stdout"},
		ErrorOutputPaths:   []string{"stderr"},
		ServiceName:        "example-svc",
		Environment:        "dev",
		Async:              false,
		AsyncBufferSize:    256,
		AsyncFlushInterval: 200, // 200ms
		Sampling: config.LoggerSamplingConfig{
			Enable:     false,
			Tick:       1,
			First:      100,
			Thereafter: 100,
			Level:      "info",
		},
	}
}

func ProdConfig() config.LoggerConfig {
	return config.LoggerConfig{
		Level:              "info",
		Encoding:           "json",
		Development:        false,
		EnableCaller:       true,
		EnableStacktrace:   true,
		StacktraceLevel:    "error",
		OutputPaths:        []string{"stdout"},
		ErrorOutputPaths:   []string{"stderr"},
		ServiceName:        "example-svc",
		Environment:        "prod",
		Async:              true,
		AsyncBufferSize:    256,
		AsyncFlushInterval: 200, // 200ms
		File: config.LoggerFileRotate{
			Enable:     true,
			Filename:   "./logs/app.log",
			MaxSize:    256,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		},
		Sampling: config.LoggerSamplingConfig{
			Enable:     true,
			Tick:       1,
			First:      100,
			Thereafter: 100,
			Level:      "info",
		},
	}
}

func TestConfig() config.LoggerConfig {
	return config.LoggerConfig{
		Level:              "debug",
		Encoding:           "console",
		Development:        true,
		EnableCaller:       true,
		EnableStacktrace:   false,
		StacktraceLevel:    "error",
		OutputPaths:        []string{"stdout"},
		ErrorOutputPaths:   []string{"stderr"},
		ServiceName:        "example-svc",
		Environment:        "test",
		Async:              false,
		AsyncBufferSize:    256,
		AsyncFlushInterval: 200, // 200ms
		Sampling: config.LoggerSamplingConfig{
			Enable:     false,
			Tick:       1,
			First:      100,
			Thereafter: 100,
			Level:      "info",
		},
		File: config.LoggerFileRotate{
			Enable:     true,
			Filename:   "./logs/app.log",
			MaxSize:    256,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		},
	}
}

func parseLevel(lv string) zapcore.Level {
	switch strings.ToLower(lv) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func encoderConfig(dev bool) zapcore.EncoderConfig {
	c := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     iso8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
	}
	if dev {
		c.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return c
}

func iso8601TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339Nano))
}

// isStdOutput 判断是否为标准输出（stdout/stderr）
func isStdOutput(path string) bool {
	lower := strings.ToLower(path)
	return lower == "stdout" || lower == "stderr"
}

func writerFor(path string, async bool, bufferSize int, flushInterval int) (zapcore.WriteSyncer, error) {
	switch strings.ToLower(path) {
	case "stdout":
		// stdout 始终同步，不使用异步
		return zapcore.AddSync(os.Stdout), nil
	case "stderr":
		// stderr 始终同步，不使用异步
		return zapcore.AddSync(os.Stderr), nil
	default:
		// 文件路径：根据 async 参数决定是否异步
		// 确保日志目录存在
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %v", path, err)
		}

		if err := ensureLogDir(absPath); err != nil {
			return nil, fmt.Errorf("failed to ensure log dir for %s: %v", absPath, err)
		}

		fw := &lumberjack.Logger{
			Filename:   absPath, // 使用绝对路径
			MaxSize:    128,
			MaxBackups: 7,
			MaxAge:     7,
			Compress:   true,
		}
		fwSyncer := zapcore.AddSync(fw)
		if async {
			// 文件写入使用异步
			return newBufferedWriteSyncer(fwSyncer, bufferSize, flushInterval), nil
		}
		return fwSyncer, nil
	}
}

// ensureLogDir 确保日志文件所在目录存在
func ensureLogDir(filename string) error {
	dir := filepath.Dir(filename)
	if dir == "" || dir == "." {
		return nil
	}
	// 检查是否为绝对路径或相对路径
	if !filepath.IsAbs(dir) {
		// 相对路径，转换为绝对路径
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		dir = absDir
	}
	return os.MkdirAll(dir, 0755)
}

// bufferedWriteSyncer 缓冲写入同步器（用于异步写入）
type bufferedWriteSyncer struct {
	ws            zapcore.WriteSyncer
	buffer        *buffer.Buffer
	mu            sync.Mutex
	stopCh        chan struct{}
	flushCh       chan struct{}
	doneCh        chan struct{}
	flushInterval time.Duration // 刷新间隔
	stopped       bool          // 是否已停止
	stopMu        sync.Mutex    // 保护 stopped 字段
}

func newBufferedWriteSyncer(ws zapcore.WriteSyncer, bufferSize int, flushInterval int) *bufferedWriteSyncer {
	if bufferSize <= 0 {
		bufferSize = 256
	}
	// 刷新间隔：默认 200ms，最小 50ms，最大 5s
	if flushInterval <= 0 {
		flushInterval = 200 // 默认 200ms
	} else if flushInterval < 50 {
		flushInterval = 50 // 最小 50ms
	} else if flushInterval > 5000 {
		flushInterval = 5000 // 最大 5s
	}

	bws := &bufferedWriteSyncer{
		ws:            ws,
		buffer:        buffer.NewPool().Get(),
		stopCh:        make(chan struct{}),
		flushCh:       make(chan struct{}, 1),
		doneCh:        make(chan struct{}),
		flushInterval: time.Duration(flushInterval) * time.Millisecond,
		stopped:       false,
	}

	go bws.flushLoop()

	// 注册到全局列表，用于 Close 时停止
	syncersMutex.Lock()
	bufferedSyncers = append(bufferedSyncers, bws)
	syncersMutex.Unlock()

	return bws
}

func (bws *bufferedWriteSyncer) Write(p []byte) (n int, err error) {
	bws.mu.Lock()
	defer bws.mu.Unlock()
	return bws.buffer.Write(p)
}

func (bws *bufferedWriteSyncer) Sync() error {
	bws.stopMu.Lock()
	stopped := bws.stopped
	bws.stopMu.Unlock()

	if stopped {
		// 已停止，直接刷新并同步
		bws.mu.Lock()
		if bws.buffer.Len() > 0 {
			data := bws.buffer.Bytes()
			_, _ = bws.ws.Write(data)
			bws.buffer.Reset()
		}
		bws.mu.Unlock()
		return bws.ws.Sync()
	}

	// 触发立即刷新（通过 flushCh）
	select {
	case bws.flushCh <- struct{}{}:
		// 成功发送刷新信号，flushLoop 会处理
		// 等待一小段时间确保刷新完成（非阻塞）
		time.Sleep(100 * time.Millisecond)
	default:
		// flushCh 已满，直接刷新（避免阻塞）
		bws.mu.Lock()
		if bws.buffer.Len() > 0 {
			data := bws.buffer.Bytes()
			_, _ = bws.ws.Write(data)
			bws.buffer.Reset()
		}
		bws.mu.Unlock()
		// 同步底层 WriteSyncer
		_ = bws.ws.Sync()
	}
	return nil
}

func (bws *bufferedWriteSyncer) flushLoop() {
	ticker := time.NewTicker(bws.flushInterval) // 使用配置的刷新间隔
	defer ticker.Stop()

	for {
		select {
		case <-bws.stopCh:
			// 停止时，最后一次刷新并同步
			bws.mu.Lock()
			if bws.buffer.Len() > 0 {
				data := bws.buffer.Bytes()
				_, _ = bws.ws.Write(data)
				bws.buffer.Reset()
			}
			bws.mu.Unlock()
			// 同步底层 WriteSyncer
			_ = bws.ws.Sync()
			close(bws.doneCh)
			return
		case <-ticker.C:
			// 定时刷新
			bws.mu.Lock()
			if bws.buffer.Len() > 0 {
				data := bws.buffer.Bytes()
				_, _ = bws.ws.Write(data)
				bws.buffer.Reset()
				// 定时刷新时同步
				_ = bws.ws.Sync()
			}
			bws.mu.Unlock()
		case <-bws.flushCh:
			// 立即刷新（由 Sync 触发）
			bws.mu.Lock()
			if bws.buffer.Len() > 0 {
				data := bws.buffer.Bytes()
				_, _ = bws.ws.Write(data)
				bws.buffer.Reset()
			}
			bws.mu.Unlock()
			// 立即刷新时同步
			_ = bws.ws.Sync()
		}
	}
}

func (bws *bufferedWriteSyncer) flush() {
	bws.mu.Lock()
	defer bws.mu.Unlock()

	if bws.buffer.Len() == 0 {
		return
	}

	// 写入底层 WriteSyncer（不调用 Sync，避免阻塞）
	// Sync 会在定时刷新时自动调用
	data := bws.buffer.Bytes()
	if len(data) > 0 {
		_, _ = bws.ws.Write(data)
		bws.buffer.Reset()
		// 不在这里调用 Sync，避免阻塞 flushLoop
		// Sync 会在定时刷新时由 flushLoop 调用
	}
}

func (bws *bufferedWriteSyncer) Stop() {
	bws.stopMu.Lock()
	if bws.stopped {
		bws.stopMu.Unlock()
		return
	}
	bws.stopped = true
	bws.stopMu.Unlock()

	close(bws.stopCh)
	// 等待 goroutine 退出，确保最后一次刷新完成
	select {
	case <-bws.doneCh:
		// goroutine 已退出，最后一次刷新已完成
	case <-time.After(2 * time.Second):
		// 超时，强制退出（但最后一次刷新应该已经完成）
	}
}

func defaultInt(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}
