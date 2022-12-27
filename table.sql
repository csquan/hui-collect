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
    COMMENT ='交易';

