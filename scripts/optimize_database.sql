-- 数据库优化脚本
-- 为 Teable Go Backend 添加性能优化索引

-- ========================================
-- 1. 用户表索引
-- ========================================

-- 邮箱唯一索引（加速登录查询）
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email 
ON users(email) 
WHERE deleted_time IS NULL;

-- 用户状态索引
CREATE INDEX IF NOT EXISTS idx_users_status 
ON users(deleted_time, deactivated_time);

-- 用户创建时间索引（用于统计和排序）
CREATE INDEX IF NOT EXISTS idx_users_created 
ON users(created_time DESC);

-- ========================================
-- 2. 空间表索引
-- ========================================

-- 空间所有者索引
CREATE INDEX IF NOT EXISTS idx_spaces_owner 
ON spaces(owner_id) 
WHERE deleted_time IS NULL;

-- 空间创建时间索引
CREATE INDEX IF NOT EXISTS idx_spaces_created 
ON spaces(created_time DESC) 
WHERE deleted_time IS NULL;

-- ========================================
-- 3. 空间协作者表索引
-- ========================================

-- 复合索引：空间ID + 用户ID（唯一约束）
CREATE UNIQUE INDEX IF NOT EXISTS idx_space_collaborators_unique 
ON space_collaborators(space_id, user_id) 
WHERE deleted_time IS NULL;

-- 用户ID索引（查询用户的所有空间）
CREATE INDEX IF NOT EXISTS idx_space_collaborators_user 
ON space_collaborators(user_id) 
WHERE deleted_time IS NULL;

-- ========================================
-- 4. 基础表索引
-- ========================================

-- 空间ID索引
CREATE INDEX IF NOT EXISTS idx_bases_space 
ON bases(space_id) 
WHERE deleted_time IS NULL;

-- 创建时间索引
CREATE INDEX IF NOT EXISTS idx_bases_created 
ON bases(created_time DESC) 
WHERE deleted_time IS NULL;

-- ========================================
-- 5. 数据表索引
-- ========================================

-- 基础表ID索引
CREATE INDEX IF NOT EXISTS idx_tables_base 
ON tables(base_id) 
WHERE deleted_time IS NULL;

-- 表名索引（用于搜索）
CREATE INDEX IF NOT EXISTS idx_tables_name 
ON tables(name) 
WHERE deleted_time IS NULL;

-- ========================================
-- 6. 字段表索引
-- ========================================

-- 表ID + 显示顺序复合索引
CREATE INDEX IF NOT EXISTS idx_fields_table_order 
ON fields(table_id, display_order) 
WHERE deleted_time IS NULL;

-- 字段类型索引
CREATE INDEX IF NOT EXISTS idx_fields_type 
ON fields(type) 
WHERE deleted_time IS NULL;

-- ========================================
-- 7. 记录表索引（最重要的性能优化）
-- ========================================

-- 表ID + 创建时间复合索引（常用查询）
CREATE INDEX IF NOT EXISTS idx_records_table_created 
ON records(table_id, created_time DESC) 
WHERE deleted_time IS NULL;

-- 表ID + 修改时间复合索引（最近修改查询）
CREATE INDEX IF NOT EXISTS idx_records_table_modified 
ON records(table_id, last_modified_time DESC) 
WHERE deleted_time IS NULL;

-- JSONB GIN 索引（加速 JSON 查询）
CREATE INDEX IF NOT EXISTS idx_records_data_gin 
ON records USING gin(data);

-- 创建者索引
CREATE INDEX IF NOT EXISTS idx_records_created_by 
ON records(created_by) 
WHERE deleted_time IS NULL;

-- ========================================
-- 8. 视图表索引
-- ========================================

-- 表ID索引
CREATE INDEX IF NOT EXISTS idx_views_table 
ON views(table_id) 
WHERE deleted_time IS NULL;

-- 视图类型索引
CREATE INDEX IF NOT EXISTS idx_views_type 
ON views(type) 
WHERE deleted_time IS NULL;

-- ========================================
-- 9. 权限表索引
-- ========================================

-- 用户权限查询索引
CREATE INDEX IF NOT EXISTS idx_permissions_user_resource 
ON permissions(user_id, resource_type, resource_id) 
WHERE deleted_time IS NULL;

-- 资源权限查询索引
CREATE INDEX IF NOT EXISTS idx_permissions_resource 
ON permissions(resource_type, resource_id) 
WHERE deleted_time IS NULL;

-- 角色索引
CREATE INDEX IF NOT EXISTS idx_permissions_role 
ON permissions(role) 
WHERE deleted_time IS NULL;

-- ========================================
-- 10. 附件表索引
-- ========================================

-- 记录ID索引
CREATE INDEX IF NOT EXISTS idx_attachments_record 
ON attachments(record_id) 
WHERE deleted_time IS NULL;

-- 创建者索引
CREATE INDEX IF NOT EXISTS idx_attachments_created_by 
ON attachments(created_by) 
WHERE deleted_time IS NULL;

-- 文件类型索引
CREATE INDEX IF NOT EXISTS idx_attachments_mime_type 
ON attachments(mime_type);

-- ========================================
-- 11. 通知表索引
-- ========================================

-- 用户ID + 已读状态 + 创建时间复合索引
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread 
ON notifications(user_id, is_read, created_time DESC) 
WHERE deleted_time IS NULL;

-- 类型索引
CREATE INDEX IF NOT EXISTS idx_notifications_type 
ON notifications(type) 
WHERE deleted_time IS NULL;

-- ========================================
-- 12. 搜索索引表优化
-- ========================================

-- 源ID索引
CREATE INDEX IF NOT EXISTS idx_search_indexes_source 
ON search_indexes(source_id, source_type) 
WHERE deleted_time IS NULL;

-- 全文搜索索引
CREATE INDEX IF NOT EXISTS idx_search_indexes_fulltext 
ON search_indexes USING gin(to_tsvector('english', title || ' ' || content));

-- 关键词索引
CREATE INDEX IF NOT EXISTS idx_search_indexes_keywords 
ON search_indexes USING gin(keywords);

-- ========================================
-- 13. 分析和统计
-- ========================================

-- 更新所有表的统计信息
ANALYZE users;
ANALYZE spaces;
ANALYZE space_collaborators;
ANALYZE bases;
ANALYZE tables;
ANALYZE fields;
ANALYZE records;
ANALYZE views;
ANALYZE permissions;
ANALYZE attachments;
ANALYZE notifications;
ANALYZE search_indexes;

-- ========================================
-- 14. 配置优化建议
-- ========================================

-- 以下是 PostgreSQL 配置优化建议（需要在 postgresql.conf 中设置）
-- shared_buffers = 256MB          # 共享缓冲区（根据可用内存调整）
-- effective_cache_size = 1GB      # 有效缓存大小
-- work_mem = 4MB                  # 工作内存
-- maintenance_work_mem = 64MB     # 维护工作内存
-- checkpoint_completion_target = 0.9
-- wal_buffers = 16MB
-- default_statistics_target = 100
-- random_page_cost = 1.1          # 如果使用 SSD
-- effective_io_concurrency = 200  # 如果使用 SSD

-- ========================================
-- 15. 查询性能视图
-- ========================================

-- 创建慢查询监控视图
CREATE OR REPLACE VIEW v_slow_queries AS
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    min_time,
    max_time,
    stddev_time,
    rows
FROM pg_stat_statements
WHERE mean_time > 100  -- 平均执行时间超过 100ms
ORDER BY mean_time DESC
LIMIT 50;

-- 创建索引使用情况视图
CREATE OR REPLACE VIEW v_index_usage AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- 创建表膨胀检查视图
CREATE OR REPLACE VIEW v_table_bloat AS
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- ========================================
-- 提示：执行完成后，建议运行 VACUUM ANALYZE
-- ========================================