# -*- coding:utf-8 -*-
import requests
import json
import time
import pymysql
import yaml
import pickle
import numpy as np
import json
import sys

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


class CrossItem:
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


counter_tokens = ["usd"]

def getCurrency(pair):
    tokenstr = pair.split('_')
    return tokenstr[0].lower()


def getPair(str,currencys):
    pair = Pair()
    tokenstr = str.split('/')  # 用/分割str字符串,etc:Cake/WBNB
    print(tokenstr)

    str1 = tokenstr[0].lower()
    str2 = tokenstr[1].lower()

    for key in currencys:
        for info in currencys[key]["tokens"]:
            for t in currencys[key]["tokens"][info]:
                if key == "eth" and info == "poly":
                    print("poly")
                if currencys[key]["tokens"][info]["symbol"].lower() == str1:
                    str1 = key
                if currencys[key]["tokens"][info]["symbol"].lower() == str2:
                    str2 = key


    pair.base = str1
    pair.counter = str2

    for counter in counter_tokens:
        if str1.find(counter) >= 0:
            pair.base = str2
            pair.counter = str1
        if str2.find(counter) >= 0:
            pair.base = str1
            pair.counter = str2
    return pair


def getreinfo(url):
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    e = json.loads(string)

    print(e["data"])

    return e["data"]

# 关于价格：函数获取价格填写进传入的currencys，同时将这个价格和对应币种返回
def getprojectinfo(project, url, currencys):
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    e = json.loads(string)
    reward = 0
    tvls = {}
    aprs = {}
    daily = {}

    if e["code"] != 200:
        print("project服务异常")
        sys.exit(1)

    for data in e["data"]:
        aprs[data["poolName"]] = data["apr"]

        #todo:临时修改
        tokenPair1 = getPair(data["poolName"], currencys)
        key1 = tokenPair1.base + '_' + tokenPair1.counter + '_' + project
        tvls[key1] = data["tvl"]

        for rewardToken in data["rewardTokenList"]:
            # 拼接dailyReward
            tokenPair = getPair(data["poolName"],currencys)
            key = tokenPair.base + '_' + tokenPair.counter + '_' + project
            dailyReward = float(rewardToken["dayAmount"]) * float(rewardToken["tokenPrice"])
            daily[key] = dailyReward
            reward = reward + dailyReward
        for deposit in data["depositTokenList"]:
            #首先以tokenAddress到config中查找，获取对应币种的名字
            for name in currencys:
                for token in currencys[name]["tokens"]:
                    if currencys[name]["tokens"][token]["addr"] == deposit["tokenAddress"]:
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


def getPairProject(str):
    info = str.split('_')
    ret = {}

    ret["base"] = info[0]
    ret["counter"] = info[1]
    ret["project"] = info[2]

    return ret


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


def getReParams(currency_infos, currency_dict,reinfo, beforeInfo):
    vaultInfoList = reinfo["vaultInfoList"]

    # 计算跨链的最终状态--配资结果  btc_bsc = 100 eth_bsc = 101 usdt_bsc = 102
    afterInfo = {"btc": [{"bsc": 100}, {"polygon": 200}], "eth": [{"bsc": 101}], "usdt": [{"bsc": 102}]}
    #afterInfo["pbtc"] = {"poly": poly_btc}

    #小re参数数组
    paramsList = []

    # 跨链信息 存储
    diffMap = {}

    # cross list
    crossList = []

    # 生成跨链参数, 需要考虑最小值
    for currency in afterInfo:
        for chain in ['bsc', 'polygon']:
            for info in afterInfo[currency]:
                for k in info.keys():
                    if currency in beforeInfo.keys():
                        diff = info[k] - float(beforeInfo[currency][chain]["amount"])
                        if diff > currency_dict[currency]["min"] or diff < currency_dict[currency]["min"] * -1:
                            diffMap[currency + '_' + chain] = diff  # todo:format to min decimal

    for currency in diffMap:
        targetMap = {
            'bsc': 'poly',
            'poly': 'bsc',
        }
        diff = diffMap[currency]
        crossItem = CrossItem()

        # TODO 只考虑了从HECO往其他链搬
        crossItem.From = "heco"
        crossItem.To = "bsc"

        prefixToken = getCurrency(currency)
        if float(beforeInfo[prefixToken][crossItem.From]["amount"]) > float(diff):
            crossItem.Amount = diff  # 绝对值
            beforeInfo[prefixToken][crossItem.From]["amount"] = str(float(beforeInfo[prefixToken][crossItem.From]["amount"]) - float(diff))
        else:
            # 前提是heco的大于最小额 format精度
            crossItem.Amount = beforeInfo[prefixToken][crossItem.From]["amount"]
            beforeInfo[prefixToken][crossItem.From]["amount"] = 0

        if float(beforeInfo[prefixToken]["heco"]["amount"]) > currency_dict[prefixToken]["min"]:
            #todo:format beforeInfo[currency]["heco"] 精度
            crossItem.Amount = beforeInfo[prefixToken][crossItem.From]["amount"]
            beforeInfo[prefixToken][crossItem.From]["amount"] = 0

            crossItem.FromCurrency = currency_dict[prefixToken]["tokens"][crossItem.From]["crossSymbol"]
            crossItem.ToCurrency = currency_dict[prefixToken]["tokens"][crossItem.To]["crossSymbol"]

        if float(crossItem.Amount) > 0:
            crossList.append(crossItem)

        receiveFromBridge = ReceiveFromBridgeParam()
        receiveFromBridge.ChainID = 52  # 配置
        receiveFromBridge.ChainName = "bsc"  # 配置
        receiveFromBridge.From = "configaddress2"  # 配置的签名机地址
        receiveFromBridge.To = "configaddress3"  # 配置的合约地址
        receiveFromBridge.Erc20ContractAddr = "configaddress4"  # 配置的token地址
        receiveFromBridge.Amount = float(crossItem.Amount) * 10e18  # todo:精度配置读取

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

        #存储此次找到的策略
        strategyAddresses = ""

        # 拼接策略:从api返回结果中找到对应地址 拼接规则：chain + "_" + project + "strategy"
        # 遍历8个交易对 currency_infos中的key:base_counter_project
        for key in currency_infos:
            # todo：chain_infos中不存在key对应的project的处理
            info = getPairProject(key)
            project = info["project"]
            # todo：api返回对应币种的contract_info不存在strategystr的处理
            for vaultInfo in vaultInfoList:
                        for chainName in vaultInfo["strategies"]:
                            for projectName in vaultInfo["strategies"][chainName]:
                                for strategyinfo in vaultInfo["strategies"][chainName][projectName]:
                                    if strategyinfo["tokenSymbol"] == info["base"]:  # 找到对应币种的策略信息 这里的问题：等式右边是info["base"]还是和info["counter"]的拼接？
                                        for elem in strategyinfo:
                                            if projectName.lower() == project and elem == 'strategyAddress':
                                                strategyAddresses = strategyinfo[elem]

            if strategyAddresses == "":
                print("配资的其中一个交易对策略在小re的返回数据中没有找到，请检查！")
                sys.exit(1)

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
        params.CrossBalances = crossItem
        params.ReceiveFromBridgeParams = receiveFromBridge
        params.InvestParams = invest
        params.SendToBridgeParams = sendToBridge

        # 序列化本次的小re params
        ret = pickle.dumps(params)
        # test
        aa = pickle.loads(ret)

        paramsList.append(ret)

    retList = pickle.dumps(paramsList)
    # test
    aa = pickle.loads(retList)

    return paramsList


def getAllDict(currency_dict):
    daily_dict, tvls_dict, aprs_dict, currencys_dict = {}, {}, {}, {}
    # 获取project info
    print('-------------------------------------------')
    pancakeUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    reward, daily, tvls, aprs, currencys = getprojectinfo("pancake", pancakeUrl, currency_dict)
    daily_dict.update(daily)
    tvls_dict.update(tvls)
    aprs_dict.update(aprs)
    currencys_dict.update(currencys)
    print('-------------------------------------------')
    biswapUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=476'
    reward, daily, tvls, aprs, currencys = getprojectinfo("biswap", biswapUrl, currency_dict)
    daily_dict.update(daily)
    tvls_dict.update(tvls)
    aprs_dict.update(aprs)
    currencys_dict.update(currencys)
    #print('-------------------------------------------')
    #soloUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    #reward, daily, tvls, aprs, currencys = getprojectinfo("solo", soloUrl, currency_dict)
    #daily_dict.update(daily)
    #tvls_dict.update(tvls)
    #aprs_dict.update(aprs)
    #currencys_dict.update(currencys)
    print('-------------------------------------------')
    polygonUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=112'
    reward, daily, tvls, aprs, currencys  = getprojectinfo("quickswap", polygonUrl, currency_dict)
    daily_dict.update(daily)
    tvls_dict.update(tvls)
    aprs_dict.update(aprs)
    currencys_dict.update(currencys)
    print('=================================================================')
    print("daily_dict:",daily_dict)
    print("tvls_dict:",tvls_dict)
    print("aprs_dict:",aprs_dict)
    print("currencys_dict:",currencys_dict)
    print('=================================================================')

    return daily_dict, tvls_dict, aprs_dict, currencys_dict

def getPrice(price_name, currencys_dict):
    for _, val in currencys_dict.items():
        if 'tokens' in val:
            for _, items in val.get('tokens').items():
                if items.get('symbol','').lower() == price_name:
                    price = val.get('price','')
                    try:
                        price = float(price)
                    except:
                        continue
                    else:
                        return price
    return None

NAME_LIST=  ('bnb_usd_biswap', 'bnb_usd_pancake', 'bnb_usdt_biswap', 'bnb_usdt_pancake', 'cake_usd_pancake',
             'cake_usdt_pancake','btc_usdt_biswap', 'eth_usdt_biswap','eth_usdc_quickswap','eth_usdt_quickswap',
             'btc_usdc_quickswap','matic_usdc_quickswap','matic_usdt_quickswap',
            )

def matchParams(daily_dict, tvls_dict, aprs_dict, currencys_dict):
    print("--------------------PRICE--------------------")
    price_name = ("bnb", "cake", "btcb","eth", "busd", "usdt")
    print("all currencys_dict:",currencys_dict)
    price_list = []
    for p in price_name:
        price_list.append(getPrice(p, currencys_dict))
    print(price_name)
    print(price_list)
    for k,v in dict(zip(price_name,price_list)).items():
        print("%s:%s"%(k,v))
    # params for price of :
    #    ('bnb', 'cake', 'btcb', 'eth', 'busd', 'usdt')
    argsp = tuple(price_list)

    print("---------------------TVL---------------------")
    tvl_name = NAME_LIST
    #!!!!!假值，用于测试！！！！
    tvls_dict.update(matic_usdc_quickswap=10000000.0)
    tvls_dict.update(matic_usdt_quickswap=20000000.0)
    print(tvls_dict)
    tvl_list = []
    for t in tvl_name:
        tvl_list.append(tvls_dict.get(t, None))
    print(tvl_name)
    print(tvl_list)
    for k,v in dict(zip(tvl_name, tvl_list)).items():
        print("%s:%s"%(k,v))
    # params for tvl of :
    #    ( aa, bb, cc, dd, ee, ff, gg, hh, ii, jj, kk, ll, mm ) 
    #in small re aslo for :
    #    ( a, b, c, d, e, f, g, h, i, j, k, l, m ) 
    argstt = tuple(map(float, tvl_list))
    print("--------------------Daily--------------------")
    #!!!!!假值，用于测试！！！！
    daily_dict.update(matic_usdc_quickswap=1000.0)
    daily_dict.update(matic_usdt_quickswap=2000.0)
    print("all daily_dict:",daily_dict)
    daily_name = NAME_LIST
    daily_reward = []
    for d in daily_name:
        daily_reward.append(daily_dict.get(d, None))
    print(daily_name)
    print(daily_reward)
    for k,v in dict(zip(daily_name,daily_reward)).items():
        print("%s:%s"%(k,v))
    # params for tvl of :
    #    ( A, B, C, D, E, F, G, H, I, J, K, L, M ) 
    argsr = tuple(map(float, daily_reward))
    print("---------------------Aprs---------------------")
    #print("all aprs_dict:",aprs_dict)
    return argsp, argsr, argstt

def outputReTask():
    # 读取config
    conf = read_yaml("./config.yaml")

    conf_currency_dict = conf.get("currencies")
    currencyName = conf_currency_dict.keys()

    daily_dict, tvls_dict, aprs_dict, currencys_dict = getAllDict(conf_currency_dict)

    argsp, argsr, argstt = matchParams(daily_dict, tvls_dict, aprs_dict, currencys_dict)
    
    reUrl = 'http://neptune-hermes-mgt-h5.test-15.huobiapps.com/v2/v1/open/re'
    reinfos = getreinfo(reUrl)
    threshold = reinfos["threshold"]
    vaultInfoList = reinfos["vaultInfoList"]
    
    # 整理出阈值，当前值 进行比较
    # {btc:{bsc:{amount:"1", controllerAddress:""},...}}
    beforeInfo = {}

    # 计算跨链的初始状态--todo:这里多个etc对应的值
    for vault in vaultInfoList:
        for name in currencyName:
            controller = {}
            if vault["tokenSymbol"].lower().find(name) > 0:
                for chain in vault["activeAmount"].keys():
                    controller[chain.lower()] = vault["activeAmount"][chain]
            if controller:
                beforeInfo[name.lower()] = controller
                # total = total + vault.activeAmount[chain]

    # 得到poly上的btc量
    btc_total = 0
    for controller in beforeInfo["btc"]:
        btc_total = btc_total + float(beforeInfo["btc"][controller]["amount"])

    poly_btc = btc_total - 100

    from re_optimize import computeTarget
    # 三个todo值，这里先做假值，后面需要到config取
    todo_btc = 1000
    todo_eth = 20000
    todo_usdt = 400000
    btc_bsc, eth_bsc, usdt_bsc = computeTarget(todo_btc, todo_eth, todo_usdt, argsr, argstt)
    print("three target:",btc_bsc, eth_bsc, usdt_bsc)
    # Bnb_q  bnb在bsc上的bnb总量，以下变量类似定义，应该re返回获取，但是目前返回数据不全，所以测试数据 其中Btcb_q Eth_q Usdt_q 由配资内部计算可得
    bnb_q  = 3000
    cake_q = 2000
    btcb_q = 500000 + btc_bsc
    eth_q  = 600000 + eth_bsc
    busd_q = 10000
    usdt_q = 8000 + usdt_bsc
    
    argsq = (bnb_q, busd_q, btcb_q, eth_q, usdt_q, cake_q)
    print("argsq:",argsq)

    # 下面进行配资计算
    from re_optimize import doCompute
    X = doCompute(argsq, argsp, argsr, argstt)
    X = np.array(X).reshape(4, 4)
    print('compute res:',X)
    # 这里先生成一个测试矩阵X，模拟配资计算结果x0-15
    #X = np.arange(16).reshape(4, 4)

    # 交易对赋值
    currencyPair_infos = getPairinfo(X)

    # 拼接结果字串
    paramsList = getReParams(currencyPair_infos, conf_currency_dict, reinfos, beforeInfo)

    # write db
    conn = pymysql.connect(host=conf["database"]["host"], port=conf["database"]["port"], user=conf["database"]["user"],
                           passwd=conf["database"]["passwd"], db=conf["database"]["db"], charset='utf8')
    print(conn)

    # cursor = db.cursor()

    # 遍历paramsList，每个元素写入
    # cursor.execute('''insert into Rebalance_params values()''')

    # cursor.close()
    # db.commit()
    conn.close()



if __name__ == '__main__':
    # 首先读取api的pool——info，将5个值累加，判断门槛

    # 获取pool infos todo:判断条件需要修改
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

    total = poolinfo.heco_uncross_quantity + poolinfo.crossed_quantity_in_bsc_controller + poolinfo.crossed_quantity_in_poly_controller + poolinfo.bsc_vault_unre_qunatity + poolinfo.poly_vault_unre_qunatity

    if total < 100:
       sys.exit(1)

    outputReTask()


