# -*- coding:utf-8 -*-
import requests
import json

def functionname(url):
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
            print("dailyReward is")
            print(dailyReward)
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

if __name__ == '__main__':
    print("+++++pancake")
    pancakeUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    functionname(pancakeUrl)
    print("+++++biswap")
    biswapUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=476'
    functionname(biswapUrl)
    print("+++++solo")
    soloUrl = 'https://api.schoolbuy.top/hg/v1/project/pool/list?projectId=63'
    functionname(soloUrl)
