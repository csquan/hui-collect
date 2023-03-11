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
