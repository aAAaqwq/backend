package repo

import (
	"backend/internal/db/mysql"
	"backend/internal/model"
	"backend/pkg/utils"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) CreateUser(user *model.User) error {
	query := "INSERT INTO user (uid, role, username, email, password_hash) VALUES (?, ?, ?, ?, ?)"
	_, err := mysql.MysqlCli.Client.Exec(query,
		user.UID, user.Role, user.Username, user.Email, user.PasswordHash)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(user *model.User) error {
	query := "UPDATE user SET update_at = ? "
	args := []interface{}{user.UpdateAt}

	if !utils.IsEmpty(user.Role) {
		query += " , role = ?"
		args = append(args, user.Role)
	}
	if !utils.IsEmpty(user.Username) {
		query += " , username = ?"
		args = append(args, user.Username)
	}
	if !utils.IsEmpty(user.Email) {
		query += " , email = ?"
		args = append(args, user.Email)
	}
	if !utils.IsEmpty(user.PasswordHash) {
		query += " , password_hash = ?"
		args = append(args, user.PasswordHash)
	}

	query += " WHERE uid = ?"
	args = append(args, user.UID)

	_, err := mysql.MysqlCli.Client.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}

// DeleteUser 删除用户（级联删除user_dev中的绑定关系）
func (r *UserRepository) DeleteUser(uid int64) error {
	// 开启事务
	tx, err := mysql.MysqlCli.Client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. 先删除user_dev表中的绑定关系
	_, err = tx.Exec("DELETE FROM user_dev WHERE uid = ?", uid)
	if err != nil {
		return err
	}

	// 2. 再删除user表中的用户记录
	_, err = tx.Exec("DELETE FROM user WHERE uid = ?", uid)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}

func (r *UserRepository) GetUserByUID(uid int64) (*model.User, error) {
	var user model.User
	err := mysql.MysqlCli.Client.QueryRow("SELECT * FROM user WHERE uid = ?", uid).
		Scan(&user.UID,&user.Role,&user.Username,&user.Email,&user.PasswordHash,&user.CreateAt,&user.UpdateAt)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := mysql.MysqlCli.Client.QueryRow("SELECT * FROM user WHERE username = ?", username).
	Scan(&user.UID,&user.Role,&user.Username,&user.Email,&user.PasswordHash,&user.CreateAt,&user.UpdateAt)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := mysql.MysqlCli.Client.QueryRow("SELECT * FROM user WHERE email = ?", email).
	Scan(&user.UID,&user.Role,&user.Username,&user.Email,&user.PasswordHash,&user.CreateAt,&user.UpdateAt)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

// GetUsers 获取用户列表（分页）
func (r *UserRepository) GetUsers(page, pageSize int, role, keyword, sortBy, sortOrder string) ([]*model.User, int64, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if !utils.IsEmpty(role) {
		whereClause += " AND role = ?"
		args = append(args, role)
	}
	if !utils.IsEmpty(keyword) {
		whereClause += " AND (username LIKE ? OR email LIKE ?)"
		keywordPattern := "%" + keyword + "%"
		args = append(args, keywordPattern, keywordPattern)
	}

	// 排序
	if utils.IsEmpty(sortBy) {
		sortBy = "create_at"
	}
	if utils.IsEmpty(sortOrder) {
		sortOrder = "DESC"
	}
	orderClause := "ORDER BY " + sortBy + " " + sortOrder

	// 分页
	offset := (page - 1) * pageSize
	limitClause := "LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM user " + whereClause
	err := mysql.MysqlCli.Client.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := "SELECT uid, role, username, email, create_at, update_at FROM user " + whereClause + " " + orderClause + " " + limitClause
	rows, err := mysql.MysqlCli.Client.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(&user.UID, &user.Role, &user.Username, &user.Email, &user.CreateAt, &user.UpdateAt)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetUserDevices 获取用户绑定的设备列表（已迁移到DeviceUserRepository）
// 保留此方法以保持兼容性，实际调用DeviceUserRepository
func (r *UserRepository) GetUserDevices(uid int64, page, pageSize int, devType string, devStatus *int, permissionLevel string, isActive *bool) ([]*model.Device, int64, error) {
	deviceUserRepo := NewDeviceUserRepository()
	return deviceUserRepo.GetUserDevices(uid, page, pageSize, devType, devStatus, permissionLevel, isActive)
}
