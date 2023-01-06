CREATE TABLE `t_transaction_task` (
    `f_id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_parent_id`         char(42)            NOT NULL DEFAULT '0' COMMENT 'parent_id',
    `f_uuid`              char(42)            NOT NULL DEFAULT '' COMMENT 'uuid-唯一业务流水号',
    `f_uid`               char(42)            NOT NULL DEFAULT '' COMMENT 'user id-同一用户uid相同',
    `f_request_id`        varchar(255)        NOT NULL DEFAULT '' COMMENT 'request id',
    `f_chain_id`          int(11)             NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `f_from`              char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_to`                char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_receiver`          char(42)            NOT NULL DEFAULT '' COMMENT 'receiver addr',
    `f_value`             char(42)            NOT NULL DEFAULT '' COMMENT 'value',
    `f_nonce`             int(11)             NOT NULL DEFAULT 0 COMMENT 'nonce',
    `f_gas_price`         varchar(255)        NOT NULL DEFAULT 0 COMMENT 'gas_price',
    `f_input_data`        text COMMENT '合约的input data',
    `f_sig`               text COMMENT 'sig',
    `f_sign_hash`         char(66)            NOT NULL DEFAULT '' COMMENT 'sign hash--签名前的hash',
    `f_tx_hash`           char(66)            NOT NULL DEFAULT '' COMMENT 'tx hash--交易广播的hash',
    `f_gas_limit`         varchar(255)        NOT NULL DEFAULT '2000000' COMMENT 'gas_price',
    `f_receipt`           text COMMENT '交易的收据',
    `f_state`             tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'state',
    `f_type`             tinyint(4)          NOT NULL DEFAULT '0' COMMENT  'type',
    `f_error`             char(66)            NOT NULL DEFAULT '' COMMENT 'error',
    `f_retry_times`       int(11)             NOT NULL DEFAULT 0 COMMENT 'retry_times--错误后',
    `f_created_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`f_state`),
    KEY `idx_from_nonce` (`f_from`, `f_nonce`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COMMENT ='归集交易表';

DROP TABLE IF EXISTS `t_src_tx`;
CREATE TABLE `t_src_tx`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `tx_hash`          char(66)       NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `addr`             char(42)       NOT NULL DEFAULT '' COMMENT '合约地址',
    `sender`           char(42)       NOT NULL DEFAULT '' COMMENT '发送地址',
    `receiver`         char(42)       NOT NULL DEFAULT '' COMMENT '接收地址',
    `token_cnt`        decimal(65, 0) NOT NULL DEFAULT '0' COMMENT 'token 个数',
    `log_index`        int(11) DEFAULT NULL COMMENT 'transaction log index',
    `token_cnt_origin` varchar(78)    NOT NULL DEFAULT '' COMMENT 'token cnt origin value',
    `create_time`      timestamp      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `block_state`      tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:ok 1:rollback',
    `collect_state`    tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:ready 1:ing 2:ed',
    `block_num`        bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'block number',
    `block_time`       int(11) NOT NULL DEFAULT '0' COMMENT '打包时间',
    `type`             tinyint(4) NOT NULL DEFAULT '0' COMMENT '交易类型-0-打gas交易 1-可以直接归集',
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY                `idx_txhash` (`tx_hash`),
    KEY                `idx_addr` (`addr`),
    KEY                `idx_sender` (`sender`),
    KEY                `idx_receiver` (`receiver`),
    KEY                `idx_block_num` (`block_num`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='归集源交易表';


DROP TABLE IF EXISTS `t_account`;
CREATE TABLE `t_account`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `addr`           char(42)       NOT NULL DEFAULT '' COMMENT 'address',
    `balance`        decimal(65, 0) NOT NULL DEFAULT '0' COMMENT '账户数额',
    `height`         bigint(20) NOT NULL DEFAULT '0' COMMENT '更新时区块高度',
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    UNIQUE KEY `uk_addr` (`addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='总账本';

DROP TABLE IF EXISTS `t_monitor`;
CREATE TABLE `t_monitor`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `addr`           char(42)       NOT NULL DEFAULT '' COMMENT 'address',
    `height`         bigint(20) NOT NULL DEFAULT '0' COMMENT '更新时区块高度',
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    UNIQUE KEY `uk_addr` (`addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='监控表';