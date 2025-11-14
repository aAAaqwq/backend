package influxdb

import (
	"backend/config"
	"backend/pkg/logger"
	"context"
	"fmt"
	"strings"
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

// WritePoint 写入单个数据点到InfluxDB
// measurement: 测量名称（表名）
// tags: 标签（索引字段）
// fields: 字段（数据字段）
// timestamp: 时间戳（纳秒）
func (c *InfluxDBClient) WritePoint(measurement string, tags map[string]string, fields map[string]interface{}, timestamp int64) error {
	ctx := context.Background()

	// 构建Line Protocol格式的数据点
	var pointBuilder strings.Builder
	pointBuilder.WriteString(measurement)

	// 添加tags
	for k, v := range tags {
		// 转义特殊字符
		k = escapeTagKey(k)
		v = escapeTagValue(v)
		pointBuilder.WriteString(fmt.Sprintf(",%s=%s", k, v))
	}
	pointBuilder.WriteString(" ")

	// 添加fields
	fieldParts := make([]string, 0, len(fields))
	for k, v := range fields {
		k = escapeFieldKey(k)
		var fieldValue string
		switch val := v.(type) {
		case string:
			fieldValue = fmt.Sprintf("%s=\"%s\"", k, escapeStringValue(val))
		case int, int32, int64:
			fieldValue = fmt.Sprintf("%s=%di", k, val)
		case float32, float64:
			fieldValue = fmt.Sprintf("%s=%f", k, val)
		case bool:
			fieldValue = fmt.Sprintf("%s=%t", k, val)
		default:
			fieldValue = fmt.Sprintf("%s=\"%s\"", k, escapeStringValue(fmt.Sprintf("%v", val)))
		}
		fieldParts = append(fieldParts, fieldValue)
	}

	if len(fieldParts) > 0 {
		pointBuilder.WriteString(fieldParts[0])
		for i := 1; i < len(fieldParts); i++ {
			pointBuilder.WriteString("," + fieldParts[i])
		}
	}

	// 添加时间戳（纳秒）
	if timestamp > 0 {
		pointBuilder.WriteString(fmt.Sprintf(" %d", timestamp))
	}

	// 写入数据（InfluxDB3使用Write方法，参数为[]byte）
	err := c.Client.Write(ctx, []byte(pointBuilder.String()))
	if err != nil {
		return fmt.Errorf("写入InfluxDB失败: %v", err)
	}

	return nil
}

// escapeTagKey 转义tag key中的特殊字符
func escapeTagKey(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ",", "\\,"), "=", "\\="), " ", "\\ ")
}

// escapeTagValue 转义tag value中的特殊字符
func escapeTagValue(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ",", "\\,"), "=", "\\="), " ", "\\ ")
}

// escapeFieldKey 转义field key中的特殊字符
func escapeFieldKey(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ",", "\\,"), "=", "\\="), " ", "\\ ")
}

// escapeStringValue 转义字符串值中的特殊字符
func escapeStringValue(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\"", "\\\""), "\\", "\\\\")
}

// WritePoints 批量写入数据点到InfluxDB
func (c *InfluxDBClient) WritePoints(points []string) error {
	ctx := context.Background()

	for _, point := range points {
		err := c.Client.Write(ctx, []byte(point))
		if err != nil {
			return fmt.Errorf("批量写入InfluxDB失败: %v", err)
		}
	}

	return nil
}

// Query 查询InfluxDB数据
func (c *InfluxDBClient) Query(query string) (interface{}, error) {
	ctx := context.Background()

	result, err := c.Client.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询InfluxDB失败: %v", err)
	}

	return result, nil
}

// QuerySensorData 查询传感器时序数据
// devID: 设备ID
// startTime: 开始时间（Unix时间戳，秒）
// endTime: 结束时间（Unix时间戳，秒）
// limit: 限制返回数量
func (c *InfluxDBClient) QuerySensorData(devID int64, startTime, endTime *int64, limit int) (interface{}, error) {
	// InfluxDB 3.0使用SQL语法
	query := fmt.Sprintf(`SELECT * FROM sensor_data WHERE dev_id = %d`, devID)

	if startTime != nil {
		// 转换为纳秒时间戳
		nsTimestamp := *startTime * 1000000000
		query += fmt.Sprintf(" AND time >= %d", nsTimestamp)
	}
	if endTime != nil {
		// 转换为纳秒时间戳
		nsTimestamp := *endTime * 1000000000
		query += fmt.Sprintf(" AND time <= %d", nsTimestamp)
	}

	query += " ORDER BY time DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	return c.Query(query)
}

// DeleteSensorData 删除传感器数据
// devID: 设备ID
// startTime: 开始时间（Unix时间戳，秒）
// endTime: 结束时间（Unix时间戳，秒）
func (c *InfluxDBClient) DeleteSensorData(devID int64, startTime, endTime *int64) error {
	ctx := context.Background()

	// InfluxDB 3.0使用SQL语法进行删除
	query := fmt.Sprintf(`DELETE FROM sensor_data WHERE dev_id = %d`, devID)

	if startTime != nil {
		nsTimestamp := *startTime * 1000000000
		query += fmt.Sprintf(" AND time >= %d", nsTimestamp)
	}
	if endTime != nil {
		nsTimestamp := *endTime * 1000000000
		query += fmt.Sprintf(" AND time <= %d", nsTimestamp)
	}

	// 执行删除查询
	_, err := c.Client.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("删除InfluxDB数据失败: %v", err)
	}

	return nil
}

// DeleteSensorDataByDataID 根据数据ID删除传感器数据（需要从metadata中获取时间范围）
func (c *InfluxDBClient) DeleteSensorDataByDataID(devID int64, timestamp int64) error {
	// 删除指定时间点的数据
	startTime := timestamp - 1 // 前后各1秒范围
	endTime := timestamp + 1

	return c.DeleteSensorData(devID, &startTime, &endTime)
}
