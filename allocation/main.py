# -*- coding:utf-8 -*-
import requests
import json
import time
import pymysql
import yaml

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


#读取config
conf = read_yaml("../config.yaml")

def getprojectinfo(url):
    ret = requests.get(url)
    string = str(ret.content,'utf-8')
    e = json.loads(string)
    reward = 0
    prices = {}
    tvls = {}
    aprs = {}
    for data in e["data"]:
        tvls[data["poolName"]] = data["tvl"]
        aprs[data["poolName"]] = data["apr"]
        for rewardToken in data["rewardTokenList"]:
            prices[rewardToken["tokenSymbol"]] = rewardToken["tokenPrice"]
            dailyReward = float(rewardToken["dayAmount"])*float(rewardToken["tokenPrice"])
            reward = reward + dailyReward
    print("totalReward is")
    print(reward)
    print("prices dict is")
    for token in prices:
        print(token+':'+prices[token])
    print("tvls dict is")
    for poolName in tvls:
        print(poolName + ':' + tvls[poolName])
    print("aprs dict is")
    for poolName in aprs:
        print(poolName + ':' + aprs[poolName])

    return prices, reward, tvls, aprs


def getpoolinfo(url):
    ret = requests.get(url)
    string = str(ret.content,'utf-8')
    e = json.loads(string)
    pool_infos = {}
    for pool_info in e["data"]["pools_info"]:
        pool = Pool()
        pool[pool_info["symbol"]] = pool_info
    return pool_infos

def read_yaml(path):
    with open(path, 'r', encoding='utf8') as f:
        return yaml.safe_load(f.read())

def getconnectinfo(connstr):
    strlist = connstr.split('@')  # 用逗号分割str字符串，并保存到列表
    print(strlist)
    str1 = strlist[0]             # 包含用户名密码的字串
    str2 = strlist[1]             # 包含Ip端口数据库的字串

    user_endpos = str1.index(":")
    user = str1[0:user_endpos]
    password = str1[user_endpos+1:len(str1)]

    host_startpos = str2.index("(") + 1
    host_endpos = str2.index(":")

    host = str2[host_startpos:host_endpos]
    port_endpos  = str2.index(")")
    port = str2[host_endpos + 1:port_endpos]

    db_startpos = str2.index("/")
    db_endpos = str2.index("?")

    db = str2[db_startpos + 1:db_endpos]

    return user, password, host, port, db


if __name__ == '__main__':
    #获取project info
    print("+++++pancake")
    pancakeUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    pancakeinfos = getprojectinfo(pancakeUrl)

    print("+++++biswap")
    biswapUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=476'
    biswapinfos = getprojectinfo(biswapUrl)

    print("+++++solo")
    soloUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    soloinfos = getprojectinfo(soloUrl)

    #配资计算


    #获取pool info
    pool_infos = {}
    #infoUrl = ''
    #pool_infos = getpoolinfo(infoUrl)

    #测试数据
    pool_infos["HBTC"] = {}

    #组装结果参数
    crossBalance = CrossBalanceItem()
    crossBalance.FromChain = "heco"
    crossBalance.ToChain = "bsc" #这里从配资获得,是多个跨链，应该分别处理，这里假设heco向bsc跨链
    crossBalance.FromAddr = "configaddress1" #配置-从config读取
    crossBalance.ToAddr = "configaddress2"   #配置的签名机地址
    crossBalance.FromCurrency = "hbtc" #配置
    crossBalance.ToCurrency = "btc"   #配置

    symbol = "HBTC"
    info = pool_infos[symbol]
    crossBalance.Amount = 0  #这里从配资结果得到：需要跨的数量减去已经在bsc上但是未投出去的数量，即：crossed_quantity_in_bsc_controller

    receiveFromBridge = ReceiveFromBridgeParam()
    receiveFromBridge.ChainID = 52       #配置
    receiveFromBridge.ChainName = "bsc"  #配置
    receiveFromBridge.From = "configaddress2"   #配置的签名机地址
    receiveFromBridge.To = "configaddress3"  # 配置的合约地址
    receiveFromBridge.Erc20ContractAddr = "configaddress4"  # 配置的token地址

    receiveFromBridge.Amount = 1*10e18  #精度配置读取
    #生成全局唯一的task🆔并保存币种和taskID的对应关系
    TaskIds = {}
    t = time.time()
    receiveFromBridge.TaskID = int(round(t * 1000)) #毫秒级时间戳
    TaskIds["BTC"] = receiveFromBridge.TaskID

    invest = InvestParam()
    invest.ChinId = 52         #配置
    invest.ChainName = "bsc"   #配置
    invest.From = "configaddress2" #配置的签名机地址
    invest.To = "configaddress3"   # 配置的合约地址 ----这个应该是contract_info中对应链的vault地址

    #info = pool_infos[symbol]["contract_info"]["bsc_pancakestrategy"]
    #strategyAddresses = [info]

    #这里应该是配置中有很多策略和对应地址，程序需要拼接策略，找到对应地址
    strategystr = "bsc_pancake_btc_usdt"
    strategys = conf["strategyes"]

    strategyAddresses = ""  #策略地址
    for key in strategys:
        print(key + ':' + strategys[key])
        if strategystr in key:
            strategyAddresses = strategys[key]


    baseTokenAmount = [0]    #值从配资的计算结果得到
    counterTokenAmount = [0] #值从配资计算结果得到

    invest.StrategyAddresses = [0]  #从info中取（contract_info）
    invest.BaseTokenAmount = baseTokenAmount
    invest.CounterTokenAmount = counterTokenAmount

    sendToBridge = SendToBridgeParam()

    sendToBridge.ChainId = 52
    sendToBridge.ChainName = "bsc"
    sendToBridge.From = "configaddress2" #配置的签名机地址
    sendToBridge.To = "configaddress3"   # 配置的合约地址
    sendToBridge.BridgeAddress = "" #配置的地址
    sendToBridge.Amount = 1 * 10e18  # 精度配置读取
    sendToBridge.TaskID = TaskIds["BTC"]

    params = Params()
    params.CrossBalances = crossBalance
    params.ReceiveFromBridgeParams = receiveFromBridge
    params.InvestParams = invest
    params.SendToBridgeParams = sendToBridge

    #write db
    connect = getconnectinfo(conf["database"]["db"])
    print(connect)
    conn = pymysql.connect(host='127.0.0.1', port=3306, user='root', passwd='csquan253905', db='reblance', charset = 'utf8')
    print(conn)

    #cursor = db.cursor()

    #cursor.execute('''insert into Rebalance_params values()''')

    #cursor.close()
    #db.commit()
    conn.close()


