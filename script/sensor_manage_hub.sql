DROP DATABASE IF EXISTS sensor_manage_hub;
CREATE DATABASE sensor_manage_hub; 

USE sensor_manage_hub;

-- ==============================================
-- User表 (用户表)
-- ==============================================
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
    `uid` bigint NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `role` enum('admin','user') NOT NULL DEFAULT 'user' COMMENT '用户角色',
    `username` varchar(50) NOT NULL COMMENT '用户姓名',
    `email` varchar(100) NOT NULL COMMENT '邮箱地址',
    `password_hash` varchar(255) DEFAULT NULL COMMENT '密码哈希',
    `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`uid`),
    UNIQUE KEY `uk_email` (`email`),
		UNIQUE KEY `uk_username` (`username`),
    KEY `idx_role` (`role`),
    KEY `idx_create_at` (`create_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ==============================================
-- User_Dev表 (用户设备绑定表)
-- ==============================================
DROP TABLE IF EXISTS `user_dev`;
CREATE TABLE `user_dev` (
    `uid` bigint NOT NULL COMMENT '用户ID',
    `dev_id` bigint NOT NULL COMMENT '设备ID',
    `permission_level` enum('r','w','rw') NOT NULL DEFAULT 'rw' COMMENT '权限级别',
    `bind_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    PRIMARY KEY (`uid`, `dev_id`),
    KEY `idx_dev_id` (`dev_id`),             -- 查询dev绑定的用户时优化
    KEY `idx_permission_level` (`permission_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户设备绑定表';

-- ==============================================
-- Device表 (传感器设备表)
-- ==============================================
DROP TABLE IF EXISTS `device`;
CREATE TABLE `device` (
    `dev_id` bigint NOT NULL AUTO_INCREMENT COMMENT '设备唯一ID',
    `dev_name` varchar(100) NOT NULL COMMENT '设备名字',
    `dev_status` smallint UNSIGNED NOT NULL DEFAULT 0 COMMENT '设备状态: 0离线/1在线/2异常',
    `dev_type` varchar(30) NOT NULL COMMENT '设备类型',
    `dev_power` int UNSIGNED DEFAULT NULL COMMENT '设备电量',
    `model` varchar(50) DEFAULT NULL COMMENT '硬件型号',
    `version` varchar(20) DEFAULT NULL COMMENT '硬件版本',
    `sampling_rate` int UNSIGNED DEFAULT NULL COMMENT '采样频率',
    `offline_threshold` int UNSIGNED DEFAULT NULL COMMENT '离线判断阈值',
    `upload_interval` int UNSIGNED DEFAULT NULL COMMENT '数据上报间隔',
    `extended_config` json DEFAULT NULL COMMENT '扩展配置',
    `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`dev_id`),
    UNIQUE KEY `uk_dev_name` (`dev_name`),
    KEY `idx_dev_status` (`dev_status`),
    KEY `idx_dev_type` (`dev_type`),
    KEY `idx_create_at` (`create_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='传感器设备表';
-- ==============================================
-- AlertEvent表 (告警事件表)
-- ==============================================
DROP TABLE IF EXISTS `alert_event`;
CREATE TABLE `alert_event` (
    `alert_id` bigint NOT NULL AUTO_INCREMENT COMMENT '告警ID',
    `data_id` bigint NOT NULL COMMENT '数据ID',
    `dev_id` bigint NOT NULL COMMENT '设备ID',
    `alert_type` enum('dev','data') NOT NULL COMMENT '告警类型',
    `alert_message` text NOT NULL COMMENT '告警消息',
    `alert_status` enum('active','resolved','ignored') NOT NULL DEFAULT 'active' COMMENT '告警状态',
    `triggered_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '触发时间',
    `resolved_at` datetime DEFAULT NULL COMMENT '解决时间',
    PRIMARY KEY (`alert_id`),
    KEY `idx_data_id` (`data_id`),
		KEY `idx_dev_id` (`dev_id`),
    KEY `idx_alert_type` (`alert_type`),
    KEY `idx_alert_status` (`alert_status`),
    KEY `idx_triggered_at` (`triggered_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警事件表';

-- ==============================================
-- MetaData表 (传感器元数据表)
-- ==============================================
DROP TABLE IF EXISTS `metadata`;
CREATE TABLE `metadata` (
    `data_id` bigint NOT NULL AUTO_INCREMENT COMMENT '数据ID',
    `dev_id` bigint NOT NULL COMMENT '设备ID',
    `data_type` enum('file_data','time_series') NOT NULL COMMENT '数据类型',
    `quality_score` varchar(10) NOT NULL COMMENT '数据质量评分0-100',
    `extra_data` json DEFAULT NULL COMMENT '其他额外数据',
    `timestamp` timestamp NOT NULL COMMENT '时间戳',
    PRIMARY KEY (`data_id`),
    KEY `idx_dev_id` (`dev_id`),
    KEY `idx_data_type` (`data_type`),
    KEY `idx_quality_score` (`quality_score`),
    KEY `idx_timestamp` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='元数据表';

-- ==============================================
-- SystemLog表 (系统日志表)
-- ==============================================
DROP TABLE IF EXISTS `system_log`;
CREATE TABLE `system_log` (
    `log_id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `type` enum('debug','info','warning','error','critical') NOT NULL DEFAULT 'info' COMMENT '日志级别',
    `level` tinyint unsigned NOT NULL COMMENT '日志级别',
    `message` text NOT NULL COMMENT '日志消息',
    `user_agent` varchar(255) DEFAULT NULL COMMENT '用户代理',
    `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`log_id`),
    KEY `idx_level` (`level`),
    KEY `idx_type` (`type`),
    KEY `idx_create_at` (`create_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统日志表';
