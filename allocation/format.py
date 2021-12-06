# -*- coding:utf-8 -*-
import requests
import yaml
import numpy as np
import json
import utils

from functools import reduce
from decimal import *
from sqlalchemy.orm import sessionmaker
from orm import *
from scipy.optimize import minimize

targetChain = ['bsc', 'polygon']


def get_pool_info(url):
    # 存储从api获取的poolinfo
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    print(string)
    e = json.loads(string)

    return e['data']


def read_yaml(path):
    with open(path, 'r', encoding='utf8') as f:
        return yaml.safe_load(f.read())


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
        print(string)
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
            'amount': Decimal(10000000),
        },
        'polygon': {
            'amount': Decimal(100)
        }
    }

    beforeInfo['eth'] = {
        'bsc': {
            'amount': Decimal(1),
        },
        'heco': {
            'amount': Decimal(0),
        },
        'polygon': {
            'amount': Decimal(0)
        }
    }
    beforeInfo['btc'] = {
        'bsc': {
            'amount': Decimal(1),
        },
        'heco': {
            'amount': Decimal(0),
        },
        'polygon': {
            'amount': Decimal(0)
        }
    }

    beforeInfo['cake'] = {
        'bsc': {
            'amount': Decimal(1),
        }
    }

    beforeInfo['bnb'] = {
        'bsc': {
            'amount': Decimal(20000),
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
        # Project('polygon', 'quickswap', 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=112'),
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
                if currency is not None and currency not in price:
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

                caps[chain] += (dailyReward[key] * Decimal(365) / Decimal(0.18) - tvl[key])

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

    print("cross info:{}", json.dumps(crossList, cls=utils.DecimalEncoder))

    calc_invest_bsc(beforeInfo, price, dailyReward, apr, tvl)
#
#
# def calc_invest_poly(balance_info_dict, price_dict, daily_reward_dict, apr_dict, tvl_dict):
#     # 确定最终的资产分布情况
#     # X(0) = bnb_amt_biswap_busd
#     # X(3) = busd_amt_biswap_bnb
#
#     # X(1) = bnb_amt_pancake_busd
#     # X(2) = busd_amt_pancake_bnb
#
#     # X(15) = cake_amt_pancake_busd
#     # X(4) = busd_amt_pancake_cake
#
#     # X(5) = bnb_amt_biswap_usdt
#     # X(11) = usdt_amt_biswap_bnb
#
#     # X(6) = bnb_amt_pancake_usdt
#     # X(12) = usdt_amt_pancake_bnb
#
#     # X(7) = btcb_amt_biswap_usdt
#     # X(10) = usdt_amt_biswap_btcb
#
#     # X(8) = eth_amt_biswap_usdt
#     # X(9) = usdt_amt_biswap_eth
#
#     # X(14) = cake_amt_pancake_usdt
#     # X(13) = usdt_amt_pancake_cake
#
#     # X(16) = eth_amt_quickswap_usdc
#     # X(21) = usdc_amt_quickswap_eth
#
#     # X(17) = eth_amt_quickswap_usdt
#     # X(22) = usdt_amt_quickswap_eth
#
#     # X(18) = btc_amt_quickswap_usdc
#     # X(23) = usdc_amt_quickswap_btc
#
#     # X(19) = matic_amt_quickswap_usdc
#     # X(24) = usdc_amt_quickswap_matic
#
#     # X(20) = matic_amt_quickswap_usdt
#     # X(25) = usdt_amt_quickswap_matic
#
#     # X(26) = bnb_bsolotop_amt
#     # X(27) = cake_bsolotop_amt
#     # X(28) = btcb_bsolotop_amt
#     # X(29) = eth_bsolotop_amt
#     # X(30) = busd_bsolotop_amt
#     # X(31) = usdt_bsolotop_amt
#
#     # X(32) = btc_psolotop_amt
#     # X(33) = eth_psolotop_amt
#     # X(34) = matic_psolotop_amt
#     # X(35) = usdt_psolotop_amt
#     # X(36) = usdc_psolotop_amt
#
#     def get_amount(currency, chain):
#         if currency not in balance_info_dict:
#             return 0
#
#         if chain not in balance_info_dict[currency]:
#             return 0
#
#         if 'amount' not in balance_info_dict[currency][chain]:
#             return 0
#
#         return float(balance_info_dict[currency][chain]['amount'])
#
#     def get_price(currency):
#         if currency not in price_dict:
#             return 0
#
#         return float(price_dict[currency])
#
#     def get_reward(key):
#         if key not in daily_reward_dict:
#             return 0
#
#         return float(daily_reward_dict[key])
#
#     def get_apr(key):
#         if key not in apr_dict:
#             return 0
#
#         return float(apr_dict[key])
#
#     def get_tvl(key):
#         if key not in tvl_dict:
#             return 0
#
#         return float(tvl_dict[key])
#
#     def function(x):
#         input = [
#             ['bsc', 'biswap', 'bnb', 'usd', x[0], x[3]],
#             ['bsc', 'pancake', 'bnb', 'usd', x[1], x[2]],
#             ['bsc', 'pancake', 'cake', 'usd', x[15], x[4]],
#             ['bsc', 'biswap', 'bnb', 'usdt', x[5], x[11]],
#             ['bsc', 'pancake', 'bnb', 'usdt', x[6], x[12]],
#             ['bsc', 'biswap', 'btc', 'usdt', x[7], x[10]],
#             ['bsc', 'biswap', 'eth', 'usdt', x[8], x[9]],
#             ['bsc', 'pancake', 'cake', 'usdt', x[14], x[13]],
#             # ['polygon', 'quickswap', 'eth', 'usdc', x[16], x[21]],
#             # ['polygon', 'quickswap', 'eth', 'usdt', x[17], x[22]],
#             # ['polygon', 'quickswap', 'btc', 'usdc', x[18], x[23]],
#             # ['polygon', 'quickswap', 'matic', 'usdc', x[19], x[24]],
#             # ['polygon', 'quickswap', 'matic', 'usdt', x[20], x[25]],
#         ]
#
#         total = 0
#
#         for [chain, project, c1, c2, x1, x2] in input:
#             key = '{}_{}_{}_{}'.format(chain, project, c1, c2)
#             if get_apr(key) < 0.18:
#                 # print('apr too low:{}, {}', key, get_apr(key))
#                 continue
#
#             base = get_price(c1) * x1 + get_price(c2) * x2 + get_tvl(key)
#
#             if base <= 0.1:
#                 continue
#
#             r = get_reward(key) * (get_price(c1) * x1 + get_price(c2) * x2) / base
#             total += r
#
#         return total * -1
#
#     # bounds = (
#     #     (0, get_amount('bnb', 'bsc')),
#     #     (0, get_amount('bnb', 'bsc')),
#     #     (0, get_amount('usd', 'bsc')),
#     #     (0, get_amount('usd', 'bsc')),
#     #     (0, get_amount('usd', 'bsc')),
#     #     (0, get_amount('bnb', 'bsc')),
#     #     (0, get_amount('bnb', 'bsc')),
#     #     (0, get_amount('btc', 'bsc')),
#     #     (0, get_amount('eth', 'bsc')),
#     #     (0, get_amount('usdt', 'bsc')),
#     #     (0, get_amount('usdt', 'bsc')),
#     #     (0, get_amount('usdt', 'bsc')),
#     #     (0, get_amount('usdt', 'bsc')),
#     #     (0, get_amount('usdt', 'bsc')),
#     #     (0, get_amount('cake', 'bsc')),
#     #     (0, get_amount('cake', 'bsc')),
#     # )
#
#     cons = [
#         {'type': 'ineq', 'fun': lambda x: x[0]},
#         {'type': 'ineq', 'fun': lambda x: x[1]},
#         {'type': 'ineq', 'fun': lambda x: x[2]},
#         {'type': 'ineq', 'fun': lambda x: x[3]},
#         {'type': 'ineq', 'fun': lambda x: x[4]},
#         {'type': 'ineq', 'fun': lambda x: x[5]},
#         {'type': 'ineq', 'fun': lambda x: x[6]},
#         {'type': 'ineq', 'fun': lambda x: x[7]},
#         {'type': 'ineq', 'fun': lambda x: x[8]},
#         {'type': 'ineq', 'fun': lambda x: x[9]},
#         {'type': 'ineq', 'fun': lambda x: x[10]},
#         {'type': 'ineq', 'fun': lambda x: x[11]},
#         {'type': 'ineq', 'fun': lambda x: x[12]},
#         {'type': 'ineq', 'fun': lambda x: x[13]},
#         {'type': 'ineq', 'fun': lambda x: x[14]},
#         {'type': 'ineq', 'fun': lambda x: x[15]},
#         # {'type': 'ineq', 'fun': lambda x: x[16]},
#         # {'type': 'ineq', 'fun': lambda x: x[17]},
#         # {'type': 'ineq', 'fun': lambda x: x[18]},
#         # {'type': 'ineq', 'fun': lambda x: x[19]},
#         # {'type': 'ineq', 'fun': lambda x: x[20]},
#         # {'type': 'ineq', 'fun': lambda x: x[21]},
#         # {'type': 'ineq', 'fun': lambda x: x[22]},
#         # {'type': 'ineq', 'fun': lambda x: x[23]},
#         # {'type': 'ineq', 'fun': lambda x: x[24]},
#         # {'type': 'ineq', 'fun': lambda x: x[25]},
#
#         {'type': 'ineq', 'fun': lambda x: get_amount('btc', 'bsc') - x[7]},
#         {'type': 'ineq', 'fun': lambda x: get_amount('eth', 'bsc') - x[8]},
#         {'type': 'ineq', 'fun': lambda x: get_amount('cake', 'bsc') - x[14] - x[15]},
#         {'type': 'ineq', 'fun': lambda x: get_amount('bnb', 'bsc') - x[0] - x[1] - x[5] - x[6]},
#         {'type': 'ineq', 'fun': lambda x: get_amount('usd', 'bsc') - x[2] - x[3] - x[4]},
#         {'type': 'ineq', 'fun': lambda x: get_amount('usdt', 'bsc') - x[9] - x[10] - x[11] - x[12] - x[13]},
#
#         # {'type': 'ineq', 'fun': lambda x: get_amount('btc', 'polygon') - x[18]},
#         # {'type': 'ineq', 'fun': lambda x: get_amount('eth', 'polygon') - x[16] - x[17]},
#         # {'type': 'ineq', 'fun': lambda x: get_amount('matic', 'polygon') - x[19] - x[20]},
#         # {'type': 'ineq', 'fun': lambda x: get_amount('usdt', 'polygon') - x[22] - x[25]},
#         # {'type': 'ineq', 'fun': lambda x: get_amount('usdc', 'polygon') - x[23] - x[24]},
#
#         {'type': 'eq', 'fun': lambda x: x[0] * get_price('bnb') - x[3] * get_price('usd')},
#         {'type': 'eq', 'fun': lambda x: x[1] * get_price('bnb') - x[2] * get_price('usd')},
#         {'type': 'eq', 'fun': lambda x: x[15] * get_price('cake') - x[4] * get_price('usd')},
#         {'type': 'eq', 'fun': lambda x: x[5] * get_price('bnb') - x[11] * get_price('usdt')},
#         {'type': 'eq', 'fun': lambda x: x[6] * get_price('bnb') - x[12] * get_price('usdt')},
#         {'type': 'eq', 'fun': lambda x: x[7] * get_price('btc') - x[10] * get_price('usdt')},
#         {'type': 'eq', 'fun': lambda x: x[8] * get_price('eth') - x[9] * get_price('usdt')},
#         {'type': 'eq', 'fun': lambda x: x[14] * get_price('cake') - x[13] * get_price('usdt')},
#         # {'type': 'eq', 'fun': lambda x: x[16] * get_price('eth') - x[21] * get_price('usdc')},
#         # {'type': 'eq', 'fun': lambda x: x[17] * get_price('eth') - x[22] * get_price('usdt')},
#         # {'type': 'eq', 'fun': lambda x: x[18] * get_price('btc') - x[23] * get_price('usdc')},
#         # {'type': 'eq', 'fun': lambda x: x[19] * get_price('matic') - x[24] * get_price('usdc')},
#         # {'type': 'eq', 'fun': lambda x: x[20] * get_price('matic') - x[25] * get_price('usdt')},
#     ]
#
#     initx = [0] * 16
#
#     # print(initx)
#     res = minimize(function,
#                    x0=np.array(initx),
#                    #    bounds=bounds,
#                    constraints=cons,
#                    )
#     print("""
#     total info: btc:{} eth:{}, cake:{} bnb:{} usdt:{}
#     success: {}
#     message: {}
#     value: {}
#     price btc:{} eth:{} cake:{} bnb:{} usdt:{}
#     biswap tvl:{} apr:{} btc:{} usdt:{}
#     biswap tvl:{} apr:{} eth:{} usdt:{}
#     biswap tvl:{} apr:{} bnb:{} usdt:{}
#     pancake tvl:{} apr:{} bnb:{} usdt:{}
#     pancake tvl:{} apr:{} cake:{} usdt:{}
#     total invest btc:{} eth:{} cake:{} bnb:{} usdt:{}
#     diff btc:{} eth:{} cake:{} bnb:{} usdt:{}
#     """.format(
#         get_amount('btc', 'bsc'), get_amount('eth', 'bsc'), get_amount('cake', 'bsc'),
#         get_amount('bnb', 'bsc'),
#         get_amount('usdt', 'bsc'),
#         res.success,
#         res.message,
#         res.fun,
#         get_price('btc'), get_price('eth'), get_price('cake'), get_price('bnb'), get_price('usdt'),
#         tvl_dict['bsc_biswap_btc_usdt'], apr_dict['bsc_biswap_btc_usdt'], res.x[7], res.x[10],
#         tvl_dict['bsc_biswap_eth_usdt'], apr_dict['bsc_biswap_eth_usdt'], res.x[8], res.x[9],
#         tvl_dict['bsc_biswap_bnb_usdt'], apr_dict['bsc_biswap_bnb_usdt'], res.x[5], res.x[11],
#         tvl_dict['bsc_pancake_bnb_usdt'], apr_dict['bsc_pancake_bnb_usdt'], res.x[6], res.x[12],
#         tvl_dict['bsc_pancake_cake_usdt'], apr_dict['bsc_pancake_cake_usdt'], res.x[14], res.x[13],
#         res.x[7], res.x[8], res.x[14], res.x[5] + res.x[6],
#                                        res.x[9] + res.x[10] + res.x[11] + res.x[12] + res.x[13],
#                                        get_amount('btc', 'bsc') - res.x[7],
#                                        get_amount('eth', 'bsc') - res.x[8],
#                                        get_amount('cake', 'bsc') - res.x[14],
#                                        get_amount('bnb', 'bsc') - res.x[5] - res.x[6],
#                                        get_amount('usdt', 'bsc') - (
#                                                res.x[9] + res.x[10] + res.x[11] + res.x[12] + res.x[13])))
#     print(res.x)


def calc_invest_bsc(balance_info_dict, price_dict, daily_reward_dict, apr_dict, tvl_dict):
    # 确定最终的资产分布情况
    # X(0) = bnb_amt_biswap_busd
    # X(3) = busd_amt_biswap_bnb

    # X(1) = bnb_amt_pancake_busd
    # X(2) = busd_amt_pancake_bnb

    # X(15) = cake_amt_pancake_busd
    # X(4) = busd_amt_pancake_cake

    # X(5) = bnb_amt_biswap_usdt
    # X(11) = usdt_amt_biswap_bnb

    # X(6) = bnb_amt_pancake_usdt
    # X(12) = usdt_amt_pancake_bnb

    # X(7) = btcb_amt_biswap_usdt
    # X(10) = usdt_amt_biswap_btcb

    # X(8) = eth_amt_biswap_usdt
    # X(9) = usdt_amt_biswap_eth

    # X(14) = cake_amt_pancake_usdt
    # X(13) = usdt_amt_pancake_cake

    def get_amount(currency, chain):
        if currency not in balance_info_dict:
            return 0.00001

        if chain not in balance_info_dict[currency]:
            return 0.00001

        if 'amount' not in balance_info_dict[currency][chain]:
            return 0.00001

        return float(balance_info_dict[currency][chain]['amount'])

    def get_price(currency):
        if currency not in price_dict:
            return 0

        return float(price_dict[currency])

    def get_reward(key):
        if key not in daily_reward_dict:
            return 0

        return float(daily_reward_dict[key])

    def get_apr(key):
        if key not in apr_dict:
            return 0

        return float(apr_dict[key])

    def get_tvl(key):
        if key not in tvl_dict:
            return 0

        return float(tvl_dict[key])

    def function(x):
        input = [
            ['bsc', 'biswap', 'bnb', 'usd', x[0], x[3]],
            ['bsc', 'pancake', 'bnb', 'usd', x[1], x[2]],
            ['bsc', 'pancake', 'cake', 'usd', x[15], x[4]],
            ['bsc', 'biswap', 'bnb', 'usdt', x[5], x[11]],
            ['bsc', 'pancake', 'bnb', 'usdt', x[6], x[12]],
            ['bsc', 'biswap', 'btc', 'usdt', x[7], x[10]],
            ['bsc', 'biswap', 'eth', 'usdt', x[8], x[9]],
            ['bsc', 'pancake', 'cake', 'usdt', x[14], x[13]],
        ]

        total = 0

        for [chain, project, c1, c2, x1, x2] in input:
            key = '{}_{}_{}_{}'.format(chain, project, c1, c2)
            if get_apr(key) < 0.18:
                # print('apr too low:{}, {}', key, get_apr(key))
                continue

            base = get_price(c1) * x1 + get_price(c2) * x2 + get_tvl(key)

            # if base <= 0.01:
            #     continue

            r = get_reward(key) * (get_price(c1) * x1 + get_price(c2) * x2) / base
            total += r

        return total * -1

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
    )

    cons = [
        # {'type': 'ineq', 'fun': lambda x: x[0]},
        # {'type': 'ineq', 'fun': lambda x: x[1]},
        # {'type': 'ineq', 'fun': lambda x: x[2]},
        # {'type': 'ineq', 'fun': lambda x: x[3]},
        # {'type': 'ineq', 'fun': lambda x: x[4]},
        # {'type': 'ineq', 'fun': lambda x: x[5]},
        # {'type': 'ineq', 'fun': lambda x: x[6]},
        # {'type': 'ineq', 'fun': lambda x: x[7]},
        # {'type': 'ineq', 'fun': lambda x: x[8]},
        # {'type': 'ineq', 'fun': lambda x: x[9]},
        # {'type': 'ineq', 'fun': lambda x: x[10]},
        # {'type': 'ineq', 'fun': lambda x: x[11]},
        # {'type': 'ineq', 'fun': lambda x: x[12]},
        # {'type': 'ineq', 'fun': lambda x: x[13]},
        # {'type': 'ineq', 'fun': lambda x: x[14]},
        # {'type': 'ineq', 'fun': lambda x: x[15]},
        #
        #
        {'type': 'ineq', 'fun': lambda x: get_amount('btc', 'bsc') - x[7]},
        {'type': 'ineq', 'fun': lambda x: get_amount('eth', 'bsc') - x[8]},
        {'type': 'ineq', 'fun': lambda x: get_amount('cake', 'bsc') - x[14] - x[15]},
        {'type': 'ineq', 'fun': lambda x: get_amount('bnb', 'bsc') - x[0] - x[1] - x[5] - x[6]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usd', 'bsc') - x[2] - x[3] - x[4]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usdt', 'bsc') - x[9] - x[10] - x[11] - x[12] - x[13]},

        # {'type': 'eq', 'fun': lambda x: x[0] * get_price('bnb') - x[3] * get_price('usd')},
        # {'type': 'eq', 'fun': lambda x: x[1] * get_price('bnb') - x[2] * get_price('usd')},
        {'type': 'eq', 'fun': lambda x: x[15] * get_price('cake') - x[4] * get_price('usd')},
        # {'type': 'eq', 'fun': lambda x: x[5] * get_price('bnb') - x[11] * get_price('usdt')},
        # {'type': 'eq', 'fun': lambda x: x[6] * get_price('bnb') - x[12] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[7] * get_price('btc') - x[10] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[8] * get_price('eth') - x[9] * get_price('usdt')},
        # {'type': 'eq', 'fun': lambda x: x[14] * get_price('cake') - x[13] * get_price('usdt')},
    ]

    initx = [0] * 16

    # print(initx)
    res = minimize(function,
                   x0=np.array(initx),
                   bounds=bounds,
                   constraints=cons,
                   )
    print("""
    total info: btc:{} eth:{}, cake:{} bnb:{} usdt:{}
    success: {}
    message: {}
    value: {}
    price btc:{} eth:{} cake:{} bnb:{} usdt:{}
    biswap tvl:{} apr:{} btc:{} usdt:{}
    biswap tvl:{} apr:{} eth:{} usdt:{}
    biswap tvl:{} apr:{} bnb:{} usdt:{}
    pancake tvl:{} apr:{} bnb:{} usdt:{}
    pancake tvl:{} apr:{} cake:{} usdt:{}
    total invest btc:{} eth:{} cake:{} bnb:{} usdt:{}
    diff btc:{} eth:{} cake:{} bnb:{} usdt:{}
    """.format(
        get_amount('btc', 'bsc'), get_amount('eth', 'bsc'), get_amount('cake', 'bsc'),
        get_amount('bnb', 'bsc'),
        get_amount('usdt', 'bsc'),
        res.success,
        res.message,
        res.fun,
        get_price('btc'), get_price('eth'), get_price('cake'), get_price('bnb'), get_price('usdt'),
        tvl_dict['bsc_biswap_btc_usdt'], apr_dict['bsc_biswap_btc_usdt'], res.x[7], res.x[10],
        tvl_dict['bsc_biswap_eth_usdt'], apr_dict['bsc_biswap_eth_usdt'], res.x[8], res.x[9],
        tvl_dict['bsc_biswap_bnb_usdt'], apr_dict['bsc_biswap_bnb_usdt'], res.x[5], res.x[11],
        tvl_dict['bsc_pancake_bnb_usdt'], apr_dict['bsc_pancake_bnb_usdt'], res.x[6], res.x[12],
        tvl_dict['bsc_pancake_cake_usdt'], apr_dict['bsc_pancake_cake_usdt'], res.x[14], res.x[13],
        res.x[7], res.x[8], res.x[14], res.x[5] + res.x[6],
                                       res.x[9] + res.x[10] + res.x[11] + res.x[12] + res.x[13],
                                       get_amount('btc', 'bsc') - res.x[7],
                                       get_amount('eth', 'bsc') - res.x[8],
                                       get_amount('cake', 'bsc') - res.x[14],
                                       get_amount('bnb', 'bsc') - res.x[5] - res.x[6],
                                       get_amount('usdt', 'bsc') - (
                                               res.x[9] + res.x[10] + res.x[11] + res.x[12] + res.x[13])))
    print(res.x)


# def calc_invest(chain, balance_info_dict, price_dict, daily_reward_dict, apr_dict, tvl_dict):
#     # 项目列表
#     lps = apr_dict.items()
#     sorted(apr_dict.items(), key=lambda item: item[1], reverse=True)
#
#     pass


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
