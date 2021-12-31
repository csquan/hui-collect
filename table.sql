DROP TABLE IF EXISTS `t_full_rebalance_task`;
CREATE TABLE `t_full_rebalance_task` (
    `f_id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_state`      tinyint(4)          NOT NULL DEFAULT '0' COMMENT '',
    `f_params`     text                NOT NULL COMMENT '任务数据',
    `f_message`    text                NOT NULL COMMENT '',
    `f_created_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`f_state`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='大r任务表';

DROP TABLE IF EXISTS `t_part_rebalance_task`;
CREATE TABLE `t_part_rebalance_task` (
    `f_id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_task_id`           varchar(20) NOT NULL DEFAULT '' COMMENT '全局任务id',
    `f_full_rebalance_id` bigint(20) unsigned COMMENT '大r任务id',
    `f_state`             tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'init build ongoing success failed',
    `f_params`            text                NOT NULL COMMENT '任务数据',
    `f_message`           text                NOT NULL COMMENT '',
    `f_create_unix`       bigint(20)          not null default 0 comment '任务创建的unix时间戳',
    `f_created_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`f_state`, `f_create_unix`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='小r任务表';

DROP TABLE IF EXISTS `t_transaction_task`;
CREATE TABLE `t_transaction_task` (
    `f_id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_full_rebalance_id` bigint(20) unsigned COMMENT '大r任务id',
    `f_rebalance_id`      int(11)             NOT NULL DEFAULT '0' COMMENT '小r任务 id',
    `f_message`           text                NOT NULL COMMENT '',
    `f_type`              tinyint(4)          NOT NULL DEFAULT '0' COMMENT '0:transferIn 1:invest',
    `f_params`            text                NOT NULL COMMENT '任务数据',
    `f_state`             tinyint(4)          NOT NULL DEFAULT '0' COMMENT '',
    `f_chain_id`          int(11)             NOT NULL DEFAULT '0' COMMENT 'chain_id',
    `f_chain_name`        char(20)            NOT NULL DEFAULT '' COMMENT 'chain name',
    `f_from`              char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_to`                char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_nonce`             int(11)             NOT NULL DEFAULT 0 COMMENT 'nonce',
    `f_gas_price`         varchar(255)        NOT NULL DEFAULT 0 COMMENT 'gas_price',
    `f_contract_address`  char(42)            NOT NULL DEFAULT '' COMMENT 'contract addr',
    `f_input_data`        text                NOT NULL COMMENT '合约的inputdata',
    `f_cipher`            text                NOT NULL COMMENT '加密数据的解密key',
    `f_encrypt_data`      text                NOT NULL COMMENT '加密后数据',
    `f_signed_data`       text                NOT NULL COMMENT '签名后数据',
    `f_order_id`          bigint              NOT NULL DEFAULT 0 COMMENT '订单ID',
    `f_hash`              char(66)            NOT NULL DEFAULT '' COMMENT 'transaction hash',
    `f_gas_limit`         varchar(255)        NOT NULL DEFAULT '2000000' COMMENT 'gas_price',
    `f_amount`            varchar(255)        NOT NULL DEFAULT '0' COMMENT '交易中的value值，合约交易可填0',
    `f_quantity`          varchar(255)        NOT NULL DEFAULT '0' COMMENT '资产金额，1000000000000000000，按币种最小精度记',
    `f_created_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`        timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */,
    KEY `idx_state` (`f_state`),
    KEY `idx_from_nonce` (`f_from`, `f_nonce`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='交易';



DROP TABLE IF EXISTS `t_cross_task`;
CREATE TABLE `t_cross_task` (
    `f_id`              bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_rebalance_id`    int(11)             NOT NULL DEFAULT '0' COMMENT 'part rebalance task id',
    `f_chain_from`      varchar(10)         NOT NULL DEFAULT '' COMMENT 'from chain',
    `f_chain_to`        varchar(10)         NOT NULL DEFAULT '' COMMENT 'to chain',
    `f_chain_from_addr` char(42)            NOT NULL DEFAULT '' COMMENT 'from addr',
    `f_chain_to_addr`   char(42)            NOT NULL DEFAULT '' COMMENT 'to addr',
    `f_currency_from`   varchar(10)         NOT NULL DEFAULT '' COMMENT 'from currency',
    `f_currency_to`     varchar(10)         NOT NULL DEFAULT '' COMMENT 'to currency',
    `f_amount`          varchar(256)        NOT NULL DEFAULT '0' COMMENT 'amount',
    `f_state`           tinyint(4)          NOT NULL DEFAULT 0 COMMENT 'task state 0:等待创建子任务 1:子任务创建完成 2:任务完成',
    `f_created_at`      timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`      timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`),
    KEY `idx_rebalance_id` (`f_rebalance_id`),
    KEY `idx_state` (`f_state`)
)
    DEFAULT CHARSET = utf8mb4;

DROP TABLE IF EXISTS `t_cross_sub_task`;
CREATE TABLE `t_cross_sub_task` (
    `f_id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_parent_id`      bigint(20) unsigned NOT NULL DEFAULT 0,
    `f_task_no`        int(11)             NOT NULL DEFAULT 0 COMMENT 'task number 相同parent_id下保持唯一递增',
    `f_bridge_task_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '在跨链服务中的taskid',
    `f_amount`         varchar(256)        NOT NULL DEFAULT '0' COMMENT '跨链数量',
    `f_state`          tinyint(4)          NOT NULL DEFAULT 0 COMMENT 'task state 0:等待跨链 1:跨链执行中 2:跨链完成',
    `f_created_at`     timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`     timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`),
    UNIQUE `uk_parent_id_task_no` (`f_parent_id`, `f_task_no`)
)
    DEFAULT CHARSET = utf8mb4;



drop table IF EXISTS `t_strategy`;
create TABLE `t_strategy` (
    `f_id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_chain`      varchar(255)        NOT NULL COMMENT '链名称',
    `f_project`    varchar(255)        NOT NULL COMMENT '项目名称',
    `f_currency0`  varchar(255)        NOT NULL COMMENT '币种0',
    `f_currency1`  varchar(255) COMMENT '币种1',
    `f_enabled`    bool                         default false comment '是否开启',
    `f_created_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='投资策略表';


insert into t_strategy(f_chain, f_project, f_currency0, f_currency1, f_enabled)
values ('bsc', 'biswap', 'bnb', 'usd', false),
       ('bsc', 'pancakeswap', 'bnb', 'usd', false),
       ('bsc', 'biswap', 'bnb', 'usdt', false),
       ('bsc', 'pancakeswap', 'bnb', 'usdt', false),
       ('bsc', 'pancakeswap', 'cake', 'usd', false),
       ('bsc', 'pancakeswap', 'cake', 'usdt', false),
       ('bsc', 'biswap', 'btc', 'usdt', false),
       ('bsc', 'biswap', 'eth', 'usdt', false),
       ('bsc', 'solo.top', 'bnb', null, false),
       ('bsc', 'solo.top', 'cake', null, false),
       ('bsc', 'solo.top', 'btc', null, false),
       ('bsc', 'solo.top', 'eth', null, false),
       ('bsc', 'solo.top', 'usdt', null, false),
       ('bsc', 'solo.top', 'usd', null, false),
       ('polygon', 'quickswap', 'eth', 'usdc', false),
       ('polygon', 'quickswap', 'eth', 'usdt', false),
       ('polygon', 'quickswap', 'matic', 'usdc', false),
       ('polygon', 'quickswap', 'matic', 'usdt', false),
       ('polygon', 'solo.top', 'eth', null, false),
       ('polygon', 'solo.top', 'matic', null, false),
       ('polygon', 'solo.top', 'usdt', null, false),
       ('polygon', 'solo.top', 'usdc', null, false);


drop table IF EXISTS `t_currency`;
create TABLE `t_currency` (
    `f_id`          bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_name`        varchar(255)        NOT NULL COMMENT '名称',
    `f_cross_min`   decimal(20, 8) COMMENT '跨链的最小额度',
    `f_invest_min`  decimal(20, 8) COMMENT '投资的最小额度',
    `f_cross_scale` integer COMMENT '跨链的最小精度',
    `f_created_at`  timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`  timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='资产表';


insert into t_currency(f_name, f_cross_min, f_invest_min, f_cross_scale)
values ('btc', 0.001,0.0001, 3),
       ('bnb', null, 0.01, null),
       ('cake', null, 0.1, null),
       ('matic', null, 0.1, null),
       ('eth', 0.01, 0.01, 2),
       ('usdt', 10, 1, 0),
       ('usdc', 10, 1, 0),
       ('dai', 10, 1, 0),
       ('usd', 10, 1, 0);


drop table IF EXISTS `t_token`;
create TABLE `t_token` (
    `f_id`           bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_currency`     varchar(255)        NOT NULL COMMENT '币种',
    `f_chain`        varchar(255)        NOT NULL COMMENT '链',
    `f_symbol`       varchar(255)        NOT NULL COMMENT 'token symbol',
    `f_address`      varchar(255)        NOT NULL COMMENT 'token contract address',
    `f_decimal`      integer             NOT NULL COMMENT '精度',
    `f_cross_symbol` varchar(255) COMMENT 'cross bridge symbol',
    `f_created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='币种表';

insert into t_token(f_currency, f_chain, f_symbol, f_address, f_decimal, f_cross_symbol)
values ('btc', 'heco', 'HBTC', '0x66a79d23e58475d2738179ca52cd0b41d73f0bea', 18, 'btc'),
       ('btc', 'bsc', 'BTCB', '0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c', 18, 'btc'),
#        ('btc', 'polygon', 'WBTC', '0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6', 8, 'btc'),
       ('bnb', 'bsc', 'WBNB', '0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c', 18, 'bnb'),
       ('cake', 'bsc', 'CAKE', '0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82', 18, 'cake'),
       ('matic', 'bsc', 'MATIC', '0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270', 18, 'matic'),
       ('eth', 'heco', 'ETH', '0x64ff637fb478863b7468bc97d30a5bf3a428a1fd', 18, 'eth'),
       ('eth', 'bsc', 'ETH', '0x2170ed0880ac9a755fd29b2688956bd959f933f8', 18, 'eth'),
       ('eth', 'polygon', 'WETH', '0x7ceb23fd6bc0add59e62ac25578270cff1b9f619', 18, 'eth'),
       ('usdt', 'heco', 'USDT', '0xa71edc38d189767582c38a3145b5873052c3e47a', 18, 'usdt'),
       ('usdt', 'bsc', 'USDT', '0x55d398326f99059ff775485246999027b3197955', 18, 'usdt'),
       ('usdt', 'polygon', 'USDT', '0xc2132d05d31c914a87c6611c10748aeb04b58e8f', 6, 'usdt'),
       ('usdc', 'heco', 'USDC', '0x9362bbef4b8313a8aa9f0c9808b80577aa26b73b', 6, 'usdc'),
       ('usdc', 'polygon', 'USDC', '0x2791bca1f2de4661ed88a30c99a7a9449aa84174', 6, 'usdc'),
       ('dai', 'heco', 'DAI-HECO', '0x9362bbef4b8313a8aa9f0c9808b80577aa26b73b', 18, 'dai'),
       ('dai', 'polygon', 'DAI', '0x8f3cf7ad23cd3cadbd9735aff958023239c6a063', 18, 'dai'),
       ('usd', 'heco', 'HUSD', '0x0298c2b32eae4da002a15f36fdf7615bea3da047', 8, 'usd'),
       ('usd', 'bsc', 'BUSD', '0xe9e7cea3dedca5984780bafc599bd69add087d56', 18, 'usd');
