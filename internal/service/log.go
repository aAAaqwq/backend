package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
)

type LogService struct {
	logRepo *repo.LogRepository
}

func NewLogService() *LogService {
	return &LogService{logRepo: repo.NewLogRepository()}
}

// CreateLog 创建日志
func (s *LogService) CreateLog(log *model.Log) (*model.Log, error) {
	log.LogID = utils.GetDefaultSnowflake().Generate()
	log.CreateAt = utils.GetCurrentTime()

	err := s.logRepo.CreateLog(log)
	if err != nil {
		return nil, err
	}

	return log, nil
}

// GetLog 获取日志
func (s *LogService) GetLog(logID int64) (*model.Log, error) {
	return s.logRepo.GetLog(logID)
}

// GetLogs 查询日志列表
func (s *LogService) GetLogs(logType string, level *int, startTime, endTime string) ([]*model.Log, error) {
	return s.logRepo.GetLogs(logType, level, startTime, endTime)
}

// DeleteLog 删除日志
func (s *LogService) DeleteLog(logID int64) error {
	return s.logRepo.DeleteLog(logID)
}
