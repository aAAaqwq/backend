# SensorDataHub Backend

传感器数据中心后端服务 - 基于 Go 语言的高性能物联网数据管理平台。

## 项目简介

SensorDataHub 是一个专为物联网传感器数据管理设计的后端服务，提供设备管理、时序数据存储、文件管理、告警通知等核心功能。

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.24 |
| Web 框架 | Gin |
| 关系数据库 | MySQL |
| 时序数据库 | InfluxDB3 |
| 对象存储 | MinIO |
| 日志框架 | Zap + Lumberjack |
| 配置管理 | Viper |
| 认证方式 | JWT |

## 项目结构

```
backend/
├── cmd/                    # 程序入口
│   └── main.go
├── config/                 # 配置管理
│   └── config.go
├── internal/               # 内部模块
│   ├── db/                 # 数据库连接
│   │   ├── influxdb/       # InfluxDB 客户端
│   │   ├── minio/          # MinIO 客户端
│   │   └── mysql/          # MySQL 客户端
│   ├── handler/            # HTTP 处理器
│   ├── middleware/         # 中间件
│   ├── model/              # 数据模型
│   ├── repo/               # 数据访问层
│   ├── route/              # 路由注册
│   └── service/            # 业务逻辑层
├── pkg/                    # 公共工具包
│   ├── logger/             # 日志工具
│   └── utils/              # 通用工具
├── doc/                    # API 文档
├── script/                 # 数据库脚本
├── data/                   # 数据目录
├── docker-compose.yml      # Docker 编排
└── Dockerfile              # 镜像构建
```

## 快速开始

### 环境要求

- Go 1.24+
- Docker & Docker Compose
- MySQL 8.0+
- InfluxDB3
- MinIO

### 使用 Docker Compose 部署

1. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，设置数据库密码等敏感信息
```

2. 启动服务

```bash
docker-compose up -d
```

3. 服务端口

| 服务 | 端口 |
|------|------|
| Backend API | 12000 |
| MySQL | 3306 |
| MinIO API | 9000 |
| MinIO Console | 9001 |
| InfluxDB | 8181 |
| InfluxDB UI | 8888 |

### 本地开发

1. 安装依赖

```bash
go mod download
```

2. 配置文件

```bash
# 复制并编辑配置文件
cp config/dev.yaml.example config/dev.yaml
```

3. 运行服务

```bash
go run cmd/main.go
```

## API 接口

基础路径: `/api/v1`

### 用户管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/users/register` | 用户注册 | - |
| POST | `/users/login` | 用户登录 | - |
| GET | `/users` | 获取当前用户信息 | JWT |
| GET | `/users/all` | 获取所有用户 | JWT + Admin |
| PUT | `/users` | 更新用户信息 | JWT |
| PUT | `/users/password` | 修改密码 | JWT |
| DELETE | `/users` | 删除用户 | JWT + Admin |
| GET | `/users/bind_devices` | 获取用户绑定的设备 | JWT |

### 设备管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/devices` | 创建设备 | JWT |
| GET | `/devices` | 获取设备列表 | JWT |
| PUT | `/devices` | 更新设备信息 | JWT |
| DELETE | `/devices` | 删除设备 | JWT |
| GET | `/devices/statistics` | 获取设备统计 | JWT |

### 设备用户绑定

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/devices/bind_user` | 绑定设备与用户 | JWT |
| DELETE | `/devices/unbind_user` | 解绑设备与用户 | JWT |
| GET | `/devices/bind_users` | 获取设备绑定的用户 | JWT |

### 传感器数据

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/device/data` | 上传传感器数据 | JWT |
| POST | `/device/data/timeseries` | 查询时序数据 | JWT |
| DELETE | `/device/data/timeseries` | 删除时序数据 | JWT |
| GET | `/device/data/statistic` | 获取数据统计 | JWT |
| POST | `/device/data/file/presigned_url` | 获取文件上传预签名URL | JWT |
| GET | `/device/data/file/list` | 获取文件列表 | JWT |
| GET | `/device/data/file/download` | 下载文件 | JWT |
| DELETE | `/device/data/file` | 删除文件 | JWT |

### 告警信息

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/warning_info` | 创建告警 | JWT |
| GET | `/warning_info` | 获取告警列表 | JWT |
| PUT | `/warning_info` | 更新告警信息 | JWT |
| DELETE | `/warning_info` | 删除告警 | JWT |

### 系统日志

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/logs` | 创建日志 | JWT + Admin |
| GET | `/logs` | 获取日志列表 | JWT + Admin |
| DELETE | `/logs` | 删除日志 | JWT + Admin |

## 配置说明

配置文件位于 `config/` 目录，支持 YAML 格式：

```yaml
server:
  host: "0.0.0.0"
  port: "12000"
  mode: "release"  # debug/release

mysql:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your_password"
  database: "sensor_hub"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10
  max_lifetime: 60

influxdb:
  host: "http://localhost:8181"
  token: "your_token"
  database: "sensor_data"

minio:
  endpoint: "localhost:9000"
  access_key_id: "your_access_key"
  secret_access_key: "your_secret_key"
  use_ssl: false
  region: "us-east-1"

jwt:
  secret: "your_jwt_secret"

logger:
  level: "info"
  encoding: "json"
  development: false
  enableCaller: true
  serviceName: "sensor-data-hub"
  environment: "prod"
```

## 数据模型

### Device (设备)

| 字段 | 类型 | 描述 |
|------|------|------|
| dev_id | int64 | 设备ID |
| dev_name | string | 设备名称 |
| dev_status | int | 状态 (0:离线, 1:在线, 2:异常) |
| dev_type | string | 设备类型 |
| dev_power | int | 电量 |
| model | string | 硬件型号 |
| version | string | 硬件版本 |
| sampling_rate | int | 采样频率 |
| upload_interval | int | 上报间隔 |
| offline_threshold | int | 离线判断阈值 |
| extended_config | json | 扩展配置 |

### SensorData (传感器数据)

时序数据点结构:

| 字段 | 类型 | 描述 |
|------|------|------|
| measurement | string | 测量名称 |
| tags | map | 标签 |
| fields | map | 字段值 |
| timestamp | int64 | 时间戳 |

## 许可证

MIT License
