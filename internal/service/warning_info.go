package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
)

type WarningInfoService struct {
	warningRepo *repo.WarningInfoRepository
}

func NewWarningInfoService() *WarningInfoService {
	return &WarningInfoService{warningRepo: repo.NewWarningInfoRepository()}
}

// CreateWarningInfo 创建告警信息
func (s *WarningInfoService) CreateWarningInfo(warning *model.WarningInfo) (*model.WarningInfo, error) {
	warning.AlertID = utils.GetDefaultSnowflake().Generate()
	if warning.TriggeredAt.IsZero() {
		warning.TriggeredAt = utils.GetCurrentTime()
	}

	// ResolvedAt 默认为 nil (NULL)，表示未解决
	// 不需要额外处理，指针类型默认就是 nil

	err := s.warningRepo.CreateWarningInfo(warning)
	if err != nil {
		return nil, err
	}

	return warning, nil
}

// GetWarningInfo 获取告警信息
func (s *WarningInfoService) GetWarningInfo(alertID int64) (*model.WarningInfo, error) {
	return s.warningRepo.GetWarningInfo(alertID)
}

// GetWarningInfoList 获取告警信息列表
func (s *WarningInfoService) GetWarningInfoList(page, pageSize int, alertType, alertStatus string, devID, dataID *int64) ([]*model.WarningInfo, int64, error) {
	return s.warningRepo.GetWarningInfoList(page, pageSize, alertType, alertStatus, devID, dataID)
}

// UpdateWarningInfo 更新告警信息
func (s *WarningInfoService) UpdateWarningInfo(warning *model.WarningInfo) (*model.WarningInfo, error) {
	err := s.warningRepo.UpdateWarningInfo(warning)
	if err != nil {
		return nil, err
	}
	return warning, nil
}

// UpdateWarningStatus 更新告警状态
// 如果状态变为 resolved，自动设置 resolved_at 时间
func (s *WarningInfoService) UpdateWarningStatus(alertID int64, alertStatus string) (*model.WarningInfo, error) {
	// 先获取原有的告警信息
	warning, err := s.warningRepo.GetWarningInfo(alertID)
	if err != nil {
		return nil, err
	}

	// 更新状态
	warning.AlertStatus = alertStatus

	// 如果状态变为 resolved，设置 resolved_at 时间
	if alertStatus == "resolved" && warning.ResolvedAt == nil {
		now := utils.GetCurrentTime()
		warning.ResolvedAt = &now
	}

	// 更新到数据库
	err = s.warningRepo.UpdateWarningStatus(alertID, alertStatus, warning.ResolvedAt)
	if err != nil {
		return nil, err
	}

	return warning, nil
}

// DeleteWarningInfo 删除告警信息
func (s *WarningInfoService) DeleteWarningInfo(alertID int64) error {
	return s.warningRepo.DeleteWarningInfo(alertID)
}
