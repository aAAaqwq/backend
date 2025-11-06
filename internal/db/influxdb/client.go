package influxdb

import (
	"backend/config"
	"backend/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

// InfluxDBCli InfluxDB客户端实例
var InfluxDBCli *influxdb3.Client

type InfluxDBClient struct {
	Client *influxdb3.Client
}

// GetInfluxDBClient 获取InfluxDB客户端
// 使用 config.InfluxDBConfig 作为参数类型
func GetInfluxDBClient(cfg config.InfluxDBConfig) (*InfluxDBClient, error) {
	if InfluxDBCli != nil {
		return &InfluxDBClient{Client: InfluxDBCli}, nil
	}
	client, err := InitInfluxDBClient(cfg)
	if err != nil {
		return nil, err
	}
	InfluxDBCli = client
	return &InfluxDBClient{Client: client}, nil
}

// InitInfluxDBClient 初始化InfluxDB客户端
func InitInfluxDBClient(cfg config.InfluxDBConfig) (*influxdb3.Client, error) {
	// 创建InfluxDB客户端
	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:     cfg.Host,
		Token:    cfg.Token,
		Database: cfg.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("创建InfluxDB客户端失败: %v", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 可以通过执行一个简单的查询来测试连接
	query := "SELECT 1"
	_, err = client.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("InfluxDB连接测试失败: %v", err)
	}

	logger.L().Info("InfluxDB客户端初始化成功")
	return client, nil
}

func (c *InfluxDBClient) Close() error {
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}
