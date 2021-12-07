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
            'amount': Decimal(1000),
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
            'amount': Decimal(2000),
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
        return

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

    def generate_key(chain, project, currencies):
        currencies = list(filter(lambda x: x is not None, currencies))
        currencies.sort(key=len, reverse=True)
        return "{}_{}_{}".format(chain, project, '_'.join(currencies))

    for p in projects:
        info = p.get_info()

        for pool in info:
            for token in pool['rewardTokenList'] + pool['depositTokenList']:
                currency = find_currency_by_address(session, format_addr(token['tokenAddress']))
                if currency is not None and currency not in price:
                    price[currency] = Decimal(token['tokenPrice'])

                names = [format_token_name(currencyName, c)[1] for c in pool['poolName'].split("/")]
                key = generate_key(p.chain, p.name, names)

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

                key = generate_key(s.chain, s.project, [s.currency0, s.currency1])
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

    calc_invest(session, 'bsc', beforeInfo, price, dailyReward, apr, tvl)


def get_info_by_strategy_str(lp):
    data = lp.split('_')
    if len(data) < 3:
        return None, None
    elif len(data) == 3:
        return data, data[2], None
    else:
        return data, data[2], data[3]


def calc_invest(session, chain, balance_info_dict, price_dict, daily_reward_dict, apr_dict, tvl_dict):
    # 项目列表

    def f(key):
        infos = get_info_by_strategy_str(key)
        strategies = [s for s in
                      find_strategies_by_chain_project_and_currencies(session, chain, infos[0][1], infos[1], infos[2])]
        return len(strategies) > 0

    invest_calc_result = {}

    detla = Decimal(0.005)

    while True:
        lpKeys = [k for k in sorted(apr_dict, key=apr_dict.get, reverse=True)]
        lpKeys = list(filter(f, lpKeys))
        if len(lpKeys) <= 0:
            break

        print(lpKeys)
        # 找到top1 与top2
        top = []
        apr1 = apr_dict[lpKeys[0]]
        aprTarget = Decimal(0.01)
        for key in lpKeys:

            if abs(apr_dict[key] - apr1) < detla:
                top.append(key)
            else:
                aprTarget = apr_dict[key]
                break

        for key in top:
            filled, vol, changes = fill_cap(chain, key, daily_reward_dict, tvl_dict, balance_info_dict, price_dict,
                                            max(aprTarget, apr1 - detla))

            if key not in invest_calc_result:
                invest_calc_result[key] = {}

            for k, v in changes.items():
                if k not in invest_calc_result[key]:
                    invest_calc_result[key][k] = Decimal(0)

                invest_calc_result[key][k] += v * -1
                balance_info_dict[k][chain]['amount'] += v

            tvl_dict[key] += vol
            apr_dict[key] = daily_reward_dict[key] * 365 / tvl_dict[key]

            # 说明没有对应资产了
            if not filled:
                lpKeys.remove(key)
                apr_dict.pop(key)

    print('invest info:{}'.format(json.dumps(invest_calc_result, cls=utils.DecimalEncoder)))
    # invest_items = []
    # #生成配资结果
    # for k, balances in invest_calc_result.items():
    #     infos = get_info_by_strategy_str(key)
    #     strategies = [s for s in
    #                   find_strategies_by_chain_project_and_currencies(session, chain, infos[0][1], infos[1], infos[2])]
    #
    #     strategy = strategies[0]
    #     data = {}
    #     for currency, balance in balances.items():
    #         pass


def get_balance(balance_dict, currency, chain):
    if currency not in balance_dict:
        return Decimal(0)
    if chain not in balance_dict[currency]:
        return Decimal(0)
    return balance_dict[currency][chain]['amount']


def get_price(price_dict, currency):
    if currency not in price_dict:
        return Decimal(0)

    return price_dict[currency]


# 返回值 1. 是否填满了，2 填充量是多少 3.资产余额变化
def fill_cap(chain, strategy, daily_reward_dict, tvl_dict, balance_dict, price_dict, target_apr):
    cap = (daily_reward_dict[strategy] * Decimal(365) / target_apr - tvl_dict[strategy])
    data, c0, c1 = get_info_by_strategy_str(strategy)
    if c0 is None:
        return False, 0, {}

    # 单币
    if c1 is None:
        if get_price(price_dict, c0) == Decimal(0):
            return False, 0, {}

        vol = min(cap / get_price(price_dict, c0), get_balance(balance_dict, c0, chain))
        if vol <= 0:
            return False, 0, {}

        return cap == vol, vol, {c0: -1 * vol}

    # 双币
    vol = min(get_balance(balance_dict, c0, chain) * get_price(price_dict, c0),
              get_balance(balance_dict, c1, chain) * get_price(price_dict, c1), cap / 2)

    if vol <= 0:
        return False, 0, {}

    amount0 = vol / get_price(price_dict, c0)
    amount1 = vol / get_price(price_dict, c1)
    if amount0 > 0 and amount1 > 0:
        return cap == vol * 2, vol * 2, {c0: - amount0, c1: -amount1}

    return False, 0, {}


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
