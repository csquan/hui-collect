DROP TABLE IF EXISTS `t_part_rebalance_task`;
CREATE TABLE `t_part_rebalance_task`
(
    `f_id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_state`      tinyint(4) NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `f_params`     text      NOT NULL COMMENT '任务数据',
    `f_message`    text      NOT NULL COMMENT '',
    `f_created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY            `idx_state` (`f_state`)
) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='小r任务表';

DROP TABLE IF EXISTS `t_transaction_task`;
CREATE TABLE `t_transaction_task`
(
    `f_id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_rebalance_id`     int(11) NOT NULL DEFAULT '0' COMMENT 'rebalance id',
    `f_message`          text      NOT NULL COMMENT '',
    `f_type`             tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:transferIn 1:invest',
    `f_params`           text      NOT NULL COMMENT '任务数据',
    `f_state`            tinyint(4) NOT NULL DEFAULT '0' COMMENT '',
    `f_chain_id`         int(11) NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `f_chain_name`       char(20)  NOT NULL DEFAULT '' COMMENT 'chain name',
    `f_from`             char(42)  NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_to`               char(42)  NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_contract_address` char(42)  NOT NULL DEFAULT '' COMMENT 'contract addr',
    `f_input_data`       text      NOT NULL COMMENT '合约的inputdata',
    `f_cipher`           text      NOT NULL COMMENT '加密数据的解密key',
    `f_encrypt_data`     text      NOT NULL COMMENT '加密后数据',
    `f_signed_data`      text      NOT NULL COMMENT '签名后数据',
    `f_order_id`         int(11) NOT NULL DEFAULT 0 COMMENT '订单ID',
    `f_hash`             char(66)  NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `f_created_at`       timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`       timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY                  `idx_state` (`f_state`)
) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='交易';


DROP TABLE IF EXISTS `t_approve`;
CREATE TABLE `t_approve`
(
    `f_id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_from`       char(42)  NOT NULL DEFAULT '' COMMENT 'erc20 token addr',
    `f_token`      char(42)  NOT NULL DEFAULT '' COMMENT 'erc20 token addr',
    `f_spender`    char(42)  NOT NULL DEFAULT '' COMMENT 'spender addr',
    `f_created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY            `idx_token_spender` (`f_token`,`f_spender`)
) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='授权';


DROP TABLE IF EXISTS `t_cross_task`;
CREATE TABLE `t_cross_task`
(
    `f_id`              bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_rebalance_id`    int(11) NOT NULL DEFAULT '0' COMMENT 'rebalance task id',
    `f_chain_from`      varchar(10)  NOT NULL DEFAULT '' COMMENT 'from chain',
    `f_chain_to`        varchar(10)  NOT NULL DEFAULT '' COMMENT 'to chain',
    `f_chain_from_addr` char(42)     NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_chain_to_addr`   char(42)     NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_currency_from`   varchar(10)  NOT NULL DEFAULT '' COMMENT 'from currency',
    `f_currency_to`     varchar(10)  NOT NULL DEFAULT '' COMMENT 'to currency',
    `f_amount`          varchar(256) NOT NULL DEFAULT '0' COMMENT 'amount',
    `f_state`           tinyint(4) NOT NULL DEFAULT 0 COMMENT 'task state 0:等待创建子任务 1:子任务创建完成 2:任务完成',
    `f_created_at`      timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`      timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`),
    KEY               `idx_rebalance_id` (`f_rebalance_id`),
    KEY               `idx_state` (`f_state`)
) DEFAULT CHARSET = utf8mb4;

DROP TABLE IF EXISTS `t_cross_sub_task`;
CREATE TABLE `t_cross_sub_task`
(
    `f_id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_parent_id`      bigint(20) unsigned NOT NULL DEFAULT 0,
    `f_task_no`        int(11) NOT NULL DEFAULT 0 COMMENT 'task number 相同parent_id下保持唯一递增',
    `f_bridge_task_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '在跨链服务中的taskid',
    `f_amount`         varchar(256) NOT NULL DEFAULT '0' COMMENT '跨链数量',
    `f_state`          tinyint(4) NOT NULL DEFAULT 0 COMMENT 'task state 0:等待跨链 1:跨链执行中 2:跨链完成',
    `f_created_at`     timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`     timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`),
    UNIQUE              `uk_parent_id_task_no` (`f_parent_id`,`f_task_no`)
) DEFAULT CHARSET = utf8mb4;
