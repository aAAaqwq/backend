package influxdb

import (
	"backend/config"
	"backend/internal/model"
	"backend/pkg/logger"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

const (
	MaxBatchSize     = 5000            // 最大批量写入大小
	MaxRetries       = 3               // 最大重试次数
	Timeout          = 5 * time.Second // 超时时间
	LimitQueryPoints = 6000            // 查询限制点数
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type QueryOptions struct {
	Measurement string            // 必填：emg / ecg / temperature ...
	Tags        map[string]string // 可选：data_id/dev_id/channel 等精确匹配
	Fields      []string          // 要哪些字段，为空表示全字段（*）

	TimeRange       TimeRange     // 必填：时间窗口
	DownsampleEvery time.Duration // 下采样间隔，可空；为空则自动根据 LimitPoints 算
	Aggregate       string        // "", "mean", "max", "min", "sum" 等
	LimitPoints     int           // 最大点数，没填给个默认值（例如 6000）
}

// InfluxDBCli InfluxDB客户端实例
var InfluxDBCli *InfluxDBClient

type InfluxDBClient struct {
	Client *influxdb3.Client
}

// GetInfluxDBClient 获取InfluxDB客户端
// 使用 config.InfluxDBConfig 作为参数类型
func GetInfluxDBClient(cfg config.InfluxDBConfig) (*InfluxDBClient, error) {
	if InfluxDBCli != nil {
		return InfluxDBCli, nil
	}
	client, err := InitInfluxDBClient(cfg)
	if err != nil {
		return nil, err
	}
	InfluxDBCli = &InfluxDBClient{Client: client}
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

	// 测试连接并检查数据库是否存在
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试查询数据库列表来检查数据库是否存在
	// InfluxDB v3 使用 SHOW DATABASES 命令
	showDbQuery := "SHOW DATABASES"
	iter, err := client.Query(ctx, showDbQuery)
	if err != nil {
		return nil, fmt.Errorf("InfluxDB连接测试失败: %v", err)
	}

	// 检查数据库是否存在
	databaseExists := false
	for iter.Next() {
		record := iter.Value()
		// 遍历记录查找数据库名称
		for _, val := range record {
			if dbName, ok := val.(string); ok && dbName == cfg.Database {
				databaseExists = true
				break
			}
		}
		if databaseExists {
			break
		}
	}

	// 如果数据库不存在，则创建
	if !databaseExists {
		logger.L().Info("数据库不存在，正在创建", logger.WithString("database", cfg.Database))
		createDbQuery := fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.Database)
		_, err = client.Query(ctx, createDbQuery)
		if err != nil {
			return nil, fmt.Errorf("创建数据库失败: %v", err)
		}
		logger.L().Info("数据库创建成功", logger.WithString("database", cfg.Database))
	} else {
		logger.L().Info("数据库已存在", logger.WithString("database", cfg.Database))
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
func (c *InfluxDBClient) WritePoints(points []model.Point) error {
	if len(points) == 0 {
		return nil
	}
	for i := 0; i < len(points); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(points) {
			end = len(points)
		}
		batchPoints := points[i:end]
		err := c.WriteBatchOnce(batchPoints)
		if err != nil {
			return fmt.Errorf("批量写入InfluxDB失败: %v", err)
		}
	}
	return nil
}

// WriteBatchOnce 一次性写入一批数据
func (c *InfluxDBClient) WriteBatchOnce(batch []model.Point) error {
	sdkPoints := make([]*influxdb3.Point, 0, len(batch))
	for _, p := range batch {
		pt := &influxdb3.Point{
			Values: &influxdb3.PointValues{
				Tags:   p.Tags, // Tags已经是map[string]string类型
				Fields: p.Fields,
			},
		}
		pt.SetMeasurement(p.Measurement)
		// timestamp是Unix时间戳（秒），需要转换为纳秒
		pt.SetTimestamp(time.Unix(p.Timestamp, 0))
		sdkPoints = append(sdkPoints, pt)
	}

	// 带超时 & 简单重试
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if err := c.Client.WritePoints(ctx, sdkPoints); err != nil {
			if !isRetryable(err) || attempt == MaxRetries {
				return err
			}
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}

// isRetryable 判断是否可重试
func isRetryable(err error) bool {
	return true
}

// Query 查询InfluxDB数据
func (c *InfluxDBClient) Query(opts QueryOptions) (interface{}, error) {
	// 基本参数校验
	if opts.Measurement == "" {
		return nil, fmt.Errorf("measurement is required")
	}
	if opts.TimeRange.Start.IsZero() || opts.TimeRange.End.IsZero() {
		return nil, fmt.Errorf("time range is required")
	}

	// 默认点数限制，避免一次拉爆
	if opts.LimitPoints == 0 {
		opts.LimitPoints = 6000
	}

	// 自动计算下采样间隔（没指定的话）
	if opts.DownsampleEvery == 0 && opts.LimitPoints > 0 {
		dur := opts.TimeRange.End.Sub(opts.TimeRange.Start)
		step := dur / time.Duration(opts.LimitPoints)
		if step < time.Millisecond {
			step = time.Millisecond
		}
		opts.DownsampleEvery = step
	}

	// 1. 拼 SQL
	sql, err := buildQuerySQL(opts)
	if err != nil {
		return nil, err
	}

	logger.L().Info("构建SQL语句完成", logger.WithAny("sql",sql))

	// 2. 设定超时
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	// 3. 简单重试（可选）
	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		points, err := c.queryOnce(ctx, sql)
		if err == nil {
			return points, nil
		}
		if !isRetryable(err) || attempt == maxRetries {
			return nil, err
		}
		lastErr = err
		time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
	}
	return nil, lastErr
}

// buildSQL 构建SQL查询语句
func buildQuerySQL(opts QueryOptions) (string, error) {
	if opts.Measurement == "" {
		return "", fmt.Errorf("measurement required")
	}

	var sb strings.Builder

	// 1. SELECT
	if opts.DownsampleEvery > 0 && opts.Aggregate != "" {
		// 下采样 + 聚合
		agg := strings.ToUpper(opts.Aggregate) // "MEAN" / "AVG" / "MAX" / "MIN"...
		// 这里只演示对单字段 value 下采样
		sb.WriteString("SELECT ")
		sb.WriteString(fmt.Sprintf(
			"time_bucket('%s', time) AS bucket, %s(value) AS value",
			opts.DownsampleEvery.String(),
			agg,
		))
	} else {
		// 原始点
		if len(opts.Fields) == 0 {
			// 没有指定fields，查询所有字段
			sb.WriteString("SELECT *")
		} else {
			// 指定了fields，去重后查询
			fieldSet := make(map[string]bool)
			var uniqueFields []string

			// 去重fields
			for _, f := range opts.Fields {
				if !fieldSet[f] {
					fieldSet[f] = true
					uniqueFields = append(uniqueFields, f)
				}
			}

			sb.WriteString("SELECT ")
			sb.WriteString(strings.Join(uniqueFields, ", "))
		}
	}

	// 2. FROM
	sb.WriteString(" FROM ")
	sb.WriteString(`"` + opts.Measurement + `"`)

	// 3. WHERE
	var conds []string
	// InfluxDB v3的time字段是Timestamp类型，需要使用timestamp literal格式
	// 格式：'2024-12-03T12:00:00Z' 或使用 CAST
	startTime := opts.TimeRange.Start.Format(time.RFC3339Nano)
	endTime := opts.TimeRange.End.Format(time.RFC3339Nano)
	conds = append(conds,
		fmt.Sprintf("time >= '%s'", startTime),
	)
	conds = append(conds,
		fmt.Sprintf("time <= '%s'", endTime),
	)
	for k, v := range opts.Tags {
		conds = append(conds, fmt.Sprintf(`%s = '%s'`, k, v))
	}
	if len(conds) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conds, " AND "))
	}

	// 4. GROUP BY / ORDER BY
	if opts.DownsampleEvery > 0 && opts.Aggregate != "" {
		sb.WriteString(" GROUP BY bucket")
		sb.WriteString(" ORDER BY bucket")
	} else {
		sb.WriteString(" ORDER BY time")
	}

	// 5. LIMIT
	if opts.LimitPoints > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", opts.LimitPoints))
	}

	return sb.String(), nil
}

// queryOnce 执行一次查询
func (c *InfluxDBClient) queryOnce(ctx context.Context, sql string) (interface{}, error) {
	it, err := c.Client.QueryPointValue(ctx, sql)
	if err != nil {
		return nil, err
	}

	var points []model.Point
	for {
		pv, err := it.Next()
		if err == influxdb3.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// PointValue 里已经把 measurement / tags / fields / time 解好
		// 具体字段名看 influxdb3-go 的 PointValue 定义，这里按常见写法示例
		p := model.Point{
			Timestamp: pv.Timestamp.Unix(), // 转换为Unix时间戳（秒）
			Fields:    make(map[string]interface{}),
			Tags:      make(map[string]string), // tags为string类型
		}

		// 1）tags（转换为string类型）
		for k, v := range pv.Tags {
			p.Tags[k] = fmt.Sprintf("%v", v)
		}
		// 2）fields（按需要做类型断言）
		for k, v := range pv.Fields {
			p.Fields[k] = v
		}

		points = append(points, p)
	}

	return points, nil
}

// GetSeriesDataStatistics 获取时序数据统计信息
func (c *InfluxDBClient) GetSeriesDataStatistics(measurement string, devID int64) (map[string]interface{}, error) {
	ctx := context.Background()
	stats := make(map[string]interface{})

	// 构建查询SQL（InfluxDB3使用SQL语法）
	// 1. 查询总数
	countSQL := fmt.Sprintf(
		`SELECT COUNT(*) as count_value FROM "%s" WHERE dev_id = '%d'`,
		measurement, devID,
	)

	// 执行总数查询
	it, err := c.Client.QueryPointValue(ctx, countSQL)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}

	var total int64 = 0
	for {
		pv, err := it.Next()
		if err == influxdb3.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取总数结果失败: %v", err)
		}

		// 尝试从 Fields 中获取 count_value
		if val, ok := pv.Fields["count_value"]; ok {
			switch v := val.(type) {
			case int64:
				total = v
			case float64:
				total = int64(v)
			case int:
				total = int64(v)
			}
			break
		}
	}
	stats["total"] = total

	// 2. 查询异常数据数量（quality_score < 30，quality_score 是 tag，存储为字符串）
	// 需要将字符串转换为数字进行比较
	abnormalSQL := fmt.Sprintf(
		`SELECT COUNT(*) as count_value FROM "%s" WHERE dev_id = '%d' AND CAST(quality_score AS DOUBLE) < 30`,
		measurement, devID,
	)

	// 执行异常数据查询
	it, err = c.Client.QueryPointValue(ctx, abnormalSQL)
	if err != nil {
		// 如果 CAST 失败，尝试字符串比较（不太准确但作为 fallback）
		abnormalSQL = fmt.Sprintf(
			`SELECT COUNT(*) as count_value FROM "%s" WHERE dev_id = '%d' AND quality_score < '30'`,
			measurement, devID,
		)
		it, err = c.Client.QueryPointValue(ctx, abnormalSQL)
		if err != nil {
			// 如果还是失败，返回 0
			stats["abnormal"] = int64(0)
			return stats, nil
		}
	}

	var abnormal int64 = 0
	for {
		pv, err := it.Next()
		if err == influxdb3.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取异常数据结果失败: %v", err)
		}

		// 尝试从 Fields 中获取 count_value
		if val, ok := pv.Fields["count_value"]; ok {
			switch v := val.(type) {
			case int64:
				abnormal = v
			case float64:
				abnormal = int64(v)
			case int:
				abnormal = int64(v)
			}
			break
		}
	}
	stats["abnormal"] = abnormal

	return stats, nil
}
