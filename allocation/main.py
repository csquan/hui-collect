# -*- coding:utf-8 -*-
import requests
import json
import time
import pymysql

class Pool:
    pass

class CrossBalanceItem:
    pass

class ReceiveFromBridgeParam:
    pass

class InvestParam:
    pass

class Params:
    pass


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
    crossBalance.ToChain = "bsc" #这里从郭获得,是多个跨链，应该分别处理，这里假设heco向bsc跨链
    crossBalance.FromAddr = "configaddress1" #配置-从config读取
    crossBalance.ToAddr = "configaddress2"   #配置的签名机地址
    crossBalance.FromCurrency = "hbtc" #hbtc?
    crossBalance.ToCurrency = "btc"   #btc?

    symbol = "HBTC"
    info = pool_infos[symbol]
    crossBalance.Amount = 0 #info["heco_uncross_quantity"] + info["crossed_quantity_in_bsc_controller"] + info["crossed_quantity_in_poly_controller"] + info["bsc_vault_unre_qunatity"] + info["bsc_vault_unre_qunatity"]

    receiveFromBridge = ReceiveFromBridgeParam()
    receiveFromBridge.ChainID = 52
    receiveFromBridge.ChainName = "bsc"
    receiveFromBridge.From = "configaddress2"   #配置的签名机地址
    receiveFromBridge.To = "configaddress3"  # 配置的合约地址
    receiveFromBridge.Erc20ContractAddr = "configaddress4"  # 配置的token地址
    #下面的精度值从哪里取？这里假设跨1个btc
    receiveFromBridge.Amount = 1*10e18
    #生成全局唯一的task🆔并保存币种和taskID的对应关系
    TaskIds = {}
    t = time.time()
    receiveFromBridge.TaskID = int(round(t * 1000)) #毫秒级时间戳
    TaskIds["BTC"] = receiveFromBridge.TaskID
    #这里下面哪里还能用到TaskIds["BTC"]？

    #这里跨的币种是BTC，从pool_infos[symbol]找到BSC对应的策略
    invest = InvestParam()
    invest.ChinId = 52
    invest.ChainName = "bsc"
    invest.From =  "configaddress2"   #配置的签名机地址
    invest.To = "configaddress3"  # 配置的合约地址

    #这里以pancake为例，实际中应该是郭给的结果中指定,这里少个counter对？
    info = pool_infos[symbol]["contract_info"]["bsc_pancakestrategy"]
    strategyAddresses = [info]
    baseTokenAmount = [0]  #值从郭的计算结果得到

    counterTokenAmount = [0] #遍历郭给的每一个币种，在pancakeinfos中找到基础货币，取depositTokenList中的tokenAmount 以最大精度计算？

    invest.StrategyAddresses = strategyAddresses
    invest.BaseTokenAmount = baseTokenAmount
    invest.CounterTokenAmount = counterTokenAmount

    params = Params()
    params.CrossBalances = crossBalance
    params.ReceiveFromBridgeParams = receiveFromBridge
    params.InvestParams = invest

    #写入db
    db = pymysql.connect('localhost', 'root', '1234', 'rebalance')
    cursor = db.cursor()

    cursor.execute('''insert into Rebalance_params values()''')

    cursor.close()
    db.commit()
    db.close()