# -*- coding:utf-8 -*-
import requests
import json
import time
import pymysql
import yaml
import pickle
import numpy as np


class ChainParams:
    pass


class ProjectParams:
    pass


class CurrencyParams:
    pass


class StategyParams:
    pass


class CurrencyAmount:
    pass


class Token:
    pass


class Pool:
    pass


class CrossBalanceItem:
    pass


class ReceiveFromBridgeParam:
    pass


class InvestParam:
    pass


class SendToBridgeParam:
    pass


class Params:
    pass

#api暂时不可用，测试结构体
class PoolInfo:
    pass

class contract_info:
    pass

# todo:1.放入config中 2.solo怎么考虑？如果作为key 三个链都有这个项目？HSOLO BSOLO PSOLO？
# 项目和链的对应关系
chain_infos = {"pancake": "BSC", "biswap": "BSC", "quickswap": "POLY"}


def getprojectinfo(url):
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    e = json.loads(string)
    reward = 0
    prices = {}
    tvls = {}
    aprs = {}
    for data in e["data"]:  #这里要指定project和交易对的dailyreward
        tvls[data["poolName"]] = data["tvl"]
        aprs[data["poolName"]] = data["apr"]
        for rewardToken in data["rewardTokenList"]:
            prices[rewardToken["tokenSymbol"]] = rewardToken["tokenPrice"]
            dailyReward = float(rewardToken["dayAmount"]) * float(rewardToken["tokenPrice"])
            reward = reward + dailyReward
    print("totalReward is")
    print(reward)
    print("prices dict is")
    for token in prices:
        print(token + ':' + prices[token])
    print("tvls dict is")
    for poolName in tvls:
        print(poolName + ':' + tvls[poolName])
    print("aprs dict is")
    for poolName in aprs:
        print(poolName + ':' + aprs[poolName])

    return prices, reward, tvls, aprs


def getpoolinfo(url):
    # 存储从api获取的poolinfo
    pool_infos = {}
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    e = json.loads(string)
    for pool_info in e["data"]["pools_info"]:
        pool = Pool()
        pool_infos[pool_info["symbol"]] = pool_info
    return pool_infos


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


def getPairinfo(X):
    # 存储配资计算的交易对数量结果
    currency_info = {}
    token = Token()
    currency = CurrencyAmount()

    token.amount = X[0][0]  # X(0)
    token.name = "BNB"
    currency.base = token
    token.amount = X[0][3]  # X(3)
    token.name = "BUSD"
    currency.counter = token
    currency_info["biswap"] = currency

    token.amount = X[0][1]  # X(1)
    token.name = "BNB"
    currency.base = token
    token.amount = X[0][2]  # X(2)
    token.name = "BUSD"
    currency.counter = token
    currency_info["pancake"] = currency

    token.amount = X[1][0]  # X(4)
    token.name = "BUSD"
    currency.base = token
    token.amount = X[3][3]  # X(15)
    token.name = "CAKE"
    currency.counter = token
    currency_info["biswap"] = currency

    token.amount = X[1][1]  # X(5)
    token.name = "BNB"
    currency.base = token
    token.amount = X[2][3]  # X(11)
    token.name = "USDT"
    currency.counter = token
    currency_info["biswap"] = currency

    token.amount = X[1][2]  # X(6)
    token.name = "BNB"
    currency.base = token
    token.amount = X[3][0]  # X(12)：
    token.name = "USDT"
    currency.counter = token
    currency_info["biswap"] = currency

    token.amount = X[1][3]  # X(7)
    token.name = "BTCB"
    currency.base = token
    token.amount = X[2][2]  # X(10)
    token.name = "USDT"
    currency.counter = token
    currency_info["biswap"] = currency

    token.amount = X[2][0]  # X(8)
    token.name = "ETH"
    currency.base = token
    token.amount = X[2][1]  # X(9)
    token.name = "USDT"
    currency.counter = token
    currency_info["biswap"] = currency

    token.amount = X[3][1]  # X(13)
    token.name = "USDT"
    currency.base = token
    token.amount = X[3][2]  # X(14)
    token.name = "CAKE"
    currency.counter = token
    currency_info["pancake"] = currency

    return currency_info


def getReParams(currency_infos, pool_infos, btc_bsc):
    crossBalance = CrossBalanceItem()
    crossBalance.FromChain = "heco"
    crossBalance.ToChain = "bsc"
    crossBalance.FromAddr = "configaddress1"  # 配置-从config读取
    crossBalance.ToAddr = "configaddress2"  # 配置的签名机地址
    crossBalance.FromCurrency = "hbtc"  # 配置
    crossBalance.ToCurrency = "btc"  # 配置

    symbol = "HBTC"
    info = pool_infos[symbol]
    # 这里从配资结果得到：需要跨的数量减去已经在bsc上但是未投出去的数量，即：crossed_quantity_in_bsc_controller
    # 问题 1.配资计算返回btc_bsc,eth_bsc,usdt_bsc，本次如何知道用那个作为减数？就是跨btc/etc/usdt？
    crossBalance.Amount = btc_bsc - info["crossed_quantity_in_bsc_controller"]

    receiveFromBridge = ReceiveFromBridgeParam()
    receiveFromBridge.ChainID = 52  # 配置
    receiveFromBridge.ChainName = "bsc"  # 配置
    receiveFromBridge.From = "configaddress2"  # 配置的签名机地址
    receiveFromBridge.To = "configaddress3"  # 配置的合约地址
    receiveFromBridge.Erc20ContractAddr = "configaddress4"  # 配置的token地址

    # 问题 2.这里的amount和上面的crossBalance amount是相等么？
    receiveFromBridge.Amount = crossBalance.Amount * 10e18  # 精度配置读取
    # 生成全局唯一的task🆔并保存币种和taskID的对应关系
    TaskIds = {}
    t = time.time()
    receiveFromBridge.TaskID = int(round(t * 1000))  # 毫秒级时间戳
    TaskIds["BTC"] = receiveFromBridge.TaskID

    invest = InvestParam()
    invest.ChinId = 52  # 配置
    invest.ChainName = "bsc"  # 配置
    invest.From = "configaddress2"  # 配置的签名机地址
    invest.To = "configaddress3"  # 配置的合约地址 ----这个应该是contract_info中对应链的vault地址

    invest.StrategyAddresses = []
    invest.BaseTokenAmount = []
    invest.CounterTokenAmount = []

    # 拼接策略:从api返回结果中找到对应地址 拼接规则：chain + "_" + project + "strategy"
    # 遍历8个交易对 currency_infos中的key是project名字 value是交易对
    for key in currency_infos:
        # todo：chain_infos中不存在key对应的project的处理
        chain = chain_infos[key]
        strategystr = chain + "_" + key + "strategy"
        # todo：api返回对应币种的contract_info不存在strategystr的处理
        strategyAddresses = info["contract_info"][strategystr]
        baseTokenAmount = currency_infos[key].base.amount
        counterTokenAmount = currency_infos[key].counter.amount

        invest.StrategyAddresses.append(strategyAddresses)
        invest.BaseTokenAmount.append(baseTokenAmount)
        invest.CounterTokenAmount.append(counterTokenAmount)

    sendToBridge = SendToBridgeParam()

    sendToBridge.ChainId = 52
    sendToBridge.ChainName = "bsc"
    sendToBridge.From = "configaddress2"  # 配置的签名机地址
    sendToBridge.To = "configaddress3"  # 配置的合约地址
    sendToBridge.BridgeAddress = ""  # 配置的地址
    sendToBridge.Amount = 1 * 10e18  # 精度配置读取
    sendToBridge.TaskID = TaskIds["BTC"]

    params = Params()
    params.CrossBalances = crossBalance
    params.ReceiveFromBridgeParams = receiveFromBridge
    params.InvestParams = invest
    params.SendToBridgeParams = sendToBridge

    ret = pickle.dumps(params)
    return ret


if __name__ == '__main__':
    # 读取config
    conf = read_yaml("../config.yaml")

    # 获取project info
    pancakeUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    pancakeinfos = getprojectinfo(pancakeUrl)

    biswapUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=476'
    biswapinfos = getprojectinfo(biswapUrl)

    soloUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    soloinfos = getprojectinfo(soloUrl)

    # 获取pool info
    #pools_url = ''
    #pool_infos = getpoolinfo(pools_url)
    #由于API暂时不可用，造测试数据


    # 配资计算
    btc_bsc = 100
    eth_bsc = 100
    usdt_bsc = 100
    X = np.random.randint(1, 100, (4, 4))
    # 交易对赋值
    currency_infos = getPairinfo(X)

    # 拼接结果字串
    parambytes = getReParams(currency_infos, pool_infos, btc_bsc)

    # write db
    connect = getconnectinfo(conf["database"]["db"])
    print(connect)
    conn = pymysql.connect(host='127.0.0.1', port=3306, user='root', passwd='csquan253905', db='reblance',
                           charset='utf8')
    print(conn)

    # cursor = db.cursor()

    # cursor.execute('''insert into Rebalance_params values()''')

    # cursor.close()
    # db.commit()
    conn.close()
