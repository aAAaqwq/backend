package main

import (
	"backend/config"
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/internal/db/mysql"
	"backend/pkg/logger"
	"fmt"
	"os"
)

var AppClient *Client

type Client struct {
	InfluxDB *influxdb.InfluxDBClient
	MinIO    *minio.MinIOClient
	Mysql    *mysql.MysqlClient
}

func InitClient() error {
	// 1. 初始化配置
	cfg, err := config.InitConfig("./config/dev.yaml")
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}
	logger.L().Info("配置加载成功", logger.WithAny("cfg", cfg))

	err = logger.InitGlobal(cfg.Logger)
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

func main() {
	// 初始化客户端
	if err := InitClient(); err != nil {
		fmt.Println("初始化客户端失败", err)
		os.Exit(1)
	}

	defer CloseClient()

}
