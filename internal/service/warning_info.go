package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
	"errors"
)

type WarningInfoService struct {
	warningRepo    *repo.WarningInfoRepository
	deviceUserRepo *repo.DeviceUserRepository
}

func NewWarningInfoService() *WarningInfoService {
	return &WarningInfoService{
		warningRepo:    repo.NewWarningInfoRepository(),
		deviceUserRepo: repo.NewDeviceUserRepository(),
	}
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

// GetWarningInfoList 获取告警信息列表（带权限控制）
func (s *WarningInfoService) GetWarningInfoList(page, pageSize int, alertType, alertStatus string, devID, dataID *int64, uid int64, role string) ([]*model.WarningInfo, int64, error) {
	var devIDs []int64

	// 如果是普通用户，需要获取其有权限的设备列表
	if role != model.RoleAdmin {
		userDevIDs, err := s.deviceUserRepo.GetUserDeviceIDs(uid)
		if err != nil {
			return nil, 0, err
		}
		devIDs = userDevIDs

		// 如果用户没有任何设备权限，直接返回空列表
		if len(devIDs) == 0 {
			return []*model.WarningInfo{}, 0, nil
		}

		// 如果请求中指定了 devID，需要检查用户是否有权限访问该设备
		if devID != nil {
			hasPermission := false
			for _, id := range devIDs {
				if id == *devID {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				return nil, 0, errors.New("您没有权限查看该设备的告警信息")
			}
		}
	}

	// 管理员不传 devIDs，可以查看所有设备的告警
	return s.warningRepo.GetWarningInfoList(page, pageSize, alertType, alertStatus, devID, dataID, devIDs)
}

// UpdateWarningInfo 更新告警信息
func (s *WarningInfoService) UpdateWarningInfo(warning *model.WarningInfo) (*model.WarningInfo, error) {
	err := s.warningRepo.UpdateWarningInfo(warning)
	if err != nil {
		return nil, err
	}
	return warning, nil
}

// UpdateWarningStatus 更新告警状态（带权限控制）
// 如果状态变为 resolved，自动设置 resolved_at 时间
func (s *WarningInfoService) UpdateWarningStatus(alertID int64, alertStatus string, uid int64, role string) (*model.WarningInfo, error) {
	// 先获取原有的告警信息
	warning, err := s.warningRepo.GetWarningInfo(alertID)
	if err != nil {
		return nil, err
	}

	// 权限检查：普通用户需要验证是否有权限访问该设备
	if role != model.RoleAdmin {
		userDevIDs, err := s.deviceUserRepo.GetUserDeviceIDs(uid)
		if err != nil {
			return nil, err
		}

		hasPermission := false
		for _, devID := range userDevIDs {
			if devID == warning.DevID.Int64() {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			return nil, errors.New("您没有权限更新该告警信息")
		}
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

// DeleteWarningInfo 删除告警信息（带权限控制）
func (s *WarningInfoService) DeleteWarningInfo(alertID int64, uid int64, role string) error {
	// 先获取告警信息
	warning, err := s.warningRepo.GetWarningInfo(alertID)
	if err != nil {
		return err
	}

	// 权限检查：普通用户需要验证是否有权限访问该设备
	if role != model.RoleAdmin {
		userDevIDs, err := s.deviceUserRepo.GetUserDeviceIDs(uid)
		if err != nil {
			return err
		}

		hasPermission := false
		for _, devID := range userDevIDs {
			if devID == warning.DevID.Int64() {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			return errors.New("您没有权限删除该告警信息")
		}
	}

	return s.warningRepo.DeleteWarningInfo(alertID)
}

