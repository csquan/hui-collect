DROP TABLE IF EXISTS `t_part_rebalance_task`;
CREATE TABLE `t_part_rebalance_task` (
    `f_id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_state`      tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `f_params`     text                NOT NULL COMMENT '任务数据',
    `f_message`    text                NOT NULL COMMENT '',
    `f_created_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`f_state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='小r任务表';

DROP TABLE IF EXISTS `transfer_task`;
CREATE TABLE `transfer_task` (
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `rebalance_id`  int(11)             NOT NULL DEFAULT '0' COMMENT 'rebalance task id',
    `state`         tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `transfer_type` tinyint(4)          NOT NULL DEFAULT '0' COMMENT '0:transferIn 1:invest',
    `progress`      varchar(20)         NOT NULL DEFAULT '' COMMENT '当前状态的处理进度',
    `params`        text                NOT NULL COMMENT '任务数据',
    `created_at`    timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='资产转移';

DROP TABLE IF EXISTS `t_transaction_task`;
CREATE TABLE `t_transaction_task` (
    `f_id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_rebalance_id`     int(11)             NOT NULL DEFAULT '0' COMMENT 'rebalance id',
    `f_transfer_id`      int(11)             NOT NULL DEFAULT '0' COMMENT 'transfer id',
    `f_transfer_type`    tinyint(4)          NOT NULL DEFAULT '0' COMMENT '0:transferIn 1:invest',
    `f_nonce`            int(11)             NOT NULL DEFAULT '0' COMMENT 'nonce',
    `f_chain_id`         int(11)             NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `f_params`           text                NOT NULL COMMENT '任务数据',
    `f_decimal`          int(11)             NOT NULL DEFAULT '0' COMMENT 'decimal',
    `f_from`             char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_to`               char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_contract_address` char(42)            NOT NULL DEFAULT '' COMMENT 'contract addr',
    `f_value`            char(42)            NOT NULL DEFAULT '' COMMENT '数量,以最大精度计算',
    `f_input_data`       text                NOT NULL  COMMENT '合约的inputdata',
    `f_Cipher`           text                NOT NULL  COMMENT '加密数据的解密key',
    `f_EncryptData`      text                NOT NULL  COMMENT '加密后数据',
    `f_signed_data`      text                NOT NULL  COMMENT '签名后数据',
    `f_order_id`         int(11)             NOT NULL  COMMENT  '订单ID',
    `f_hash`             char(66)            NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `f_created_at`       timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`       timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='交易';

DROP TABLE IF EXISTS `invest_task`;
CREATE TABLE `invest_task` (
    `id`           bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `rebalance_id` int(11)             NOT NULL DEFAULT '0' COMMENT 'rebalance task id',
    `state`        tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `params`       text                NOT NULL COMMENT '任务数据',
    `created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='组LP';

DROP TABLE IF EXISTS `cross_task`;
CREATE TABLE `cross_task` (
    `id`              bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `rebalance_id`    int(11)             NOT NULL DEFAULT '0' COMMENT 'rebalance task id',
    `chain_from`      varchar(10)         NOT NULL DEFAULT '' COMMENT 'from chain',
    `chain_to`        varchar(10)         NOT NULL DEFAULT '' COMMENT 'to chain',
    `chain_from_addr` char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `chain_to_addr`   char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `currency_from`   varchar(10)         NOT NULL DEFAULT '' COMMENT 'from currency',
    `currency_to`     varchar(10)         NOT NULL DEFAULT '' COMMENT 'to currency',
    `amount`          varchar(256)        NOT NULL DEFAULT '0' COMMENT 'amount',
    `state`           tinyint(4)          NOT NULL DEFAULT 0 COMMENT 'task state 0:等待创建子任务 1:子任务创建完成 2:任务完成',
    `created_at`      timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_rebalance_id` (`rebalance_id`),
    KEY `idx_state` (`state`)
)
    DEFAULT CHARSET = utf8mb4;

DROP TABLE IF EXISTS `cross_sub_task`;
CREATE TABLE `cross_sub_task` (
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `parent_id`      bigint(20) unsigned NOT NULL DEFAULT 0,
    `task_no`        int(11)             NOT NULL DEFAULT 0 COMMENT 'task number 相同parent_id下保持唯一递增',
    `bridge_task_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '在跨链服务中的taskid',
    `amount`         varchar(256)        NOT NULL DEFAULT '0' COMMENT '跨链数量',
    `state`          tinyint(4)          NOT NULL DEFAULT 0 COMMENT 'task state 0:等待跨链 1:跨链执行中 2:跨链完成',
    `created_at`     timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_parent_id` (`parent_id`)
)
    DEFAULT CHARSET = utf8mb4;
