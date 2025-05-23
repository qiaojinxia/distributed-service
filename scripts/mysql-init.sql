-- MySQL 初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS distributed_service CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE distributed_service;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    username varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
    email varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    password varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    status int DEFAULT '1',
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL,
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY idx_users_username (username),
    KEY idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入测试数据 (密码为 bcrypt 哈希: password123)
INSERT INTO users (username, email, password, status, created_at, updated_at) VALUES
('admin', 'admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 1, NOW(), NOW()),
('testuser', 'test@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 1, NOW(), NOW());

-- 授权给用户
GRANT ALL PRIVILEGES ON distributed_service.* TO 'admin'@'%';
FLUSH PRIVILEGES; 