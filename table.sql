CREATE TABLE `part_rebalance_task` (
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `state`      tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `params`     text                NOT NULL DEFAULT '' COMMENT '任务数据',
    `created_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='小r任务表';

CREATE TABLE `transfer_task` (
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `rebalance_id`  int(11)             NOT NULL DEFAULT '0' COMMENT 'rebalance task id',
    `state`         tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `transfer_type` tinyint(4)          NOT NULL DEFAULT '0' COMMENT '0:transferOut 1:transferIn',
    `progress`      varchar(20)         NOT NULL DEFAULT '' COMMENT '当前状态的处理进度',
    `params`        text                NOT NULL DEFAULT '' COMMENT '任务数据',
    `created_at`    timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='资产转移';

CREATE TABLE `transaction_task` (
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `rebalance_id`     int(11)             NOT NULL DEFAULT '0' COMMENT 'rebalance id',
    `transfer_id`      int(11)             NOT NULL DEFAULT '0' COMMENT 'transfer id',
    `nonce`            int(11)             NOT NULL DEFAULT '0' COMMENT 'nonce',
    `chain_id`         int(11)             NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `from`             char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `to`               char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `state`            tinyint(4)          NOT NULL DEFAULT '0' COMMENT '',
    `contract_address` char(42)            NOT NULL DEFAULT '' COMMENT 'contract addr',
    `value`            int(11)             NOT NULL DEFAULT '0' COMMENT '数量',
    `unsigned_data`    text                NOT NULL DEFAULT '' COMMENT '签名前数据',
    `signed_data`      text                NOT NULL DEFAULT '' COMMENT '签名后数据',
    `params`           tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init unsinged signed broadcast',
    `hash`             char(66)            NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `created_at`       timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='交易';

CREATE TABLE `invest_task` (
    `id`           bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `rebalance_id` int(11)             NOT NULL DEFAULT '0' COMMENT 'rebalance task id',
    `state`        tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `params`       text                NOT NULL DEFAULT '' COMMENT '任务数据',
    `created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='组LP';
