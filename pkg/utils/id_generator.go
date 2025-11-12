package utils

import (
	"crypto/md5"
	"os"
	"strconv"
	"sync"
	"time"
)

// SnowflakeID 雪花算法ID生成器
type SnowflakeID struct {
	mutex        sync.Mutex
	epoch        int64 // 起始时间戳（毫秒）
	machineID    int64 // 机器ID（0-1023）
	datacenterID int64 // 数据中心ID（0-31）
	sequence     int64 // 序列号（0-4095）
	lastTime     int64 // 上次生成ID的时间戳
}

var (
	defaultSnowflake *SnowflakeID
	once             sync.Once
)

// NewSnowflakeID 创建雪花算法ID生成器
// machineID: 机器ID（0-1023）
// datacenterID: 数据中心ID（0-31）
func NewSnowflakeID(machineID, datacenterID int64) *SnowflakeID {
	if machineID < 0 || machineID > 1023 {
		panic("machineID must be between 0 and 1023")
	}
	if datacenterID < 0 || datacenterID > 31 {
		panic("datacenterID must be between 0 and 31")
	}

	return &SnowflakeID{
		epoch:        1609459200000, // 2021-01-01 00:00:00 UTC
		machineID:    machineID,
		datacenterID: datacenterID,
		sequence:     0,
		lastTime:     0,
	}
}

// GetDefaultSnowflake 获取默认的雪花算法ID生成器（单例模式）
// 优先从环境变量读取 SNOWFLAKE_MACHINE_ID 和 SNOWFLAKE_DATACENTER_ID
// 如果未设置，则基于主机名生成唯一值
func GetDefaultSnowflake() *SnowflakeID {
	once.Do(func() {
		machineID := getMachineIDFromEnv()
		datacenterID := getDatacenterIDFromEnv()

		// 如果环境变量未设置，基于主机名生成唯一值
		if machineID < 0 {
			machineID = generateMachineIDFromHostname()
		}
		if datacenterID < 0 {
			datacenterID = generateDatacenterIDFromHostname()
		}

		defaultSnowflake = NewSnowflakeID(machineID, datacenterID)
	})
	return defaultSnowflake
}

// getMachineIDFromEnv 从环境变量读取机器ID
// 返回 -1 表示未设置
func getMachineIDFromEnv() int64 {
	val := os.Getenv("SNOWFLAKE_MACHINE_ID")
	if val == "" {
		return -1
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil || id < 0 || id > 1023 {
		return -1
	}
	return id
}

// getDatacenterIDFromEnv 从环境变量读取数据中心ID
// 返回 -1 表示未设置
func getDatacenterIDFromEnv() int64 {
	val := os.Getenv("SNOWFLAKE_DATACENTER_ID")
	if val == "" {
		return -1
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil || id < 0 || id > 31 {
		return -1
	}
	return id
}

// generateMachineIDFromHostname 基于主机名生成机器ID（0-1023）
func generateMachineIDFromHostname() int64 {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// 使用 MD5 哈希主机名，取前 10 位作为机器ID
	hash := md5.Sum([]byte(hostname))
	machineID := int64(hash[0])<<2 | int64(hash[1])>>6
	return machineID & 0x3FF // 确保在 0-1023 范围内
}

// generateDatacenterIDFromHostname 基于主机名生成数据中心ID（0-31）
func generateDatacenterIDFromHostname() int64 {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// 使用 MD5 哈希主机名，取前 5 位作为数据中心ID
	hash := md5.Sum([]byte(hostname))
	datacenterID := int64(hash[2]) >> 3
	return datacenterID & 0x1F // 确保在 0-31 范围内
}

// Generate 生成雪花算法ID
func (s *SnowflakeID) Generate() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now().UnixMilli()

	// 如果当前时间小于上次生成ID的时间，说明时钟回拨
	if now < s.lastTime {
		// 等待时钟追上
		time.Sleep(time.Duration(s.lastTime-now) * time.Millisecond)
		now = time.Now().UnixMilli()
	}

	// 如果是同一毫秒内生成的
	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & 0xFFF // 序列号最大4095
		// 如果序列号溢出，等待下一毫秒
		if s.sequence == 0 {
			now = s.waitNextMillis(s.lastTime)
		}
	} else {
		// 新的毫秒，序列号重置
		s.sequence = 0
	}

	s.lastTime = now

	// 生成ID：时间戳(41位) + 数据中心ID(5位) + 机器ID(10位) + 序列号(12位)
	id := ((now - s.epoch) << 22) | (s.datacenterID << 17) | (s.machineID << 12) | s.sequence

	return id
}

// waitNextMillis 等待下一毫秒
func (s *SnowflakeID) waitNextMillis(lastTime int64) int64 {
	now := time.Now().UnixMilli()
	for now <= lastTime {
		now = time.Now().UnixMilli()
	}
	return now
}

// GenerateID 生成雪花算法ID（使用默认生成器）
func GenerateID() int64 {
	return GetDefaultSnowflake().Generate()
}

// GenerateIDString 生成雪花算法ID字符串
func GenerateIDString() string {
	return Int64ToString(GenerateID())
}
