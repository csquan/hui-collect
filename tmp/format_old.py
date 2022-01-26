# -*- coding:utf-8 -*-
import requests
import yaml
import json
import utils

from functools import reduce
from decimal import *
from sqlalchemy.orm import sessionmaker
from orm import *
from invest_with_opt import *

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
            'amount': Decimal(89378),
        },
    }
    beforeInfo['usd'] = {
        'bsc': {
            'amount': Decimal(34774),
        },
    }
    beforeInfo['eth'] = {
        'bsc': {
            'amount': Decimal(0),
        },
    }
    beforeInfo['btc'] = {
        'bsc': {
            'amount': Decimal(1),
        },
    }

    beforeInfo['cake'] = {
        'bsc': {
            'amount': Decimal(1000),
        }
    }

    beforeInfo['bnb'] = {
        'bsc': {
            'amount': Decimal(100),
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

    # # 没超过阈值
    # if not needReBalance:
    #     exit()

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

    calc_invest(None, 'bsc', beforeInfo, price, dailyReward, apr, tvl)


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
