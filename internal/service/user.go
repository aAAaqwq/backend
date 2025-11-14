package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
	"errors"
)

type UserService struct {
	userRepo *repo.UserRepository
}

func NewUserService() *UserService {
	return &UserService{userRepo: repo.NewUserRepository()}
}

func (s *UserService) Register(user *model.User) error {
	// 检查用户是否已存在
	if _, err := s.GetUser(user); err == nil {
		return errors.New("用户已存在")
	}
	// 生成UID
	user.UID = utils.GetDefaultSnowflake().Generate()
	// 生成用户名
	user.Username = "用户" + utils.ConvertToString(user.UID)
	// 密码哈希
	hash, err := utils.CreatePasswordHash(user.Password)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.CreateAt = utils.GetCurrentTime()
	user.UpdateAt = utils.GetCurrentTime()
	// 设置默认角色
	user.Role = model.RoleUser
	return s.userRepo.CreateUser(user)
}
func (s *UserService) Login(user *model.User) (map[string]interface{}, error) {
	// 从数据库中查询用户
	dbUser, err := s.userRepo.GetUserByEmail(user.Email)
	if err != nil {
		return nil, err
	}
	if dbUser == nil {
		return nil, errors.New("用户不存在")
	}
	// 验证密码
	if !utils.CheckPasswordHash(user.Password, dbUser.PasswordHash) {
		return nil, errors.New("密码错误")
	}

	// 生成JWT token
	token, err := utils.GenerateToken(dbUser)
	if err != nil {
		return nil, errors.New("生成token失败: " + err.Error())
	}

	// 返回token和用户信息
	return map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"uid":      dbUser.UID,
			"username": dbUser.Username,
			"email":    dbUser.Email,
			"role":     dbUser.Role,
		},
		"expires_in": 24 * 3600, // 24小时，单位：秒
	}, nil
}

func (s *UserService) GetUser(user *model.User) (*model.User, error) {
	// 如果有UID，根据UID查询
	if !utils.IsEmpty(user.UID) {
		return s.userRepo.GetUserByUID(user.UID)
	}
	// 如果有Email，根据Email查询
	if !utils.IsEmpty(user.Email) {
		return s.userRepo.GetUserByEmail(user.Email)
	}
	// 如果有Username，根据Username查询
	if !utils.IsEmpty(user.Username) {
		return s.userRepo.GetUserByUsername(user.Username)
	}
	return nil, nil
}

func (s *UserService) UpdateUserInfo(user *model.User) (*model.User, error) {
	user.UpdateAt = utils.GetCurrentTime()
	err := s.userRepo.UpdateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) ChangePassword(req *model.ChangePasswordReq) error {
	// 获取用户信息
	dbUser, err := s.userRepo.GetUserByUID(req.UID)
	if err != nil {
		return err
	}
	if dbUser == nil {
		return errors.New("用户不存在")
	}
	// 验证旧密码
	if !utils.CheckPasswordHash(req.OldPassword, dbUser.PasswordHash) {
		return errors.New("旧密码错误")
	}
	// 密码哈希
	hash, err := utils.CreatePasswordHash(req.NewPassword)
	if err != nil {
		return err
	}
	dbUser.PasswordHash = hash
	dbUser.UpdateAt = utils.GetCurrentTime()
	return s.userRepo.UpdateUser(dbUser)
}

func (s *UserService) DeleteUser(uid int64) error {
	return s.userRepo.DeleteUser(uid)
}

// GetUsers 获取用户列表（分页）
func (s *UserService) GetUsers(page, pageSize int, role, keyword, sortBy, sortOrder string) ([]*model.User, int64, error) {
	return s.userRepo.GetUsers(page, pageSize, role, keyword, sortBy, sortOrder)
}

// GetUserDevices 获取用户绑定的设备列表
func (s *UserService) GetUserDevices(uid int64, page, pageSize int, devType string, devStatus *int, permissionLevel string, isActive *bool) ([]*model.Device, int64, error) {
	deviceUserRepo := repo.NewDeviceUserRepository()
	return deviceUserRepo.GetUserDevices(uid, page, pageSize, devType, devStatus, permissionLevel, isActive)
}
