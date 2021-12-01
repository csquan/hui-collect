# -*- coding:utf-8 -*-
import requests
import json

"""对照结果go结构体
type CrossBalanceItem struct {
	FromChain    string `json:"from_chain"`
	ToChain      string `json:"to_chain"`
	FromAddr     string `json:"from_addr"`
	ToAddr       string `json:"to_addr"`
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       string `json:"Amount"`
}

type ReceiveFromBridgeParam struct {
	ChainId   int
	ChainName string
	From      string
	To        string //合约地址

	Erc20ContractAddr common.Address //erc20 token地址，用于授权

	Amount *big.Int //链上精度值的amount，需要提前转换
	TaskID *big.Int
}
type InvestParam struct {
	ChainId   int
	ChainName string
	From      string
	To        string //合约地址

	StrategyAddresses  []*common.Address
	BaseTokenAmount    []*big.Int
	CounterTokenAmount []*big.Int
}

type Params struct {
	CrossBalances           []*CrossBalanceItem       `json:"cross_balances"`
	ReceiveFromBridgeParams []*ReceiveFromBridgeParam `json:"receive_from_bridge_params"`
	InvestParams            []*InvestParam            `json:"invest_params"`
}
"""
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
        pool.chain_id = pool_info["chain_id"]
        pool.symbol = pool_info["symbol"]
        pool.heco_uncross_quantity = pool_info["heco_uncross_quantity"]
        pool.bsc_vault_unre_qunatity = pool_info["bsc_vault_unre_qunatity"]
        pool.poly_vault_unre_qunatity = pool_info["poly_vault_unre_qunatity"]
        pool.contract_info = pool_info["contract_info"]

        pool_infos[pool_info["chain"]] = pool

    return pool_infos


if __name__ == '__main__':
    print("+++++pancake")
    pancakeUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    getprojectinfo(pancakeUrl)
    print("+++++biswap")
    biswapUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=476'
    getprojectinfo(biswapUrl)
    print("+++++solo")
    soloUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    getprojectinfo(soloUrl)

    #infoUrl = ''
    #getpoolinfo(infoUrl)
