# -*- coding:utf-8 -*-
import requests
import json
import time
import pymysql
import yaml
import pickle
import numpy as np
import json


class ChainParams:
    pass


class ProjectParams:
    pass


class CurrencyParams:
    pass


class StategyParams:
    pass


class Currency:
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


# api暂时不可用，测试结构体
class PoolInfo:
    pass


class ContractInfo:
    pass


class Pair:
    pass


# todo:1.放入config中 2.solo怎么考虑？如果作为key 三个链都有这个项目？HSOLO BSOLO PSOLO？
# 项目和链的对应关系
chain_infos = {"pancake": "bsc", "biswap": "bsc", "quickswap": "poly", "hsolo": "heco", "bsolo": "bsc", "psolo": "poly"}

# 每个链上的base token
base_tokens = ["ht", "bnb", "matic"]


def getPair(str):
    pair = Pair()
    tokenstr = str.split('/')  # 用/分割str字符串,etc:Cake/WBNB
    print(tokenstr)

    str1 = tokenstr[0].lower()
    str2 = tokenstr[1].lower()

    for base in base_tokens:
        if str1.find(base):
            pair.base = str1
            pair.counter = str2
        if str2.find(base):
            pair.base = str1
            pair.counter = str2
    return pair


# 关于价格：函数获取价格填写进传入的currencys，同时将这个价格和对应币种返回
def getprojectinfo(project, url, currencys):
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    e = json.loads(string)
    reward = 0
    tvls = {}
    aprs = {}
    daily = {}
    for data in e["data"]:
        tvls[data["poolName"]] = data["tvl"]
        aprs[data["poolName"]] = data["apr"]
        for rewardToken in data["rewardTokenList"]:
            # 拼接dailyReward
            tokenPair = getPair(data["poolName"])
            key = tokenPair.base + '_' + tokenPair.counter + '_' + project
            dailyReward = float(rewardToken["dayAmount"]) * float(rewardToken["tokenPrice"])
            daily[key] = dailyReward
            reward = reward + dailyReward
        for deposit in data["depositTokenList"]:
            #首先以tokenAddress到config中查找，获取对应币种的名字
            for name in currencys:
                if currencys[name]["address"] == deposit["tokenAddress"]:
                    currencys[name]["price"] = deposit["tokenPrice"]

    print("totalReward is")
    print(reward)

    print("daily dict is")
    for d in daily:
        print(d + ':' + str(daily[d]))
    print("tvls dict is")
    for poolName in tvls:
        print(poolName + ':' + tvls[poolName])
    print("aprs dict is")
    for poolName in aprs:
        print(poolName + ':' + aprs[poolName])
    print("currencys price dict is")
    for currency in currencys:
        print(currency + ':' + str(currencys[currency]))

    return reward, daily, tvls, aprs, currencys


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
    # key: base + counter + project
    currency_info = {}

    token1 = Token()
    currency1 = Currency()

    token1.amount = X[0][0]  # X(0)
    token1.name = "bnb"
    currency1.base = token1

    token2 = Token()

    token2.amount = X[0][3]  # X(3)
    token2.name = "busd"
    currency1.counter = token2

    currency_info["bnb_busd_biswap"] = currency1

    token3 = Token()
    currency2 = Currency()

    token3.amount = X[0][1]  # X(1)
    token3.name = "bnb"
    currency2.base = token3

    token4 = Token()
    token4.amount = X[0][2]  # X(2)
    token4.name = "busd"
    currency2.counter = token4

    currency_info["bnb_busd_pancake"] = currency2

    token5 = Token()
    currency3 = Currency()

    token5.amount = X[3][3]  # X(15)
    token5.name = "cake"
    currency3.base = token5

    token6 = Token()

    token6.amount = X[1][0]  # X(4)
    token6.name = "busd"
    currency3.counter = token6
    currency_info["cake_busd_biswap"] = currency3

    token7 = Token()
    currency4 = Currency()

    token7.amount = X[1][1]  # X(5)
    token7.name = "bnb"
    currency4.base = token7

    token8 = Token()

    token8.amount = X[2][3]  # X(11)
    token8.name = "usdt"
    currency4.counter = token8
    currency_info["bnb_usdt_biswap"] = currency4

    token9 = Token()
    currency5 = Currency()

    token9.amount = X[1][2]  # X(6)
    token9.name = "bnb"
    currency5.base = token9

    token10 = Token()

    token10.amount = X[3][0]  # X(12)：
    token10.name = "usdt"
    currency5.counter = token10
    currency_info["bnb_usdt_pancake"] = currency5

    token11 = Token()
    currency6 = Currency()

    token11.amount = X[1][3]  # X(7)
    token11.name = "btcb"
    currency6.base = token11

    token12 = Token()

    token12.amount = X[2][2]  # X(10)
    token12.name = "usdt"
    currency6.counter = token12

    currency_info["btcb_usdt_biswap"] = currency6

    token13 = Token()
    currency7 = Currency()

    token13.amount = X[2][0]  # X(8)
    token13.name = "eth"
    currency7.base = token13

    token14 = Token()

    token14.amount = X[2][1]  # X(9)
    token14.name = "usdt"
    currency7.counter = token14

    currency_info["eth_usdt_biswap"] = currency7

    token15 = Token()
    currency8 = Currency()

    token15.amount = X[3][2]  # X(14)
    token15.name = "cake"
    currency8.base = token15

    token16 = Token()

    token16.amount = X[3][1]  # X(13)
    token16.name = "usdt"
    currency8.counter = token16

    currency_info["cake_usdt_pancake"] = currency8

    for key in currency_info:
        print(key + ': base name ' + currency_info[key].base.name + " amount " + str(
            currency_info[key].base.amount) + ' and counter name ' + currency_info[key].counter.name + " amount " + str(
            currency_info[key].counter.amount))

    return currency_info


def getProject(str):
    startpos = str.rindex("_")
    return str[startpos + 1:len(str)]


def obj_2_json(obj):
    return {
        "heco_vault": obj.heco_vault,
        "heco_solostrategy": obj.heco_solostrategy,
        "bsc_vault": obj.bsc_vault,
        "bsc_solostrategy": obj.bsc_solostrategy,
        "bsc_biswapstrategy": obj.bsc_biswapstrategy,
        "bsc_pancakestrategy": obj.bsc_pancakestrategy,
        "poly_vault": obj.poly_vault,
        "poly_solostrategy": obj.poly_solostrategy,
        "poly_quickswapstrategy": obj.poly_quickswapstrategy
    }


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
    # 问题 1.配资计算返回btc_bsc,eth_bsc,usdt_bsc，本次如何知道用哪一个作为减数？
    crossBalance.Amount = btc_bsc - info.crossed_quantity_in_bsc_controller

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
        project = getProject(key)
        chain = chain_infos[project]
        strategystr = chain + "_" + project + "strategy"
        # todo：api返回对应币种的contract_info不存在strategystr的处理
        # 下面的info实际应该根据币种到pool_infos中查找,这里测试 就是固定的一个值
        contract = info.contract_info
        # 将cntractjson序列化，根据键值查找
        str = json.dumps(contract, default=obj_2_json)
        jsons = json.loads(str)
        strategyAddresses = jsons[strategystr]
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

    currency_dict = conf.get("currency")

    # 获取project info
    pancakeUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    pancakeinfos = getprojectinfo("pancake", pancakeUrl, currency_dict)

    biswapUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=476'
    biswapinfos = getprojectinfo("biswap", biswapUrl, currency_dict)

    soloUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    soloinfos = getprojectinfo("solo", soloUrl, currency_dict)

    # 获取pool info
    # pools_url = ''
    # pool_infos = getpoolinfo(pools_url)
    # 由于API暂时不可用，造测试数据
    contractinfo = ContractInfo()
    contractinfo.heco_vault = "0x1",
    contractinfo.heco_solostrategy = "0x2",
    contractinfo.bsc_vault = "0x3",
    contractinfo.bsc_solostrategy = "0x4",
    contractinfo.bsc_biswapstrategy = "0x5",
    contractinfo.bsc_pancakestrategy = "0x6",
    contractinfo.poly_vault = "0x7",
    contractinfo.poly_solostrategy = "0x8",
    contractinfo.poly_quickswapstrategy = "0x9"

    poolinfo = PoolInfo()
    poolinfo.chain = "heco"
    poolinfo.chain_id = 50
    poolinfo.symbol = "HBTC"
    poolinfo.decimal = 18
    poolinfo.heco_uncross_quantity = 1000002
    poolinfo.crossed_quantity_in_bsc_controller = 2
    poolinfo.crossed_quantity_in_poly_controller = 2
    poolinfo.bsc_vault_unre_qunatity = 0
    poolinfo.poly_vault_unre_qunatity = 0
    poolinfo.contract_info = contractinfo

    pool_infos = {}
    pool_infos["HBTC"] = poolinfo
    # 造测试数据结束

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
