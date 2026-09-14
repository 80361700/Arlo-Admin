-- 审批运行时 / 菜单 / 定时 / 评论委托 / 抄送已读 / 演示单据
-- 合并原增量 002～020 的净效果（002～008 已合入 001_baseline，本文件只补基线之后部分）
-- 适用：已执行 001_baseline_v1.sql 的库，升级时导入本文件即可
-- 执行：mysql --default-character-set=utf8mb4 -u root -p arlo_admin < 002_flow_approve_runtime.sql
-- 幂等：表 IF NOT EXISTS、菜单/任务 NOT EXISTS、字段按 information_schema 判断
-- 旧零散补丁已废弃删除，勿再查找 003～020 单文件

SET NAMES utf8mb4;

-- ========== 运行时表 ==========
CREATE TABLE IF NOT EXISTS `flow_instance` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `process_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '流程定义ID',
  `process_key` varchar(64) NOT NULL DEFAULT '' COMMENT '流程标识',
  `process_name` varchar(128) NOT NULL DEFAULT '' COMMENT '流程名称快照',
  `process_version` int(11) NOT NULL DEFAULT '1' COMMENT '定义版本快照',
  `process_type` varchar(32) NOT NULL DEFAULT 'main' COMMENT '流程类型快照',
  `model_content` longtext COMMENT '模型JSON快照',
  `process_form` longtext COMMENT '表单schema快照',
  `form_data` longtext COMMENT '表单业务数据JSON',
  `current_node_key` varchar(64) NOT NULL DEFAULT '' COMMENT '当前节点key',
  `current_node_name` varchar(128) NOT NULL DEFAULT '' COMMENT '当前节点名称',
  `instance_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0审批中 1通过 2拒绝 3撤销 4终止 5超时',
  `create_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '发起人ID',
  `create_by` varchar(64) NOT NULL DEFAULT '' COMMENT '发起人姓名',
  `create_dept_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '发起人部门',
  `parent_instance_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '父实例(子流程)',
  `parent_node_key` varchar(64) NOT NULL DEFAULT '' COMMENT '父流程中子流程节点key',
  `join_gate_key` varchar(64) NOT NULL DEFAULT '' COMMENT '并行/包容汇聚门key',
  `variable` text COMMENT '运行变量JSON',
  `finish_time` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_flow_inst_process` (`process_id`),
  KEY `idx_flow_inst_create` (`create_id`),
  KEY `idx_flow_inst_state` (`instance_state`),
  KEY `idx_flow_inst_parent` (`parent_instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程实例';

CREATE TABLE IF NOT EXISTS `flow_task` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `instance_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `node_key` varchar(64) NOT NULL DEFAULT '',
  `node_name` varchar(128) NOT NULL DEFAULT '',
  `node_type` int(11) NOT NULL DEFAULT '0' COMMENT '节点类型同设计器',
  `task_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0主办 1转办 2委派 3会签 4抄送 5延时 6触发',
  `task_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0活动 1完成 2拒绝 3撤销 4终止 5跳过',
  `examine_mode` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1依次 2会签 3或签',
  `parent_task_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `from_node_key` varchar(64) NOT NULL DEFAULT '' COMMENT '来源节点',
  `gate_token` varchar(64) NOT NULL DEFAULT '' COMMENT '并行/包容分支令牌',
  `expire_time` datetime(3) DEFAULT NULL COMMENT '延时/超时截止',
  `payload` text COMMENT '节点配置摘要/延时参数等JSON',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_flow_task_inst` (`instance_id`),
  KEY `idx_flow_task_state` (`task_state`),
  KEY `idx_flow_task_node` (`instance_id`,`node_key`),
  KEY `idx_flow_task_expire` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程活动任务';

CREATE TABLE IF NOT EXISTS `flow_task_actor` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `instance_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `actor_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '处理人用户ID',
  `actor_name` varchar(64) NOT NULL DEFAULT '',
  `actor_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0用户',
  `actor_state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0待办 1同意 2拒绝 3转交 4跳过',
  `weight` int(11) NOT NULL DEFAULT '0' COMMENT '依次审批序号',
  `agent_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '实际操作人',
  `opinion` varchar(1000) NOT NULL DEFAULT '' COMMENT '意见',
  `finish_time` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_flow_actor_task` (`task_id`),
  KEY `idx_flow_actor_user` (`actor_id`,`actor_state`),
  KEY `idx_flow_actor_inst` (`instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务参与人';

CREATE TABLE IF NOT EXISTS `flow_his_task` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '原任务ID',
  `instance_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `node_key` varchar(64) NOT NULL DEFAULT '',
  `node_name` varchar(128) NOT NULL DEFAULT '',
  `node_type` int(11) NOT NULL DEFAULT '0',
  `task_type` tinyint(4) NOT NULL DEFAULT '0',
  `task_state` tinyint(4) NOT NULL DEFAULT '0',
  `examine_mode` tinyint(4) NOT NULL DEFAULT '1',
  `from_node_key` varchar(64) NOT NULL DEFAULT '',
  `gate_token` varchar(64) NOT NULL DEFAULT '',
  `payload` text,
  `created_at` datetime(3) DEFAULT NULL,
  `finish_time` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_flow_his_task_inst` (`instance_id`),
  KEY `idx_flow_his_task_tid` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='历史任务';

CREATE TABLE IF NOT EXISTS `flow_his_task_actor` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `his_task_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `task_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `instance_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `actor_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `actor_name` varchar(64) NOT NULL DEFAULT '',
  `actor_type` tinyint(4) NOT NULL DEFAULT '0',
  `actor_state` tinyint(4) NOT NULL DEFAULT '0',
  `weight` int(11) NOT NULL DEFAULT '0',
  `agent_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `opinion` varchar(1000) NOT NULL DEFAULT '',
  `finish_time` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `read_at` datetime(3) DEFAULT NULL COMMENT '抄送已读时间',
  PRIMARY KEY (`id`),
  KEY `idx_flow_his_actor_inst` (`instance_id`),
  KEY `idx_flow_his_actor_user` (`actor_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='历史任务参与人';

-- 已有库可能缺 read_at（旧 017）
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'flow_his_task_actor'
    AND COLUMN_NAME = 'read_at'
);
SET @sql := IF(
  @col_exists = 0,
  'ALTER TABLE `flow_his_task_actor` ADD COLUMN `read_at` datetime(3) DEFAULT NULL COMMENT ''抄送已读时间'' AFTER `created_at`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `flow_instance_comment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `instance_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `user_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `user_name` varchar(64) NOT NULL DEFAULT '',
  `content` varchar(500) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_flow_comment_inst` (`instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程实例评论';

CREATE TABLE IF NOT EXISTS `flow_user_delegate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '委托人',
  `to_user_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '受托人',
  `to_user_name` varchar(64) NOT NULL DEFAULT '' COMMENT '受托人姓名',
  `enabled` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1启用 0停用',
  `remark` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_flow_delegate_user` (`user_id`),
  KEY `idx_flow_delegate_to` (`to_user_id`, `enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审批委托';

CREATE TABLE IF NOT EXISTS `demo_purchase_order` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(128) NOT NULL DEFAULT '' COMMENT '名称',
  `content` varchar(512) NOT NULL DEFAULT '' COMMENT '内容',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0待审批 1审批中 2已通过 3已拒绝',
  `instance_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联流程实例',
  `process_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '绑定流程ID',
  `process_key` varchar(64) NOT NULL DEFAULT '' COMMENT '绑定流程KEY',
  `create_id` bigint unsigned NOT NULL DEFAULT 0,
  `create_by` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_demo_po_status` (`status`),
  KEY `idx_demo_po_instance` (`instance_id`),
  KEY `idx_demo_po_create` (`create_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='演示采购单（业务流程）';

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'demo_purchase_order'
    AND COLUMN_NAME = 'process_id'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `demo_purchase_order` ADD COLUMN `process_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''绑定流程ID'' AFTER `instance_id`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE `demo_purchase_order`
  MODIFY COLUMN `process_key` varchar(64) NOT NULL DEFAULT '' COMMENT '绑定流程KEY';

UPDATE `demo_purchase_order`
SET `process_key` = '', `updated_at` = NOW(3)
WHERE `instance_id` = 0 AND `process_id` = 0 AND `process_key` = 'purchaseOrder';

-- ========== 定时任务 flow_tick ==========
INSERT INTO `sys_job` (`id`, `name`, `handler`, `cron`, `params`, `status`, `remark`, `last_run_at`, `last_status`, `created_at`, `updated_at`, `deleted_at`)
SELECT 2, '流程定时推进', 'flow_tick', '* * * * *', '', 1, '延时节点到期、审批超时自动通过/拒绝、审批提醒站内信', NULL, 0, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_job` WHERE `handler` = 'flow_tick' AND `deleted_at` IS NULL);

UPDATE `sys_job`
SET `name` = '流程定时推进',
    `remark` = '延时节点到期、审批超时自动通过/拒绝、审批提醒站内信',
    `updated_at` = NOW(3)
WHERE `handler` = 'flow_tick' AND `deleted_at` IS NULL;

-- ========== 菜单：流程审批（一级目录，净布局） ==========
INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 230, 0, '流程审批', 1, '/approve', '', 'EditPen', 19, '', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 230);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 220, 230, '发起审批', 2, '/flow/approve/launch', 'flow/approve/launch/index', 'Promotion', 1, 'flow:approve:launch', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 220);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 221, 230, '待审批', 2, '/flow/approve/pending', 'flow/approve/pending/index', 'Document', 2, 'flow:approve:todo', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 221);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 226, 230, '我的申请', 2, '/flow/approve/mine', 'flow/approve/mine/index', 'Folder', 3, 'flow:approve:mine', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 226);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 227, 230, '我收到的', 2, '/flow/approve/received', 'flow/approve/received/index', 'User', 4, 'flow:approve:received', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 227);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 228, 230, '认领任务', 2, '/flow/approve/claim', 'flow/approve/claim/index', 'Box', 5, 'flow:approve:claim', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 228);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 229, 230, '已审批', 2, '/flow/approve/approved', 'flow/approve/approved/index', 'CircleCheck', 6, 'flow:approve:approved', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 229);

-- 审批详情：隐藏路由（启用但不在侧栏显示）
INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 225, 230, '审批详情', 2, '/flow/approve/detail', 'flow/approve/detail/index', '', 99, 'flow:approve:todo', 1, 0, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 225);

UPDATE `sys_menu`
SET `parent_id` = 230,
    `name` = '发起审批',
    `sort` = 1,
    `icon` = 'Promotion',
    `path` = '/flow/approve/launch',
    `component` = 'flow/approve/launch/index',
    `updated_at` = NOW(3)
WHERE `id` = 220;

UPDATE `sys_menu`
SET `parent_id` = 230,
    `name` = '待审批',
    `sort` = 2,
    `icon` = 'Document',
    `path` = '/flow/approve/pending',
    `component` = 'flow/approve/pending/index',
    `updated_at` = NOW(3)
WHERE `id` = 221;

UPDATE `sys_menu`
SET `status` = 1,
    `visible` = 0,
    `parent_id` = 230,
    `path` = '/flow/approve/detail',
    `component` = 'flow/approve/detail/index',
    `permission` = 'flow:approve:todo',
    `sort` = 99,
    `updated_at` = NOW(3)
WHERE `id` = 225;

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 222, 220, '发起', 3, '', '', '', 1, 'flow:approve:launch', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 222);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 223, 221, '待办查询', 3, '', '', '', 1, 'flow:approve:todo', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 223);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 224, 221, '审批处理', 3, '', '', '', 2, 'flow:approve:handle', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 224);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 231, 226, '申请查询', 3, '', '', '', 1, 'flow:approve:mine', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 231);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 232, 227, '抄送查询', 3, '', '', '', 1, 'flow:approve:received', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 232);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 233, 228, '认领查询', 3, '', '', '', 1, 'flow:approve:claim', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 233);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 234, 229, '已审查询', 3, '', '', '', 1, 'flow:approve:approved', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 234);

-- 流程监控：挂在「工作流」下（与流程管理同级）
INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 235, 200, '流程监控', 2, '/flow/approve/monitor', 'flow/approve/monitor/index', 'Monitor', 3, 'flow:approve:monitor', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 235);

UPDATE `sys_menu`
SET `parent_id` = 200, `sort` = 3, `name` = '流程监控',
    `path` = '/flow/approve/monitor', `component` = 'flow/approve/monitor/index',
    `updated_at` = NOW(3)
WHERE `id` = 235;

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 236, 235, '监控查询', 3, '', '', '', 1, 'flow:approve:monitor', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 236);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 237, 235, '终止流程', 3, '', '', '', 2, 'flow:approve:terminate', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 237);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 238, 235, '管理员转办', 3, '', '', '', 3, 'flow:approve:adminTransfer', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 238);

-- 演示业务单据
INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 240, 200, '测试业务单据', 2, '/business/purchase-order', 'business/demo/order', 'ShoppingCart', 4, 'business:purchaseOrder:list', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 240);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 241, 240, '单据查询', 3, '', '', '', 1, 'business:purchaseOrder:list', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 241);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 242, 240, '单据新增', 3, '', '', '', 2, 'business:purchaseOrder:add', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 242);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 243, 240, '单据删除', 3, '', '', '', 3, 'business:purchaseOrder:delete', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 243);

INSERT INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `sort`, `permission`, `status`, `visible`, `keep_alive`, `created_at`, `updated_at`, `deleted_at`)
SELECT 244, 240, '发起审批', 3, '', '', '', 4, 'business:purchaseOrder:launch', 1, 1, 1, NOW(3), NOW(3), NULL
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `id` = 244);

-- 015 曾建「流程定义」239，最终已废弃：软删
UPDATE `sys_menu`
SET `deleted_at` = NOW(3), `updated_at` = NOW(3)
WHERE `id` = 239 AND `deleted_at` IS NULL;

-- 流程管理保持列表页（防止曾被 015 改成目录）
UPDATE `sys_menu`
SET `type` = 2,
    `path` = '/flow/process',
    `component` = 'flow/process/index',
    `permission` = 'flow:process:list',
    `icon` = 'SetUp',
    `sort` = 1,
    `parent_id` = 200,
    `updated_at` = NOW(3)
WHERE `id` = 201;

UPDATE `sys_menu`
SET `parent_id` = 201, `updated_at` = NOW(3)
WHERE `id` IN (202, 203, 204, 205, 206);

INSERT INTO `sys_role_menu` (`role_id`, `menu_id`)
SELECT 1, m.id FROM `sys_menu` m
WHERE m.id IN (
  220, 221, 222, 223, 224, 225, 226, 227, 228, 229,
  230, 231, 232, 233, 234, 235, 236, 237, 238,
  240, 241, 242, 243, 244
)
  AND NOT EXISTS (SELECT 1 FROM `sys_role_menu` rm WHERE rm.role_id = 1 AND rm.menu_id = m.id);
