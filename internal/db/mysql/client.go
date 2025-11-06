package mysql

import (
	"backend/config"
	"backend/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MysqlCli *MysqlClient

type MysqlClient struct {
	Client *sql.DB
}

// GetMysqlClient 获取Mysql客户端
// 使用 config.MysqlConfig 作为参数类型
func GetMysqlClient(cfg config.MysqlConfig) (*MysqlClient, error) {
	if MysqlCli != nil {
		return MysqlCli, nil
	}
	db, err := InitMysqlClient(cfg)
	if err != nil {
		return nil, err
	}
	MysqlCli = &MysqlClient{Client: db}
	return MysqlCli, nil
}

// InitMysqlClient 初始化Mysql客户端
func InitMysqlClient(cfg config.MysqlConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connect: %v", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Minute)

	// 测试连接是否成功（带超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping mysql server: %v", err)
	}
	logger.L().Info("Mysql客户端初始化成功")
	return db, nil
}

// Close 关闭Mysql客户端
func (db *MysqlClient) Close() error {
	if db != nil {
		return db.Client.Close()
	}
	return nil
}
