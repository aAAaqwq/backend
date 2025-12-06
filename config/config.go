package config

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// ==================== MySQL 配置 ====================
// MysqlConfig MySQL配置
type MysqlConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxLifetime  int    `yaml:"max_lifetime"` // 单位: 分钟
}

// ==================== InfluxDB 配置 ====================
// InfluxDBConfig InfluxDB配置
type InfluxDBConfig struct {
	Host     string `yaml:"host"`
	Token    string `yaml:"token"`
	Database string `yaml:"database"`
}

// ==================== MinIO 配置 ====================
// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	Region          string `yaml:"region"`
}

// ==================== Logger 配置 ====================
// LoggerFileRotate 文件轮转配置
type LoggerFileRotate struct {
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxSize"`    // 单个文件最大大小（MB），默认128
	MaxBackups int    `yaml:"maxBackups"` // 保留的备份文件数量，默认7
	MaxAge     int    `yaml:"maxAge"`     // 保留日志文件的最大天数，默认7
	Compress   bool   `yaml:"compress"`   // 是否压缩旧日志文件，默认false
	Enable     bool   `yaml:"enable"`     // 是否启用文件日志，默认false
}

// LoggerSamplingConfig 采样配置
type LoggerSamplingConfig struct {
	Enable     bool   `yaml:"enable"`     // 是否启用采样，默认false
	Tick       int    `yaml:"tick"`       // 采样时间间隔（秒），默认1
	First      int    `yaml:"first"`      // 第一个时间窗口内允许的日志条数，默认100
	Thereafter int    `yaml:"thereafter"` // 后续时间窗口内允许的日志条数，默认100
	Level      string `yaml:"level"`      // 采样级别（debug|info|warn|error），默认info
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level              string               `yaml:"level"`              // 日志级别：debug|info|warn|error|panic|fatal
	Encoding           string               `yaml:"encoding"`           // 编码格式：json|console
	Development        bool                 `yaml:"development"`        // 是否为开发模式，默认false
	EnableCaller       bool                 `yaml:"enableCaller"`       // 是否显示调用者信息，默认false
	EnableStacktrace   bool                 `yaml:"enableStacktrace"`   // 是否启用堆栈跟踪，默认false
	StacktraceLevel    string               `yaml:"stacktraceLevel"`    // 堆栈跟踪级别：debug|info|warn|error|panic|fatal，默认error
	OutputPaths        []string             `yaml:"outputPaths"`        // 输出路径列表，如：["stdout", "./logs/app.log"]
	ErrorOutputPaths   []string             `yaml:"errorOutputPaths"`   // 错误输出路径列表，如：["stderr"]
	ServiceName        string               `yaml:"serviceName"`        // 服务名称
	Environment        string               `yaml:"environment"`        // 环境：dev|test|prod
	File               LoggerFileRotate     `yaml:"file"`               // 文件轮转配置
	Sampling           LoggerSamplingConfig `yaml:"sampling"`           // 采样配置
	Async              bool                 `yaml:"async"`              // 是否异步写入文件（仅对文件有效，stdout/stderr始终同步），默认false
	AsyncBufferSize    int                  `yaml:"asyncBufferSize"`    // 异步缓冲区大小，默认256
	AsyncFlushInterval int                  `yaml:"asyncFlushInterval"` // 异步刷新间隔（毫秒），默认200ms
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}


type JWTConfig struct {
	Secret string `yaml:"secret"`
}
// ==================== 主配置结构 ====================
// Config 应用配置（集中管理所有配置）
type Config struct {
	Mysql    MysqlConfig    `yaml:"mysql"`
	InfluxDB InfluxDBConfig `yaml:"influxdb"`
	MinIO    MinIOConfig    `yaml:"minio"`
	Server   ServerConfig   `yaml:"server"`
	Logger   LoggerConfig   `yaml:"logger"`
	JWT      JWTConfig      `yaml:"jwt"`
}

// InitConfig 初始化配置（从YAML文件加载）
func InitConfig(path string) (*Config, error) {
	v := viper.New()

	if path == "" {
		path = "./config/dev.yaml"
	}
	v.SetConfigFile(path) // 绝对/相对路径都可
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %v", err)
	}

	var cfg Config
	// 让 viper 用 yaml 标签解码（否则用 mapstructure 标签）
	if err := v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "yaml"
	}); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %v", err)
	}

	return &cfg, nil
}
