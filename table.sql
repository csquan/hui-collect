CREATE TABLE `t_transaction_task` (
    `f_id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_uuid`              char(42)            NOT NULL DEFAULT '' COMMENT 'uuid-唯一业务流水号',
    `f_uid`               char(42)            NOT NULL DEFAULT '' COMMENT 'user id-同一用户uid相同',
    `f_request_id`        varchar(255)        NOT NULL DEFAULT '' COMMENT 'request id',
    `f_chain_id`          int(11)             NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `f_from`              char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_to`                char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
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
    `id`                       bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `addr_to`                  char(42)       NOT NULL DEFAULT '' COMMENT '接收地址',
    `addr_from`                char(42)       NOT NULL DEFAULT '' COMMENT '发送地址',
    `tx_hash`                  char(66)       NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `tx_index`                 int(11) NOT NULL DEFAULT '0' COMMENT 'transaction index',
    `tx_value`                 decimal(65, 0)          DEFAULT NULL COMMENT 'transaction value',
    `input`                    longtext                DEFAULT NULL COMMENT 'transaction input',
    `nonce`                    int(11) NOT NULL DEFAULT '0',
    `gas_price`                decimal(65, 0)          DEFAULT NULL COMMENT 'gas price',
    `gas_limit`                bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'gas limit',
    `gas_used`                 bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'gas used',
    `is_contract`              tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否调用合约交易',
    `is_contract_create`       tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否创建合约交易',
    `block_time`               int(11) NOT NULL DEFAULT '0' COMMENT '打包时间',
    `block_num`                bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'block number',
    `block_hash`               char(66)       NOT NULL DEFAULT '' COMMENT 'block hash',
    `exec_status`              tinyint(4) NOT NULL DEFAULT '0' COMMENT 'transaction 执行结果',
    `create_time`              timestamp      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `block_state`              tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:ok 1:fail',
    `max_fee_per_gas`          decimal(65, 0) NOT NULL DEFAULT '0' COMMENT '最高交易小费',
    `max_priority_fee_per_gas` decimal(65, 0) NOT NULL DEFAULT '0' COMMENT '最高有限小费',
    `burnt_fees`               decimal(65, 0) NOT NULL DEFAULT '0' COMMENT 'burnt fees',
    `base_fee`                 decimal(65, 0) NOT NULL DEFAULT '0' COMMENT 'base fee',
    `tx_type`                  tinyint(4) NOT NULL DEFAULT '0' COMMENT '交易类型',
    `f_state`             tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'state',
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY                        `idx_txhash_blocknum` (`tx_hash`,`block_num`),
    KEY                        `idx_addr_to` (`addr_to`),
    KEY                        `idx_addr_from` (`addr_from`),
    KEY                        `idx_block_num` (`block_num`)
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='账户表';