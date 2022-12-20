CREATE TABLE `t_transaction_task` (
    `f_id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_message`           text COMMENT 'message',
    `f_type`              tinyint(4)          NOT NULL DEFAULT '0' COMMENT '0:transferIn 1:invest',
    `f_params`            text COMMENT '任务数据',
    `f_state`             tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'state',
    `f_chain_id`          int(11)             NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `f_chain_name`        char(20)            NOT NULL DEFAULT '' COMMENT 'chain name',
    `f_from`              char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_to`                char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_nonce`             int(11)             NOT NULL DEFAULT 0 COMMENT 'nonce',
    `f_gas_price`         varchar(255)        NOT NULL DEFAULT 0 COMMENT 'gas_price',
    `f_contract_address`  char(42)            NOT NULL DEFAULT '' COMMENT 'contract addr',
    `f_input_data`        text COMMENT '合约的inputdata',
    `f_cipher`            text COMMENT '加密数据的解密key',
    `f_encrypt_data`      text COMMENT '加密后数据',
    `f_signed_data`       text COMMENT '签名后数据',
    `f_order_id`          bigint              NOT NULL DEFAULT 0 COMMENT '订单ID',
    `f_hash`              char(66)            NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `f_gas_limit`         varchar(255)        NOT NULL DEFAULT '2000000' COMMENT 'gas_price',
    `f_amount`            varchar(255)        NOT NULL DEFAULT '0' COMMENT '交易中的value值，合约交易可填0',
    `f_quantity`          varchar(255)        NOT NULL DEFAULT '0' COMMENT '资产金额，1000000000000000000，按币种最小精度记',
    `f_created_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`f_state`),
    KEY `idx_from_nonce` (`f_from`, `f_nonce`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COMMENT ='交易';

