
drop table IF EXISTS `t_strategy`;
create TABLE `t_strategy` (
    `f_id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_chain`      varchar(255)        NOT NULL COMMENT '链名称',
    `f_project`    varchar(255)        NOT NULL COMMENT '项目名称',
    `f_currency0`  varchar(255)        NOT NULL COMMENT '币种0',
    `f_currency1`  varchar(255) COMMENT '币种1',
    `f_created_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at` timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP,
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='投资策略表';



insert into t_strategy(f_chain, f_project, f_currency0, f_currency1)
values ('bsc', 'biswap', 'bnb', 'usd'),
       ('bsc', 'pancake', 'bnb', 'usd'),
       ('bsc', 'biswap', 'bnb', 'usdt'),
       ('bsc', 'pancake', 'bnb', 'usdt'),
       ('bsc', 'pancake', 'cake', 'usd'),
       ('bsc', 'pancake', 'cake', 'usdt'),
       ('bsc', 'biswap', 'btc', 'usdt'),
       ('bsc', 'biswap', 'eth', 'usdt'),
       ('bsc', 'solo', 'bnb', null),
       ('bsc', 'solo', 'cake', null),
       ('bsc', 'solo', 'btc', null),
       ('bsc', 'solo', 'eth', null),
       ('bsc', 'solo', 'usdt', null),
       ('bsc', 'solo', 'usd', null),
       ('polygon', 'quickswap', 'eth', 'usdc'),
       ('polygon', 'quickswap', 'eth', 'usdt'),
       ('polygon', 'quickswap', 'btc', 'usdc'),
       ('polygon', 'quickswap', 'matic', 'usdc'),
       ('polygon', 'quickswap', 'matic', 'usdt'),
       ('polygon', 'solo', 'btc', null),
       ('polygon', 'solo', 'eth', null),
       ('polygon', 'solo', 'matic', null),
       ('polygon', 'solo', 'usdt', null),
       ('polygon', 'solo', 'usdc', null);


drop table IF EXISTS `t_currency`;
create TABLE `t_currency` (
    `f_id`          bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_created_at`  timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`  timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP,
    `f_name`        varchar(255)        NOT NULL COMMENT '名称',
    `f_min`         decimal(20, 8) COMMENT '跨链的最小额度',
    `f_cross_scale` decimal(20, 8) COMMENT '跨链的最小精度',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='币种表';


insert into t_currency(f_name, f_min, f_cross_scale)
values ('btc', 0.001, 3),
       ('bnb', null, null),
       ('cake', null, null),
       ('matic', null, null),
       ('eth', 0.01, 2),
       ('usdt', 10, 0),
       ('usdc', 2, 0),
       ('dai', 1, 0),
       ('usd', 10, 0);


drop table IF EXISTS `t_token`;
create TABLE `t_token` (
    `f_id`           bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `f_created_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `f_updated_at`   timestamp           NOT NULL DEFAULT CURRENT_TIMESTAMP ON update CURRENT_TIMESTAMP,
    `f_currency`     varchar(255)        NOT NULL COMMENT '币种',
    `f_chain`        varchar(255)        NOT NULL COMMENT '链',
    `f_symbol`       varchar(255)        NOT NULL COMMENT 'token symbol',
    `f_address`      varchar(255)        NOT NULL COMMENT 'token contract address',
    `f_decimal`      integer             NOT NULL COMMENT '精度',
    `f_cross_symbol` varchar(255)        NOT NULL COMMENT 'cross bridge symbol',
    PRIMARY KEY (`f_id`) /*T![clustered_index] CLUSTERED */
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='币种表';

insert into t_token(f_currency, f_chain, f_symbol, f_address, f_decimal, f_cross_symbol)
values ('btc', 'heco', 'HBTC', '0x66a79d23e58475d2738179ca52cd0b41d73f0bea', 18, 'btc'),
       ('btc', 'bsc', 'BTCB', '0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c', 18, 'btc'),
       ('btc', 'polygon', 'WBTC', '0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6', 8, 'btc'),
       ('bnb', 'bsc', 'BNB', '', 18, 'bnb'),
       ('cake', 'bsc', 'CAKE', '', 18, 'cake'),
       ('matic', 'bsc', 'MATIC', '', 18, 'matic'),
       ('eth', 'heco', 'ETH', '0x64ff637fb478863b7468bc97d30a5bf3a428a1fd', 18, 'eth'),
       ('eth', 'bsc', 'ETH', '0x2170ed0880ac9a755fd29b2688956bd959f933f8', 18, 'eth'),
       ('eth', 'polygon', 'WETH', '0x7ceb23fd6bc0add59e62ac25578270cff1b9f619', 18, 'eth'),
       ('usdt', 'heco', 'USDT', '0xa71edc38d189767582c38a3145b5873052c3e47a', 18, 'usdt'),
       ('usdt', 'bsc', 'USDT', '0x55d398326f99059fF775485246999027B3197955', 18, 'usdt'),
       ('usdt', 'polygon', 'USDT', '0xc2132d05d31c914a87c6611c10748aeb04b58e8f', 6, 'usdt'),
       ('usdc', 'heco', 'USDC', '0x9362bbef4b8313a8aa9f0c9808b80577aa26b73b', 6, 'usdc'),
       ('usdc', 'polygon', 'USDC', '0x2791bca1f2de4661ed88a30c99a7a9449aa84174', 6, 'usdc'),
       ('dai', 'heco', 'DAI-HECO', '0x9362bbef4b8313a8aa9f0c9808b80577aa26b73b', 18, 'dai'),
       ('dai', 'polygon', 'DAI', '0x8f3cf7ad23cd3cadbd9735aff958023239c6a063', 18, 'dai'),
       ('usd', 'heco', 'HUSD', '0x0298c2b32eae4da002a15f36fdf7615bea3da047', 8, 'husd'),
       ('usd', 'bsc', 'BUSD', '0xe9e7cea3dedca5984780bafc599bd69add087d56', 18, 'busd');
