-- 初始化数据库脚本

-- 创建数据库(如果不存在)
-- CREATE DATABASE IF NOT EXISTS teable;

-- 使用数据库
-- \c teable;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建基础表(这里只创建用户相关表作为示例)
-- 实际项目中应该使用migration工具

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(30) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255),
    salt VARCHAR(255),
    phone VARCHAR(50) UNIQUE,
    avatar VARCHAR(500),
    is_system BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    is_trial_used BOOLEAN DEFAULT FALSE,
    notify_meta TEXT,
    last_sign_time TIMESTAMP,
    deactivated_time TIMESTAMP,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_time TIMESTAMP,
    last_modified_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    permanent_deleted_time TIMESTAMP,
    ref_meta TEXT
);

-- 账户表(第三方登录)
CREATE TABLE IF NOT EXISTS account (
    id VARCHAR(30) PRIMARY KEY,
    user_id VARCHAR(30) NOT NULL,
    type VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(provider, provider_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_created_time ON users(created_time);
CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin);
CREATE INDEX IF NOT EXISTS idx_users_deactivated_time ON users(deactivated_time);

CREATE INDEX IF NOT EXISTS idx_account_user_id ON account(user_id);
CREATE INDEX IF NOT EXISTS idx_account_provider ON account(provider, provider_id);

-- 空间表
CREATE TABLE IF NOT EXISTS space (
    id VARCHAR(30) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    created_by VARCHAR(30) NOT NULL,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_time TIMESTAMP,
    last_modified_time TIMESTAMP
);

-- 空间协作者
CREATE TABLE IF NOT EXISTS space_collaborator (
    id VARCHAR(30) PRIMARY KEY,
    space_id VARCHAR(30) NOT NULL,
    user_id VARCHAR(30) NOT NULL,
    role VARCHAR(20) NOT NULL,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_space_collab_space FOREIGN KEY (space_id) REFERENCES space(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_space_collab_space ON space_collaborator(space_id);
CREATE INDEX IF NOT EXISTS idx_space_collab_user ON space_collaborator(user_id);

-- Base（逻辑库）
CREATE TABLE IF NOT EXISTS base (
    id VARCHAR(30) PRIMARY KEY,
    space_id VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    created_by VARCHAR(30) NOT NULL,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_time TIMESTAMP,
    last_modified_time TIMESTAMP,
    CONSTRAINT fk_base_space FOREIGN KEY (space_id) REFERENCES space(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_base_space ON base(space_id);

-- 表元数据
CREATE TABLE IF NOT EXISTS table_meta (
    id VARCHAR(30) PRIMARY KEY,
    base_id VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    db_table_name VARCHAR(255) NOT NULL UNIQUE,
    "order" INT DEFAULT 0,
    version INT DEFAULT 1,
    created_by VARCHAR(30) NOT NULL,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_time TIMESTAMP,
    last_modified_time TIMESTAMP,
    CONSTRAINT fk_table_meta_base FOREIGN KEY (base_id) REFERENCES base(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_table_meta_base ON table_meta(base_id);
CREATE INDEX IF NOT EXISTS idx_table_meta_dbname ON table_meta(db_table_name);

-- 字段定义（简化）
CREATE TABLE IF NOT EXISTS field (
    id VARCHAR(30) PRIMARY KEY,
    table_id VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    cell_value_type VARCHAR(50) NOT NULL,
    is_multiple_cell_value BOOLEAN,
    db_field_type VARCHAR(50) NOT NULL,
    db_field_name VARCHAR(255) NOT NULL,
    not_null BOOLEAN,
    "unique" BOOLEAN,
    is_primary BOOLEAN,
    is_computed BOOLEAN,
    is_lookup BOOLEAN,
    "order" DECIMAL(10,2),
    version INT DEFAULT 1,
    created_by VARCHAR(30) NOT NULL,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_modified_time TIMESTAMP,
    CONSTRAINT fk_field_table FOREIGN KEY (table_id) REFERENCES table_meta(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_field_table ON field(table_id);

-- 插入默认系统用户
INSERT INTO users (
    id, 
    name, 
    email, 
    password,
    is_system, 
    is_admin,
    created_time
) VALUES (
    'usr_system_admin_000001',
    'System Admin',
    'admin@teable.ai',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: "admin123"
    true,
    true,
    CURRENT_TIMESTAMP
) ON CONFLICT (id) DO NOTHING;

-- 插入测试用户
INSERT INTO users (
    id,
    name,
    email,
    password,
    is_system,
    is_admin,
    created_time
) VALUES (
    'usr_test_user_0000001',
    'Test User',
    'test@teable.ai', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: "test123"
    false,
    false,
    CURRENT_TIMESTAMP
) ON CONFLICT (id) DO NOTHING;