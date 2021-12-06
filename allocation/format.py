# -*- coding:utf-8 -*-
import requests
import json
import yaml
import numpy as np
import json

from functools import reduce
from decimal import *
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from orm import *
from scipy.optimize import minimize

targetChain = ['bsc', 'polygon']


def get_pool_info(url):
    # 存储从api获取的poolinfo
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    e = json.loads(string)

    return e['data']


def read_yaml(path):
    with open(path, 'r', encoding='utf8') as f:
        return yaml.safe_load(f.read())


def getconnectinfo(connstr):
    strlist = connstr.split('@')  # 用逗号分割str字符串，并保存到列表
    print(strlist)
    str1 = strlist[0]  # 包含用户名密码的字串
    str2 = strlist[1]  # 包含Ip端口数据库的字串

    user_endpos = str1.index(":")
    user = str1[0:user_endpos]
    password = str1[user_endpos + 1:len(str1)]

    host_startpos = str2.index("(") + 1
    host_endpos = str2.index(":")

    host = str2[host_startpos:host_endpos]
    port_endpos = str2.index(")")
    port = str2[host_endpos + 1:port_endpos]

    db_startpos = str2.index("/")
    db_endpos = str2.index("?")

    db = str2[db_startpos + 1:db_endpos]

    return user, password, host, port, db


def format_addr(addr):
    if addr.startswith('0x'):
        return addr.lower()
    else:
        return ('0x' + addr).lower()


def format_token_name(currency_name_set, name):
    for k in currency_name_set:
        if name.lower().find(k) >= 0:
            return True, k

    return False, name


class Project:
    def __init__(self, chain, name, url):
        self.chain = chain
        self.name = name
        self.url = url

    def get_info(self):
        res = requests.get(self.url)
        string = str(res.content, 'utf-8')
        e = json.loads(string)
        return e['data']


def calc(session, currencies):
    # 注意usdt与usd的区分，别弄混了
    currencyName = [k for k in sorted(currencies.keys(), key=len, reverse=True)]

    print("currencies:{}".format(currencies))

    # 获取rebalance所需业务信息
    reBalanceInputInfo = get_pool_info("http://neptune-hermes-mgt-h5.test-15.huobiapps.com/v2/v1/open/re")

    thresholdOrg = reBalanceInputInfo['threshold']
    vaultInfoList = reBalanceInputInfo['vaultInfoList']

    # 整理出阈值，当前值 进行比较 {usdt:{bsc:{amount:"1", controllerAddress:""}}}
    thresholdFormat = {}
    beforeInfo = {}

    # 计算跨链的初始状态
    for vault in vaultInfoList:
        (ok, name) = format_token_name(currencyName, vault['tokenSymbol'])
        if ok:
            for chain in vault['activeAmount']:
                if name not in beforeInfo:
                    beforeInfo[name] = {}

                beforeInfo[name][chain.lower()] = vault['activeAmount'][chain]
                beforeInfo[name][chain.lower()]['amount'] = Decimal(beforeInfo[name][chain.lower()]['amount'])
                # total = total + vault.activeAmount[chain]
    beforeInfo['usdt'] = {
        'bsc': {
            'amount': Decimal(0),
        },
        'heco': {
            'amount': Decimal(100000000),
        },
        'polygon': {
            'amount': Decimal(100)
        }
    }
    print("init balance info for cross:{}".format(beforeInfo))

    # 计算阈值
    for threshold in thresholdOrg:
        (ok, name) = format_token_name(currencyName, threshold['tokenSymbol'])
        if ok:
            thresholdFormat[name] = threshold['thresholdAmount']
    print("threshold info after format:{}".format(thresholdFormat))

    # 比较阈值
    needReBalance = False
    for name in thresholdFormat:
        # 没有发现相关资产
        if name not in beforeInfo:
            continue

        total = Decimal(0)
        for item in beforeInfo[name].values():
            total += item['amount']

        needReBalance = total > Decimal(thresholdFormat[name])
        if needReBalance:
            break

    # 没超过阈值
    if not needReBalance:
        exit()

    # 获取apr等信息

    projects = [
        Project('bsc', 'pancake', 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'),
        Project('bsc', 'biswap', 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=476'),
        Project('bsc', 'solo', 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=76'),
        Project('polygon', 'quickswap', 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=112'),
    ]

    # chain_project_coin1_coin2
    apr = {}
    price = {}
    dailyReward = {}
    tvl = {}

    for p in projects:
        info = p.get_info()

        for pool in info:
            for token in pool['rewardTokenList'] + pool['depositTokenList']:
                currency = find_currency_by_address(session, format_addr(token['tokenAddress']))
                if currency is not None:
                    price[currency] = token['tokenPrice']

                key = "{}_{}".format(p.chain, p.name)
                for c in pool['poolName'].split("/"):
                    ok, fc = format_token_name(currencyName, c)
                    key = "{}_{}".format(key, fc)

                apr[key] = Decimal(pool['apr'])
                tvl[key] = Decimal(pool['tvl'])
                dailyReward[key] = reduce(lambda x, y: x + y,
                                          map(lambda t: Decimal(t['tokenPrice']) * Decimal(t['dayAmount']),
                                              pool['rewardTokenList']))

                key = "{}_{}".format(p.chain, p.name)
                for c in reversed(pool['poolName'].split("/")):
                    ok, fc = format_token_name(currencyName, c)
                    key = "{}_{}".format(key, fc)

                apr[key] = Decimal(pool['apr'])
                tvl[key] = Decimal(pool['tvl'])
                dailyReward[key] = reduce(lambda x, y: x + y,
                                          map(lambda t: Decimal(t['tokenPrice']) * Decimal(t['dayAmount']),
                                              pool['rewardTokenList']))

    print("apr info:{}", apr)
    print("price info:{}", price)
    print("daily reward info:{}", dailyReward)
    print("tvl info:{}", tvl)

    # 计算跨链的最终状态
    afterInfo = {}
    for currency in beforeInfo:
        strategies = {}
        caps = {}
        for chain in ['bsc', 'polygon']:
            strategies[chain] = find_strategies_by_chain_and_currency(session, chain, currency)
            caps[chain] = Decimal(0)

            for s in strategies[chain]:
                # 先忽略单币
                if s.currency1 is None:
                    continue

                key = "{}_{}_{}".format(s.chain, s.project,
                                        s.currency0) if s.currency1 is None else "{}_{}_{}_{}".format(
                    s.chain, s.project, s.currency0, s.currency1)
                if key not in apr or apr[key] < Decimal(0.18):
                    continue

                caps[chain] += (dailyReward[key] * Decimal(365) - tvl[key] * apr[key])

        total = Decimal(0)
        for item in beforeInfo[currency].values():
            total += item['amount']

        capsTotal = sum(caps.values())
        for k, v in caps.items():
            if v > 0:
                afterInfo[currency] = {
                    k: str(total * v / capsTotal)
                }

    print("calc final state:{}", afterInfo)

    # 跨链信息
    diffMap = {}
    # cross list
    crossList = []

    # 生成跨链参数, 需要考虑最小值
    for currency in afterInfo:
        for chain in ['bsc', 'polygon']:
            if chain not in afterInfo[currency] or currencies[currency].min is None:
                continue

            diff = Decimal(afterInfo[currency][chain]) - beforeInfo[currency][chain]['amount']
            if diff > Decimal(currencies[currency].min) or diff < Decimal(
                    currencies[currency].min) * -1:
                if currency not in diffMap:
                    diffMap[currency] = {}
                diffMap[currency][chain] = diff.quantize(Decimal(10) ** (-1 * currencies[currency].crossDecimal),
                                                         ROUND_DOWN)  # format to min decimal

    print("diff map:{}", diffMap)

    def add_cross_item(currency, fromChain, toChain, amount):
        if amount > Decimal(currencies[currency].min):
            beforeInfo[currency][fromChain]['amount'] -= amount
            beforeInfo[currency][toChain]['amount'] += amount

            crossList.append({
                'from': fromChain,
                'to': toChain,
                'fromCurrency': currencies[currency].tokens[fromChain].crossSymbol,
                'toCurrency': currencies[currency].tokens[toChain].crossSymbol,
                'amount': amount,
            })

    for currency in diffMap:
        targetMap = {
            'bsc': 'polygon',
            'polygon': 'bsc',
        }

        for chain in diffMap[currency]:
            diff = diffMap[currency][chain]

            if diff < 0:
                add_cross_item(currency, chain, targetMap[chain], diff * -1)

            else:
                if beforeInfo[currency]['heco']['amount'] > diff:
                    add_cross_item(currency, 'heco', chain, diff)
                else:

                    add_cross_item(currency, targetMap[chain], chain,
                                   (diff - beforeInfo[currency]['heco']['amount']).quantize(
                                       Decimal(10) ** (-1 * currencies[currency].crossDecimal),
                                       ROUND_DOWN))

                    add_cross_item(currency, 'heco', chain, beforeInfo[currency]['heco']['amount'].quantize(
                        Decimal(10) ** (-1 * currencies[currency].crossDecimal),
                        ROUND_DOWN))

    print("cross info:{}", crossList)

    # 确定最终的资产分布情况
    # X(0) = bnb_amt_biswap_busd
    # X(3) = busd_amt_biswap_bnb

    # X(1) = bnb_amt_pancake_busd
    # X(2) = busd_amt_pancake_bnb

    # X(4) = busd_amt_pancake_cake
    # X(15) = cake_amt_pancake_busd

    # X(5) = bnb_amt_biswap_usdt
    # X(11) = usdt_amt_biswap_bnb

    # X(6) = bnb_amt_pancake_usdt
    # X(12) = usdt_amt_pancake_bnb

    # X(7) = btcb_amt_biswap_usdt
    # X(10) = usdt_amt_biswap_btcb

    # X(8) = eth_amt_biswap_usdt
    # X(9) = usdt_amt_biswap_eth

    # X(13) = usdt_amt_pancake_cake
    # X(14) = cake_amt_pancake_usdt

    # X(16) = eth_amt_quickswap_usdc
    # X(21) = usdc_amt_quickswap_eth

    # X(17) = eth_amt_quickswap_usdt
    # X(22) = usdt_amt_quickswap_eth

    # X(18) = btc_amt_quickswap_usdc
    # X(23) = usdc_amt_quickswap_btc

    # X(19) = matic_amt_quickswap_usdc
    # X(24) = usdc_amt_quickswap_matic

    # X(20) = matic_amt_quickswap_usdt
    # X(25) = usdt_amt_quickswap_matic

    # X(26) = bnb_bsolotop_amt
    # X(27) = cake_bsolotop_amt
    # X(28) = btcb_bsolotop_amt
    # X(29) = eth_bsolotop_amt
    # X(30) = busd_bsolotop_amt
    # X(31) = usdt_bsolotop_amt

    # X(32) = btc_psolotop_amt
    # X(33) = eth_psolotop_amt
    # X(34) = matic_psolotop_amt
    # X(35) = usdt_psolotop_amt
    # X(36) = usdc_psolotop_amt

    def get_amount(currency, chain):
        if currency not in beforeInfo:
            return 0

        if chain not in beforeInfo[currency]:
            return 0

        if 'amount' not in beforeInfo[currency][chain]:
            return 0

        return float(beforeInfo[currency][chain]['amount'])

    def get_price(currency):
        if currency not in price:
            return 0

        return float(price[currency])

    def get_reward(key):
        if key not in dailyReward:
            return 0

        return float(dailyReward[key])

    def get_apr(key):
        if key not in apr:
            return 0

        return float(apr[key])

    def get_tvl(key):
        if key not in tvl:
            return 0

        return float(tvl[key])

    # def function(x):
    #     input = [
    #         ['bsc', 'biswap', 'bnb', 'usd', x[0], x[3]],
    #         ['bsc', 'pancake', 'bnb', 'usd', x[1], x[2]],
    #         ['bsc', 'pancake', 'cake', 'usd', x[15], x[4]],
    #         ['bsc', 'biswap', 'bnb', 'usdt', x[5], x[11]],
    #         ['bsc', 'pancake', 'bnb', 'usdt', x[6], x[12]],
    #         ['bsc', 'biswap', 'btc', 'usdt', x[7], x[10]],
    #         ['bsc', 'biswap', 'eth', 'usdt', x[8], x[9]],
    #         ['bsc', 'pancake', 'cake', 'usdt', x[14], x[13]],
    #         ['polygon', 'quickswap', 'eth', 'usdc', x[16], x[21]],
    #         ['polygon', 'quickswap', 'eth', 'usdt', x[17], x[22]],
    #         ['polygon', 'quickswap', 'btc', 'usdc', x[18], x[23]],
    #         ['polygon', 'quickswap', 'matic', 'usdc', x[19], x[24]],
    #         ['polygon', 'quickswap', 'matic', 'usdt', x[20], x[25]],
    #     ]
    #
    #     total = 0
    #
    #     for [chain, project, c1, c2, x1, x2] in input:
    #         key = '{}_{}_{}_{}'.format(chain, project, c1, c2)
    #         if get_apr(key) < 0.18:
    #             continue
    #
    #         base = (
    #                 get_price(c1) * x1 + get_price(c2) * x2 + get_tvl(key))
    #
    #         if base <= 0:
    #             continue
    #
    #         total += get_reward(key) * (get_price(c1) * x1 + get_price(c2) * x2) / base
    #     print('output', total * -1)
    #     return total * -1

    def function(x):
        input = [
            ['bsc', 'biswap', 'btc', 'usdt', x[7], x[10]],
            ['bsc', 'biswap', 'eth', 'usdt', x[8], x[9]],
            ['bsc', 'pancake', 'cake', 'usdt', x[14], x[13]],
        ]

        total = 0

        for [chain, project, c1, c2, x1, x2] in input:
            key = '{}_{}_{}_{}'.format(chain, project, c1, c2)
            if get_apr(key) < 0.18:
                continue

            base = (
                    get_price(c1) * x1 + get_price(c2) * x2 + get_tvl(key))

            if base <= 0:
                continue

            total += get_reward(key) * (get_price(c1) * x1 + get_price(c2) * x2) / base
        print('output', total * -1)
        return total * -1

    initx = [0] * 26
    bounds = (
        (0, get_amount('bnb', 'bsc')),
        (0, get_amount('bnb', 'bsc')),
        (0, get_amount('usd', 'bsc')),
        (0, get_amount('usd', 'bsc')),
        (0, get_amount('usd', 'bsc')),
        (0, get_amount('bnb', 'bsc')),
        (0, get_amount('bnb', 'bsc')),
        (0, get_amount('btc', 'bsc')),
        (0, get_amount('eth', 'bsc')),
        (0, get_amount('usdt', 'bsc')),
        (0, get_amount('usdt', 'bsc')),
        (0, get_amount('usdt', 'bsc')),
        (0, get_amount('usdt', 'bsc')),
        (0, get_amount('usdt', 'bsc')),
        (0, get_amount('cake', 'bsc')),
        (0, get_amount('cake', 'bsc')),
        (0, get_amount('eth', 'polygon')),
        (0, get_amount('eth', 'polygon')),
        (0, get_amount('btc', 'polygon')),
        (0, get_amount('matic', 'polygon')),
        (0, get_amount('matic', 'polygon')),
        (0, get_amount('usdc', 'polygon')),
        (0, get_amount('usdt', 'polygon')),
        (0, get_amount('usdc', 'polygon')),
        (0, get_amount('usdc', 'polygon')),
        (0, get_amount('usdt', 'polygon')),
    )
    cons = [
        {'type': 'ineq', 'fun': lambda x: get_amount('bnb', 'bsc') - x[0] - x[1] - x[5] - x[6]},
        {'type': 'ineq', 'fun': lambda x: get_amount('cake', 'bsc') - x[14] - x[15]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usd', 'bsc') - x[3] - x[2] - x[4]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usdt', 'bsc') - x[9] - x[10] - x[11] - x[12] - x[13]},
        {'type': 'ineq', 'fun': lambda x: get_amount('matic', 'polygon') - x[19] - x[20]},
        {'type': 'ineq', 'fun': lambda x: get_amount('eth', 'polygon') - x[16] - x[17]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usdt', 'polygon') - x[22] - x[25]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usdc', 'polygon') - x[23] - x[24]},
        {'type': 'eq', 'fun': lambda x: x[0] * get_price('bnb') - x[3] * get_price('usd')},
        {'type': 'eq', 'fun': lambda x: x[1] * get_price('bnb') - x[2] * get_price('usd')},
        {'type': 'eq', 'fun': lambda x: x[15] * get_price('cake') - x[4] * get_price('usd')},
        {'type': 'eq', 'fun': lambda x: x[5] * get_price('bnb') - x[11] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[6] * get_price('bnb') - x[12] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[7] * get_price('btc') - x[10] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[8] * get_price('eth') - x[9] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[14] * get_price('cake') - x[13] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[16] * get_price('eth') - x[21] * get_price('usdc')},
        {'type': 'eq', 'fun': lambda x: x[17] * get_price('eth') - x[22] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[18] * get_price('btc') - x[23] * get_price('usdc')},
        {'type': 'eq', 'fun': lambda x: x[19] * get_price('matic') - x[24] * get_price('usdc')},
        {'type': 'eq', 'fun': lambda x: x[20] * get_price('matic') - x[25] * get_price('usdt')},
    ]

    np.seterr(divide='ignore', invalid='ignore')
    res = minimize(function,
                   x0=np.array(initx),
                   bounds=bounds,
                   constraints=cons,
                   method= 'SLSQP'
                   )
    print("res:{}".format(res))


def calc_demo():
    def f(x):
        return x[0] + x[1]

    b = (
        (1, 2),
        (1, 3),
    )
    np.seterr(divide='ignore', invalid='ignore')
    y = minimize(f,
             x0=[1]*2,
             bounds=b)
    print(y)


if __name__ == '__main__':
    # calc_demo()
    # exit(0)
    # 读取config
    conf = read_yaml("./wang.yaml")

    db = create_engine(conf['db'],
                       encoding='utf-8',  # 编码格式
                       echo=False,  # 是否开启sql执行语句的日志输出
                       pool_recycle=-1  # 多久之后对线程池中的线程进行一次连接的回收（重置） （默认为-1）,其实session并不会被close
                       )
    session = sessionmaker(db)()

    currencies = {x.name: x for x in session.query(Currency).all()}

    for t in session.query(Token).all():
        curr = currencies[t.currency]
        if not hasattr(curr, 'tokens'):
            curr.tokens = {}
        curr.tokens[t.chain] = t

    calc(session, currencies)
