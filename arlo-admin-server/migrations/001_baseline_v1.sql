-- ============================================================
-- Arlo Admin 数据库基线 v1
-- 整合原 001～026 迭代后的最终态（结构从运行库导出，种子已清洗）
--
-- 全新安装：
--   mysql --default-character-set=utf8mb4 -u用户 -p < migrations/001_baseline_v1.sql
--
-- 后续变更：新增 002_*.sql 起增量补丁；勿改本文件既有语义。
-- 旧迭代脚本：archive/pre_v1/
-- ============================================================

CREATE DATABASE IF NOT EXISTS `arlo_admin`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE `arlo_admin`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `sys_config`;
CREATE TABLE `sys_config` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '配置名称',
  `key` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '配置键',
  `value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '配置值',
  `type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '类型: 1=文本, 2=JSON, 3=开关, 4=数字',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';
DROP TABLE IF EXISTS `sys_dept`;
CREATE TABLE `sys_dept` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `parent_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '上级部门ID，0为根',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '部门名称',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `leader` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '负责人',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '联系电话',
  `email` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '邮箱',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';
DROP TABLE IF EXISTS `sys_dict_data`;
CREATE TABLE `sys_dict_data` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `dict_type_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '字典类型ID',
  `label` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字典标签',
  `value` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字典值',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `is_default` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否默认: 0=否, 1=是',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_dict_type_id` (`dict_type_id`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典数据表';
DROP TABLE IF EXISTS `sys_dict_type`;
CREATE TABLE `sys_dict_type` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字典名称',
  `code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字典编码',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典类型表';
DROP TABLE IF EXISTS `sys_file`;
CREATE TABLE `sys_file` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `access_key` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '访问钥（不可猜）',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '原始文件名',
  `path` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '存储路径',
  `size` bigint(20) NOT NULL DEFAULT '0' COMMENT '文件大小(字节)',
  `mime_type` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `category` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'other' COMMENT '文件分类: image/video/audio/document/other',
  `is_public` tinyint(4) NOT NULL DEFAULT '1' COMMENT '是否公开可匿名访问: 0否 1是（默认公开）',
  `md5` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'MD5校验值',
  `uploader_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '上传者ID',
  `uploader` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '上传者',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_access_key` (`access_key`),
  KEY `idx_uploader_id` (`uploader_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_md5` (`md5`),
  KEY `idx_is_public` (`is_public`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件上传记录表';
DROP TABLE IF EXISTS `sys_job`;
CREATE TABLE `sys_job` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL COMMENT '任务名称',
  `handler` varchar(64) NOT NULL COMMENT '处理器编码（代码注册）',
  `cron` varchar(64) NOT NULL COMMENT 'Cron：分 时 日 月 周',
  `params` varchar(512) NOT NULL DEFAULT '' COMMENT '参数 JSON',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '0暂停 1启用',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `last_run_at` datetime DEFAULT NULL COMMENT '上次执行时间',
  `last_status` tinyint(4) DEFAULT NULL COMMENT '上次状态 0失败 1成功',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_handler` (`handler`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时任务';
DROP TABLE IF EXISTS `sys_job_log`;
CREATE TABLE `sys_job_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `job_id` bigint(20) unsigned NOT NULL COMMENT '任务ID',
  `job_name` varchar(64) NOT NULL DEFAULT '' COMMENT '任务名称快照',
  `handler` varchar(64) NOT NULL DEFAULT '' COMMENT '处理器',
  `trigger_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0调度 1手动',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '0失败 1成功',
  `result` text COMMENT '结果摘要',
  `error_msg` varchar(1000) NOT NULL DEFAULT '' COMMENT '错误信息',
  `duration_ms` int(11) NOT NULL DEFAULT '0' COMMENT '耗时毫秒',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时任务执行日志';
DROP TABLE IF EXISTS `sys_login_log`;
CREATE TABLE `sys_login_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '登录IP',
  `location` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '登录地点',
  `browser` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '浏览器',
  `os` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作系统',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=失败, 1=成功',
  `msg` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '提示消息',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
  PRIMARY KEY (`id`),
  KEY `idx_username` (`username`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=101 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';
DROP TABLE IF EXISTS `sys_member`;
CREATE TABLE `sys_member` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '手机号',
  `password` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '密码(bcrypt)，可选，验证码登录时为空',
  `nickname` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '头像URL',
  `gender` tinyint(4) NOT NULL DEFAULT '0' COMMENT '性别: 0=未知, 1=男, 2=女',
  `openid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信公众号openid',
  `unionid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '微信unionid',
  `mp_openid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '小程序openid',
  `source` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'h5' COMMENT '注册来源: mini(小程序)/oa(公众号)/h5',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=正常',
  `last_login` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`),
  UNIQUE KEY `uk_openid` (`openid`),
  KEY `idx_unionid` (`unionid`),
  KEY `idx_mp_openid` (`mp_openid`),
  KEY `idx_source` (`source`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='C端会员用户表(sys_member)';
DROP TABLE IF EXISTS `sys_menu`;
CREATE TABLE `sys_menu` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `parent_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '上级菜单ID，0为根',
  `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '菜单名称',
  `type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '类型: 1=目录, 2=菜单, 3=按钮',
  `path` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '路由地址',
  `component` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '组件路径',
  `icon` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '菜单图标',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `permission` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '权限标识 (如 sys:user:add)',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `visible` tinyint(4) NOT NULL DEFAULT '1' COMMENT '是否可见: 0=隐藏, 1=显示',
  `keep_alive` tinyint(4) NOT NULL DEFAULT '1' COMMENT '是否缓存: 0=否, 1=是',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=101 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单权限表';
DROP TABLE IF EXISTS `sys_message`;
CREATE TABLE `sys_message` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `title` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '消息标题',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '消息内容',
  `type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '类型: 1=系统消息, 2=通知, 3=私信',
  `sender_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '发送者ID (0=系统)',
  `sender` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '发送者',
  `receiver_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '接收者ID (0=全部)',
  `is_read` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否已读: 0=未读, 1=已读',
  `read_at` datetime DEFAULT NULL COMMENT '阅读时间',
  `sender_deleted` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'å‘é€æ–¹å·²åˆ : 0=å¦ 1=æ˜¯',
  `receiver_deleted` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'æ”¶ä»¶æ–¹å·²åˆ : 0=å¦ 1=æ˜¯',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_receiver_id` (`receiver_id`),
  KEY `idx_is_read` (`is_read`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';
DROP TABLE IF EXISTS `sys_message_hide`;
CREATE TABLE `sys_message_hide` (
  `message_id` bigint(20) unsigned NOT NULL COMMENT '消息ID（一般为广播）',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '删除时间',
  PRIMARY KEY (`message_id`,`user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='站内信个人隐藏（广播收件删除）';
DROP TABLE IF EXISTS `sys_message_read`;
CREATE TABLE `sys_message_read` (
  `message_id` bigint(20) unsigned NOT NULL COMMENT '消息ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `read_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '已读时间',
  PRIMARY KEY (`message_id`,`user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='站内信已读（广播）';
DROP TABLE IF EXISTS `sys_notice`;
CREATE TABLE `sys_notice` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `title` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标题',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '内容',
  `type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '类型: 1=通知, 2=公告',
  `level` tinyint(4) NOT NULL DEFAULT '1' COMMENT '级别: 1=普通, 2=重要, 3=紧急',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态: 0=草稿, 1=已发布, 2=已撤回',
  `publisher_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '发布人ID',
  `publisher` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '发布人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知公告表';
DROP TABLE IF EXISTS `sys_operation_log`;
CREATE TABLE `sys_operation_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `username` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `module` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作模块',
  `action` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作类型',
  `method` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求方法',
  `url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求URL',
  `ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作IP',
  `params` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '请求参数',
  `result` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '返回结果',
  `cost_time` int(11) NOT NULL DEFAULT '0' COMMENT '耗时(毫秒)',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=失败, 1=成功',
  `error_msg` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  `user_agent` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户代理',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_module` (`module`)
) ENGINE=InnoDB AUTO_INCREMENT=9320 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
DROP TABLE IF EXISTS `sys_post`;
CREATE TABLE `sys_post` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '岗位编码',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '岗位名称',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位表';
DROP TABLE IF EXISTS `sys_role`;
CREATE TABLE `sys_role` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '角色名称',
  `code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '角色编码',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `data_scope` tinyint(4) NOT NULL DEFAULT '1' COMMENT '数据范围: 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';
DROP TABLE IF EXISTS `sys_role_dept`;
CREATE TABLE `sys_role_dept` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` bigint(20) unsigned NOT NULL COMMENT '角色ID',
  `dept_id` bigint(20) unsigned NOT NULL COMMENT '部门ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_dept` (`role_id`,`dept_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=31 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-部门数据权限关联表';
DROP TABLE IF EXISTS `sys_role_menu`;
CREATE TABLE `sys_role_menu` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` bigint(20) unsigned NOT NULL COMMENT '角色ID',
  `menu_id` bigint(20) unsigned NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_menu` (`role_id`,`menu_id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_menu_id` (`menu_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1329 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';
DROP TABLE IF EXISTS `sys_user`;
CREATE TABLE `sys_user` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '密码 (bcrypt)',
  `nickname` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '头像URL',
  `email` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '邮箱',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '手机号',
  `gender` tinyint(4) NOT NULL DEFAULT '0' COMMENT '性别: 0=未知, 1=男, 2=女',
  `dept_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '部门ID',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `last_login` datetime DEFAULT NULL COMMENT '最后登录时间',
  `pwd_updated_at` datetime DEFAULT NULL COMMENT '密码最后更新时间',
  `must_change_pwd` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否强制改密 0否 1是',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_dept_id` (`dept_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
DROP TABLE IF EXISTS `sys_user_post`;
CREATE TABLE `sys_user_post` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `post_id` bigint(20) unsigned NOT NULL COMMENT '岗位ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_post` (`user_id`,`post_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_post_id` (`post_id`)
) ENGINE=InnoDB AUTO_INCREMENT=94 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户岗位关联表';
DROP TABLE IF EXISTS `sys_user_role`;
CREATE TABLE `sys_user_role` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `role_id` bigint(20) unsigned NOT NULL COMMENT '角色ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_role` (`user_id`,`role_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- ---------- 种子数据 ----------

 SET NAMES utf8mb4 ;

INSERT INTO `sys_dept` (`id`, `parent_id`, `name`, `sort`, `leader`, `phone`, `email`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,0,'集团总部',0,'管理员','13911933370','arlo.zhang2026@gmail.com',1,'2026-07-09 14:18:40','2026-08-05 13:48:14',NULL),(2,7,'技术部',0,'','','',1,'2026-07-09 14:18:40','2026-08-05 13:49:02',NULL),(3,7,'市场部',1,'','','',1,'2026-07-09 14:18:40','2026-08-05 13:49:21',NULL),(4,7,'人事部',2,'','','',1,'2026-07-09 14:18:40','2026-08-05 13:49:32',NULL),(5,7,'财务部',3,'','','',1,'2026-07-09 14:18:40','2026-08-05 13:49:39',NULL);

 SET NAMES utf8mb4 ;

INSERT INTO `sys_post` (`id`, `code`, `name`, `sort`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'CEO','董事长',0,1,'大Boss','2026-07-09 14:18:40','2026-07-21 15:05:42',NULL),(2,'CTO','技术总监',1,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(3,'MANAGER','部门经理',2,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(4,'DEVELOPER','开发工程师',3,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(5,'TESTER','测试工程师',4,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL);

 SET NAMES utf8mb4 ;

INSERT INTO `sys_user` (`id`, `username`, `password`, `nickname`, `avatar`, `email`, `phone`, `gender`, `dept_id`, `status`, `remark`, `last_login`, `pwd_updated_at`, `must_change_pwd`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'admin','$2b$12$eg61iIX2U0Ny5MOvINXXk.jbxKx3GZQz2Mg6azDo3Nc7MGR/syuRS','超级管理员','d24ddd8a8f1811f183de3e4188e7eff7','admin@arlo.com','13800000000',2,1,1,'系统管理员','2026-08-05 19:07:53','2026-08-03 11:36:51',0,'2026-07-09 14:18:40','2026-08-05 19:07:53',NULL);

 SET NAMES utf8mb4 ;

INSERT INTO `sys_role` (`id`, `name`, `code`, `sort`, `status`, `remark`, `data_scope`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'超级管理员','super_admin',0,1,'拥有所有权限',1,'2026-07-09 14:18:40','2026-07-31 18:19:43',NULL),(2,'普通用户','user',1,1,'普通用户权限',5,'2026-07-09 14:18:40','2026-08-05 16:17:05',NULL);

 SET NAMES utf8mb4 ;

INSERT INTO `sys_user_role` (`id`, `user_id`, `role_id`) VALUES (34,1,1);

INSERT INTO `sys_user_post` (`id`, `user_id`, `post_id`) VALUES (88,1,1),(89,1,2),(90,1,3),(91,1,4),(92,1,5);

 SET NAMES utf8mb4 ;

INSERT INTO `sys_dict_type` (`id`, `name`, `code`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'用户状态','sys_user_status',1,'用户状态列表','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(2,'性别','sys_gender',1,'性别列表','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(3,'通知类型','sys_notice_type',1,'通知类型列表','2026-07-09 14:18:40','2026-07-15 16:53:00',NULL),(4,'通知级别','sys_notice_level',1,'通知级别列表','2026-07-09 14:18:40','2026-07-21 15:05:46',NULL),(5,'消息类型','sys_message_type',1,'站内信类型','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(6,'会员来源','sys_member_source',1,'会员注册来源','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(7,'公告状态','sys_notice_status',1,'通知公告状态','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(8,'数据范围','sys_data_scope',1,'角色数据权限范围','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL);

INSERT INTO `sys_dict_data` (`id`, `dict_type_id`, `label`, `value`, `sort`, `is_default`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,1,'启用','1',0,1,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(2,1,'禁用','0',1,0,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(3,2,'男','1',0,0,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(4,2,'女','2',1,0,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(5,2,'未知','0',2,1,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(6,3,'通知','1',0,1,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(7,3,'公告','2',1,0,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(8,4,'普通','1',0,1,1,'','2026-07-09 14:18:40','2026-08-04 09:36:21',NULL),(9,4,'重要','2',1,0,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(10,4,'紧急','3',2,0,1,'','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(11,5,'系统消息','1',0,1,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(12,5,'通知','2',1,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(13,5,'私信','3',2,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(14,6,'H5','h5',0,1,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(15,6,'小程序','mini',1,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(16,6,'公众号','oa',2,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(17,7,'草稿','0',0,1,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(18,7,'已发布','1',1,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(19,7,'已撤回','2',2,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(20,8,'全部数据','1',0,1,1,'','2026-08-04 10:20:43','2026-08-04 10:53:04',NULL),(21,8,'自定义','2',1,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(22,8,'本部门及以下','3',2,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(23,8,'本部门','4',3,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL),(24,8,'仅本人','5',4,0,1,'','2026-08-04 10:20:43','2026-08-04 10:20:43',NULL);

INSERT INTO `sys_config` (`id`, `name`, `key`, `value`, `type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'系统名称','sys.name','Arlo Admin',1,'系统显示名称','2026-07-09 14:18:40','2026-08-05 15:40:38',NULL),(2,'系统版本','sys.version','1.0.0.release',1,'当前系统版本','2026-07-09 14:18:40','2026-08-05 10:27:59',NULL),(3,'验证码开关','sys.captcha','true',3,'登录验证码: true=开启, false=关闭','2026-07-09 14:18:40','2026-07-20 17:36:06',NULL),(4,'初始密码','sys.init_pwd','123456',1,'新增用户默认密码','2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(5,'系统Logo','sys.logo','d24ddd8a8f1811f183de3e4188e7eff7',4,'侧边栏/登录页 Logo，图片类型','2026-07-20 17:35:06','2026-08-05 10:22:47',NULL),(6,'登录最大失败次数','sys.login.max_retry','5',1,'连续密码错误达到次数后锁定，0 表示不限制','2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(7,'登录锁定分钟数','sys.login.lock_minutes','30',1,'账号锁定时长（分钟）','2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(8,'密码最小长度','sys.pwd.min_length','6',1,'新建/修改密码的最小长度','2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(9,'密码复杂度','sys.pwd.require_complexity','false',3,'开启后需同时包含字母和数字','2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(10,'密码有效天数','sys.pwd.expire_days','0',1,'超过天数强制改密，0 表示永不过期','2026-08-03 11:36:51','2026-08-03 11:36:51',NULL);

INSERT INTO `sys_job` (`id`, `name`, `handler`, `cron`, `params`, `status`, `remark`, `last_run_at`, `last_status`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'清理登录/操作日志','log_cleanup','0 3 * * *','{\"retainDays\":20}',1,'按保留天数清理登录日志与操作日志','2026-08-05 13:27:31',1,'2026-08-04 11:10:12','2026-08-05 13:27:31',NULL);

 SET NAMES utf8mb4 ;

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,0,'系统管理',1,'/system','','Setting',40,'',1,1,1,'2026-07-09 14:18:40','2026-07-21 15:12:24',NULL),(2,0,'日志管理',1,'/log','','Document',30,'',1,1,1,'2026-07-09 14:18:40','2026-07-21 15:25:27',NULL),(3,0,'消息中心',1,'/message','','Bell',10,'',1,1,1,'2026-07-09 14:18:40','2026-07-21 15:12:32',NULL),(10,88,'用户管理',2,'/system/user','/system/user/index','UserFilled',2,'sys:user:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:07:49',NULL),(11,10,'用户新增',3,'','','',0,'sys:user:add',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(12,10,'用户编辑',3,'','','',0,'sys:user:edit',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(13,10,'用户删除',3,'','','',0,'sys:user:delete',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(14,1,'角色管理',2,'/system/role','/system/role/index','Key',1,'sys:role:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:08:51',NULL),(15,14,'角色新增',3,'','','',0,'sys:role:add',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(16,14,'角色编辑',3,'','','',0,'sys:role:edit',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(17,14,'角色删除',3,'','','',0,'sys:role:delete',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(18,1,'菜单管理',2,'/system/menu','/system/menu/index','Menu',2,'sys:menu:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:08:56',NULL),(19,18,'菜单新增',3,'','','',0,'sys:menu:add',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(20,18,'菜单编辑',3,'','','',0,'sys:menu:edit',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(21,18,'菜单删除',3,'','','',0,'sys:menu:delete',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(22,88,'部门管理',2,'/system/dept','/system/dept/index','Cpu',3,'sys:dept:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:07:54',NULL),(23,22,'部门新增',3,'','','',0,'sys:dept:add',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(24,22,'部门编辑',3,'','','',0,'sys:dept:edit',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(25,22,'部门删除',3,'','','',0,'sys:dept:delete',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(26,88,'岗位管理',2,'/system/post','/system/post/index','Files',4,'sys:post:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:07:59',NULL),(27,26,'岗位新增',3,'','','',0,'sys:post:add',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(28,26,'岗位编辑',3,'','','',0,'sys:post:edit',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(29,26,'岗位删除',3,'','','',0,'sys:post:delete',1,1,1,'2026-07-09 14:18:40','2026-07-09 14:18:40',NULL),(30,1,'字典管理',2,'/system/dict','/system/dict/index','Box',5,'sys:dict:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:09:02',NULL),(31,1,'参数配置',2,'/system/sysconfig','/system/sysconfig/index','Operation',6,'sys:sysconfig:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:09:10',NULL),(40,2,'操作日志',2,'/log/operation','/log/operation','Notification',0,'log:operation:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:08:41',NULL),(41,2,'登录日志',2,'/log/login','/log/login','ScaleToOriginal',1,'log:login:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:08:46',NULL),(50,3,'通知公告',2,'/message/notice','/message/notice/index','ChatSquare',0,'message:notice:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:01:30',NULL),(51,3,'我的消息',2,'/message/my','/message/my/index','ChatDotSquare',1,'message:my:list',1,1,1,'2026-07-09 14:18:40','2026-08-05 16:01:47',NULL),(52,51,'发送消息',3,'','','',0,'message:my:add',1,1,1,'2026-07-15 17:11:57','2026-07-21 09:24:28',NULL),(53,51,'消息详情',3,'','','',1,'message:my:view',1,1,1,'2026-07-16 10:07:06','2026-07-21 09:24:28',NULL),(54,51,'消息编辑',3,'','','',2,'message:my:edit',1,1,1,'2026-07-16 13:45:13','2026-07-21 09:24:28',NULL),(55,51,'消息删除',3,'','','',3,'message:my:delete',1,1,1,'2026-07-16 13:45:44','2026-07-21 09:24:28',NULL),(56,31,'配置新增',3,'','','',0,'sys:sysconfig:add',1,1,1,'2026-07-16 16:53:07','2026-07-21 09:24:28',NULL),(57,31,'配置编辑',3,'','','',0,'sys:sysconfig:edit',1,1,1,'2026-07-16 16:53:42','2026-07-21 09:24:28',NULL),(58,31,'配置删除',3,'','','',0,'sys:sysconfig:delete',1,1,1,'2026-07-16 16:54:02','2026-07-21 09:24:28',NULL),(59,30,'字典新增',3,'','','',0,'sys:dict:add',1,1,1,'2026-07-16 16:54:35','2026-07-16 16:54:35',NULL),(60,30,'字典编辑',3,'','','',0,'sys:dict:edit',1,1,1,'2026-07-16 16:54:50','2026-07-16 16:54:50',NULL),(61,30,'字典删除',3,'','','',0,'sys:dict:delete',1,1,1,'2026-07-16 16:55:04','2026-07-16 16:55:04',NULL),(62,30,'字典配置',3,'','','',0,'sys:dict:config',1,1,1,'2026-07-16 16:55:33','2026-07-16 16:55:33',NULL),(63,40,'日志详情',3,'','','',0,'log:operation:view',1,1,1,'2026-07-16 16:59:37','2026-07-16 17:03:57',NULL),(64,50,'公告新增',3,'','','',0,'message:notice:add',1,1,1,'2026-07-16 17:00:05','2026-07-16 17:03:27',NULL),(65,50,'公告编辑',3,'','','',0,'message:notice:edit',1,1,1,'2026-07-16 17:00:22','2026-07-16 17:03:34',NULL),(66,50,'公告查看',3,'','','',0,'message:notice:view',1,1,1,'2026-07-16 17:00:37','2026-07-16 17:03:38',NULL),(67,50,'公告删除',3,'','','',0,'message:notice:delete',1,1,1,'2026-07-16 17:00:58','2026-07-16 17:03:43',NULL),(68,51,'消息新增',3,'','','',0,'message:my:add',1,1,1,'2026-07-16 17:01:39','2026-07-16 17:01:51',NULL),(69,51,'消息编辑',3,'','','',0,'message:my:edit',1,1,1,'2026-07-16 17:02:08','2026-07-16 17:02:18',NULL),(70,88,'会员管理',2,'/system/member','/system/member/index','User',0,'sys:member:list',1,1,1,'2026-07-16 17:02:32','2026-08-05 16:07:43',NULL),(71,70,'会员编辑',3,'','','',0,'sys:member:edit',1,1,1,'2026-07-16 17:02:59','2026-07-20 16:43:10',NULL),(72,70,'会员删除',3,'','','',0,'sys:member:delete',1,1,1,'2026-07-20 16:38:36','2026-07-20 16:38:36',NULL),(80,1,'文件管理',2,'/system/file','/system/file/index','UploadFilled',7,'sys:file:list',1,1,1,'2026-07-21 09:24:28','2026-08-05 16:09:17',NULL),(81,80,'文件上传',3,'','','',0,'sys:file:upload',1,1,1,'2026-07-21 09:24:28','2026-07-21 09:24:28',NULL),(82,80,'文件删除',3,'','','',1,'sys:file:delete',1,1,1,'2026-07-21 09:24:28','2026-07-21 09:24:28',NULL),(83,0,'系统监控',1,'/monitor','','Monitor',20,'',1,1,1,'2026-07-21 14:20:53','2026-07-21 15:12:46',NULL),(84,83,'在线用户',2,'/monitor/online','/monitor/online','User',0,'monitor:online:list',1,1,1,'2026-07-21 14:20:53','2026-08-05 16:08:21',NULL),(85,84,'强制下线',3,'','','',0,'monitor:online:kick',1,1,1,'2026-07-21 14:20:53','2026-07-21 14:20:53',NULL),(88,0,'用户中心',1,'/user','','UserFilled',0,'',1,1,1,'2026-07-21 15:09:12','2026-07-21 15:09:12',NULL),(89,83,'服务监控',2,'/monitor/server','/monitor/server','Odometer',1,'monitor:server:list',1,1,1,'2026-07-21 15:48:58','2026-08-05 16:08:26',NULL),(90,10,'用户导出',3,'','','',10,'sys:user:export',1,1,1,'2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(91,10,'用户导入',3,'','','',11,'sys:user:import',1,1,1,'2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(92,40,'导出日志',3,'','','',11,'log:operation:export',1,1,1,'2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(93,41,'导出日志',3,'','','',11,'log:login:export',1,1,1,'2026-08-03 11:36:51','2026-08-03 11:36:51',NULL),(94,10,'解锁用户',3,'','','',12,'sys:user:unlock',1,1,1,'2026-08-03 11:37:04','2026-08-03 11:37:04',NULL),(95,70,'会员新增',3,'','','',1,'sys:member:add',1,1,1,'2026-08-03 14:43:42','2026-08-03 14:43:42',NULL),(96,83,'定时任务',2,'/monitor/job','/monitor/job','Timer',2,'monitor:job:list',1,1,1,'2026-08-04 11:10:12','2026-08-05 15:59:38',NULL),(97,96,'任务编辑',3,'','','',1,'monitor:job:edit',1,1,1,'2026-08-04 11:10:12','2026-08-04 11:10:12',NULL),(98,96,'任务启停',3,'','','',2,'monitor:job:status',1,1,1,'2026-08-04 11:10:12','2026-08-04 11:10:12',NULL),(99,96,'立即执行',3,'','','',3,'monitor:job:run',1,1,1,'2026-08-04 11:10:12','2026-08-04 11:10:12',NULL),(100,96,'执行日志',3,'','','',4,'monitor:job:log',1,1,1,'2026-08-04 11:10:12','2026-08-04 11:10:12',NULL);

 SET NAMES utf8mb4 ;

INSERT INTO `sys_role_menu` (`id`, `role_id`, `menu_id`) VALUES (1132,1,88),(1133,1,70),(1134,1,71),(1135,1,72),(1136,1,10),(1137,1,11),(1138,1,12),(1139,1,13),(1140,1,22),(1141,1,23),(1142,1,24),(1143,1,25),(1144,1,26),(1145,1,27),(1146,1,28),(1147,1,29),(1148,1,3),(1149,1,50),(1150,1,64),(1151,1,65),(1152,1,66),(1153,1,67),(1154,1,51),(1155,1,52),(1156,1,68),(1157,1,69),(1158,1,53),(1159,1,54),(1160,1,55),(1161,1,83),(1162,1,84),(1163,1,85),(1164,1,89),(1165,1,2),(1166,1,40),(1167,1,63),(1169,1,41),(1171,1,1),(1172,1,14),(1173,1,15),(1174,1,16),(1175,1,17),(1176,1,18),(1177,1,19),(1178,1,20),(1179,1,21),(1180,1,30),(1181,1,59),(1182,1,60),(1183,1,61),(1184,1,62),(1185,1,31),(1186,1,56),(1187,1,57),(1188,1,58),(1189,1,80),(1190,1,81),(1191,1,82),(1192,1,90),(1193,1,91),(1194,1,92),(1195,1,93),(1196,1,94),(1234,1,95),(1235,1,96),(1236,1,97),(1237,1,98),(1238,1,99),(1239,1,100);

SET FOREIGN_KEY_CHECKS = 1;

-- 基线结束 (v1)
