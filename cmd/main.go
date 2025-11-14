package main

import (
	"backend/config"
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/internal/db/mysql"
	"backend/internal/route"
	"backend/pkg/logger"
	"backend/pkg/utils"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var AppClient *Client

type Client struct {
	InfluxDB *influxdb.InfluxDBClient
	MinIO    *minio.MinIOClient
	Mysql    *mysql.MysqlClient
}

func InitClient(cfg *config.Config) error {

	err := logger.InitGlobal(cfg.Logger)
	if err != nil {
		return fmt.Errorf("初始化 logger 失败: %v", err)
	}
	logger.L().Info("Logger 初始化成功")

	// 3. 初始化Client
	AppClient = &Client{}

	// 4. 初始化InfluxDB客户端
	AppClient.InfluxDB, err = influxdb.GetInfluxDBClient(cfg.InfluxDB)
	if err != nil {
		return fmt.Errorf("初始化 InfluxDB 失败: %v", err)
	}

	// 5. 初始化MinIO客户端
	AppClient.MinIO, err = minio.GetMinIOClient(cfg.MinIO)
	if err != nil {
		return fmt.Errorf("初始化 MinIO 失败: %v", err)
	}

	// 6. 初始化Mysql客户端
	AppClient.Mysql, err = mysql.GetMysqlClient(cfg.Mysql)
	if err != nil {
		return fmt.Errorf("初始化 MySQL 失败: %v", err)
	}

	logger.L().Info("所有客户端初始化成功")
	return nil
}

func GetClient() *Client {
	return AppClient
}

func CloseClient() {
	// 关闭InfluxDB客户端
	if AppClient != nil && AppClient.InfluxDB != nil {
		AppClient.InfluxDB.Close()
	}

	// 关闭MinIO客户端,一般自动关闭

	// 关闭Mysql客户端
	if AppClient != nil && AppClient.Mysql != nil {
		AppClient.Mysql.Close()
	}

	// 记录日志（在关闭 logger 之前）
	logger.L().Info("客户端关闭成功")

	// 关闭Logger
	logger.Close()
}

func Init(cfg *config.Config) error {

	// 初始化客户端
	if err := InitClient(cfg); err != nil {
		return fmt.Errorf("初始化客户端失败: %v", err)
	}


	// 初始化JWT密钥
	utils.LoadJWTSecret(cfg)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	return nil
}

func main() {
	//  初始化配置
	cfg, err := config.InitConfig("./config/dev.yaml")
	if err != nil {
		fmt.Println("加载配置失败", err)
		os.Exit(1)
	}
	logger.L().Info("配置加载成功", logger.WithAny("cfg", cfg))

	// 初始化
	if err := Init(cfg); err != nil {
		logger.L().Error("初始化失败", logger.WithError(err))
		os.Exit(1)
	}
	defer CloseClient()

	// 默认配置
	r := gin.Default()
	//注册中间件
	// r.Use()

	//注册路由
	route.RegisterRoutes(r)

	// 启动服务器
	Addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    Addr,
		Handler: r,
	}

	//  在协程中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Fatal("服务器启动失败", logger.WithError(err))
		}
	}()
	logger.L().Info("服务器启动成功", logger.WithString("Listening ON:", Addr))

	// 监听操作系统中断信号
	quit := make(chan os.Signal, 1) // 建议使用带缓冲的通道
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞，直到收到信号
	logger.L().Info("开始优雅关闭服务器...")

	// 设置一个优雅关闭的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//  执行优雅关闭
	if err := srv.Shutdown(ctx); err != nil {
		logger.L().Fatal("服务器强制关闭", logger.WithError(err))
	}
	logger.L().Info("服务器已优雅关闭")
}
