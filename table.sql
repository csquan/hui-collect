DROP TABLE IF EXISTS `t_transaction_task`;
CREATE TABLE `t_transaction_task` (
    `f_id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_parent_ids`        text                NOT NULL DEFAULT '0' COMMENT 'parent_ids',
    `f_uuid`              char(42)            NOT NULL DEFAULT '' COMMENT 'uuid-唯一业务流水号',
    `f_uid`               char(42)            NOT NULL DEFAULT '' COMMENT 'user id-同一用户uid相同',
    `f_request_id`        varchar(255)        NOT NULL DEFAULT '' COMMENT 'request id',
    `f_chain`             char(42)            NOT NULL DEFAULT '' COMMENT 'chain',
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
    `f_type`              tinyint(4)          NOT NULL DEFAULT '0' COMMENT  'type',
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
    `f_id`                       bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `f_chain`                    char(66)  NOT NULL DEFAULT '' COMMENT '链名称',
    `f_symbol`                   char(42)  NOT NULL DEFAULT '' COMMENT '代币符号',
    `f_address`                  char(42)  NOT NULL DEFAULT '' COMMENT '账户地址',
    `f_uid`                      char(42)  NOT NULL DEFAULT '' COMMENT '账户uid',
    `f_balance`                  char(42)  NOT NULL DEFAULT '' COMMENT '代币余额',
    `f_status`                   tinyint(4) NOT NULL COMMENT '状态',
    `f_ownerType`                tinyint(4) NOT NULL COMMENT '账户类型',
    `f_collect_state`            tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:ready 1:ing 2:ed',
    `f_created_at`               timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`               timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='归集源交易表';


DROP TABLE IF EXISTS `t_monitor`;
CREATE TABLE `t_monitor`
(
    `f_id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `f_uid`            text COMMENT 'uid',
    `f_appid`          text COMMENT 'appid',
    `f_addr`           char(42)       NOT NULL DEFAULT '' COMMENT 'address',
    `f_chain`          char(42)       NOT NULL DEFAULT '' COMMENT 'chain',
    `f_height`         bigint(20) NOT NULL DEFAULT '0' COMMENT '更新时区块高度',
    `f_created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
    `f_updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP COMMENT 'time',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    UNIQUE KEY `uk_addr` (`f_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='监控表';

DROP TABLE IF EXISTS `t_monitor_hash`;
CREATE TABLE `t_monitor_hash` (
      `f_id` bigint unsigned NOT NULL AUTO_INCREMENT,
      `f_hash` char(66) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'tx hash--交易广播的hash',
      `f_chain` char(42) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '链名称',
      `f_order_id` text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin COMMENT '回调地址',
      `f_created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'time',
      `f_updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'time',
      `f_get_receipt_times` int unsigned DEFAULT NULL,
      `f_receipt_state` int DEFAULT NULL,
      `f_push_state` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
      `f_contract_addr` varchar(100) COLLATE utf8mb4_bin DEFAULT NULL,
      `f_gasLimit` int unsigned DEFAULT NULL,
      `f_gasPrice` varchar(100) COLLATE utf8mb4_bin DEFAULT NULL,
      `f_gasUsed` int unsigned DEFAULT NULL,
      `f_index` int DEFAULT NULL,
      PRIMARY KEY (`f_id`),
      UNIQUE KEY `uk_addr` (`f_hash`,`f_chain`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='监控表';
