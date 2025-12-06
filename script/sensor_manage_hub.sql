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
    `bound_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    `is_active` boolean NOT NULL DEFAULT TRUE COMMENT '是否激活',
    PRIMARY KEY (`uid`, `dev_id`),
    KEY `idx_dev_id` (`dev_id`),
    KEY `idx_permission_level` (`permission_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户设备绑定表';

-- ==============================================
-- Device表 (传感器设备表)
-- ==============================================
DROP TABLE IF EXISTS `device`;
CREATE TABLE `device` (
    `dev_id` bigint NOT NULL AUTO_INCREMENT COMMENT '设备唯一标识',
    `dev_name` varchar(100) NOT NULL COMMENT '设备名字',
    `dev_type` varchar(30) NOT NULL COMMENT '设备类型',
    `dev_power` int unsigned DEFAULT NULL COMMENT '设备电量',
    `dev_status` tinyint(3) unsigned NOT NULL DEFAULT 1 COMMENT '设备状态: 0在线/1离线/2异常',
    `firmware_version` varchar(20) DEFAULT NULL COMMENT '固件版本',
    `device_model` varchar(50) DEFAULT NULL COMMENT '设备型号',
    `sampling_frequency` int unsigned DEFAULT NULL COMMENT '采样频率',
    `data_upload_interval` int unsigned DEFAULT NULL COMMENT '数据上报间隔',
    `offline_threshold` int unsigned DEFAULT NULL COMMENT '离线判断阈值',
    `extended_config` json DEFAULT NULL COMMENT '扩展配置',
    `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`dev_id`),
    KEY `idx_dev_type` (`dev_type`),
    KEY `idx_dev_status` (`dev_status`),
    KEY `idx_create_at` (`create_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='传感器设备表';

-- ==============================================
-- AlertEvent表 (告警事件表)
-- ==============================================
DROP TABLE IF EXISTS `alert_event`;
CREATE TABLE `alert_event` (
    `alert_id` bigint NOT NULL AUTO_INCREMENT COMMENT '告警ID',
    `data_id` bigint DEFAULT NULL COMMENT '数据ID',
    `alert_type` enum('设备异常','数据异常','阈值超限','通信故障','电池低电量','维护提醒') NOT NULL COMMENT '告警类型',
    `alert_message` varchar(255) NOT NULL COMMENT '告警消息',
    `alert_status` enum('活跃','已解决','忽略') NOT NULL DEFAULT '活跃' COMMENT '告警状态',
    `triggered_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '触发时间',
    `resolved_at` datetime DEFAULT NULL COMMENT '解决时间',
    PRIMARY KEY (`alert_id`),
    KEY `idx_data_id` (`data_id`),
    KEY `idx_alert_type` (`alert_type`),
    KEY `idx_alert_status` (`alert_status`),
    KEY `idx_triggered_at` (`triggered_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警事件表';


-- ==============================================
-- MetaData表 (传感器元数据表)
-- ==============================================
DROP TABLE IF EXISTS `metadata`;
CREATE TABLE `metadata` (
    `data_id` bigint NOT NULL AUTO_INCREMENT COMMENT '元数据ID',
    `dev_id` bigint NOT NULL COMMENT '设备ID',
    `uid` bigint NOT NULL COMMENT '用户ID',
    `metadata_type` varchar(30) NOT NULL COMMENT '元数据类型',
    `storage_route` varchar(255) DEFAULT NULL COMMENT '存储路由',
    `data_credibility` decimal(3,2) DEFAULT NULL COMMENT '数据可信度',
    `timestamp` datetime NOT NULL COMMENT '时间戳',
    `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`data_id`),
    KEY `idx_dev_id` (`dev_id`),
    KEY `idx_metadata_type` (`metadata_type`),
    KEY `idx_timestamp` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='传感器元数据表';


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
