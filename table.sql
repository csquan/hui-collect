DROP TABLE IF EXISTS `t_transaction_task`;
CREATE TABLE `t_transaction_task` (
    `f_id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_parent_ids`        text                NOT NULL DEFAULT '0' COMMENT 'parent_ids',
    `f_uuid`              char(42)            NOT NULL DEFAULT '' COMMENT 'uuid-唯一业务流水号',
    `f_uid`               char(42)            NOT NULL DEFAULT '' COMMENT 'user id-同一用户uid相同',
    `f_request_id`        varchar(255)        NOT NULL DEFAULT '' COMMENT 'request id',
    `f_chain_id`          int(11)             NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `f_from`              char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_to`                char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_contract_addr`     char(42)            NOT NULL DEFAULT '' COMMENT 'contract addr',
    `f_receiver`          char(42)            NOT NULL DEFAULT '' COMMENT 'receiver addr',
    `f_value`             char(42)            NOT NULL DEFAULT '' COMMENT 'value',
    `f_amount`            char(42)            NOT NULL DEFAULT '' COMMENT 'amount',
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
    `f_id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `f_tx_hash`          char(66)       NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `f_addr`             char(42)       NOT NULL DEFAULT '' COMMENT '合约地址',
    `f_sender`           char(42)       NOT NULL DEFAULT '' COMMENT '发送地址',
    `f_receiver`         char(42)       NOT NULL DEFAULT '' COMMENT '接收地址',
    `f_token_cnt`        decimal(65, 0) NOT NULL DEFAULT '0' COMMENT 'token 个数',
    `f_log_index`        int(11) DEFAULT NULL COMMENT 'transaction log index',
    `f_token_cnt_origin` varchar(78)    NOT NULL DEFAULT '' COMMENT 'token cnt origin value',
    `f_block_state`      tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:ok 1:rollback',
    `f_collect_state`    tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:ready 1:ing 2:ed',
    `f_block_num`        bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'block number',
    `f_block_time`       int(11) NOT NULL DEFAULT '0' COMMENT '打包时间',
    `f_type`             tinyint(4) NOT NULL DEFAULT '0' COMMENT '交易类型-0-打gas交易 1-可以直接归集',
    `f_chain`            char(42)       NOT NULL DEFAULT '' COMMENT '链名称',
    `f_created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY                `idx_txhash` (`f_tx_hash`),
    KEY                `idx_addr` (`f_addr`),
    KEY                `idx_sender` (`f_sender`),
    KEY                `idx_receiver` (`f_receiver`),
    KEY                `idx_block_num` (`f_block_num`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='归集源交易表';


DROP TABLE IF EXISTS `t_account`;
CREATE TABLE `t_account`
(
    `f_id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `f_addr`           char(42)       NOT NULL DEFAULT '' COMMENT 'account address',
    `f_contract_addr`  char(42)       NOT NULL DEFAULT '' COMMENT 'contract address',
    `f_balance`        decimal(65, 0) NOT NULL DEFAULT '0' COMMENT '账户数额',
    `f_last_check`      bigint(20) NOT NULL DEFAULT '0' COMMENT '更新时区块高度',
    `f_created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    UNIQUE KEY `uk_addr` (`f_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='总账本';

DROP TABLE IF EXISTS `t_monitor`;
CREATE TABLE `t_monitor`
(
    `f_id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `f_addr`           char(42)       NOT NULL DEFAULT '' COMMENT 'address',
    `f_height`         bigint(20) NOT NULL DEFAULT '0' COMMENT '更新时区块高度',
    `f_created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    UNIQUE KEY `uk_addr` (`f_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='监控表';

DROP TABLE IF EXISTS `t_token`;
create TABLE `t_token` (
    `f_id`           bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_threshold`    varchar(255)        NOT NULL COMMENT '归集门槛',
    `f_chain`        varchar(255)        NOT NULL COMMENT '链',
    `f_symbol`       varchar(255)        NOT NULL COMMENT 'token symbol',
    `f_address`      varchar(255)        NOT NULL COMMENT 'token contract address',
    `f_decimal`      integer             NOT NULL COMMENT '精度',
    `f_created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
)
ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COMMENT ='监控币种表';

INSERT INTO t_token (f_id,f_threshold,f_chain,f_symbol,f_address,f_decimal,f_created_at,f_updated_at) VALUES
            (1,'10000000000000000000','HUI','TSC','0x99Ac689Fd1f09AdA4c0365E6497B2A824Af68557',18,'2023-01-07 10:51:17','2023-01-07 10:56:05'),
            (2,'10000000000000000000','HUI','TSC1','0xe7df395C170973654A2B054115146f02eE6DfbA5',18,'2023-01-07 10:52:17','2023-01-07 10:56:05'),
            (3,'10000000000','HUI','TSC111','0x6B98aaa1f8A92ceCA108A49CFe7ee4081B7aF8F8',9,'2023-01-07 10:53:32','2023-01-07 10:56:05');
