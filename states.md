# rebalance
|state|desc|
|:----|:---|
|0|rebalance初始化|
|1|source链——valut合约转出至bridge地址|
|2|执行跨链|
|3|target链——bridge地址转至valut地址|
|4|调用vault合约invest|
|5|成功|
|6|失败|

# full_rebalance
|state|desc|
|:----|:---|
|0|初始化|
|1|无常损失处理,调用自动划转计算无常损失，将token转至保证金账户合约|
|2|调用valut claimAll，具体token补偿无常损失|
|3|剩余余资金从保证金账户(合约）转到对冲账户(中心化账户)|
|4|资金跨回到heco|
|5|python计算part rebalance task|
|6|part rebalance 执行中|
|7|成功|
|8|失败|

# cross
|state|desc|
|:----|:---|
|0|创建跨链子任务|
|1|子任务创建成功|
|2|成功|

## cross sub
|state|desc|
|:----|:---|
|0|向bridge服务提交跨链任务|
|1|执行跨链|
|2|成功|

# transaction
|state|desc|
|:----|:---|
|0|初始化,并签名|
|1|交易审计|
|2|交易校验|
|3|发送transaction,检查receipt|
|4|成功|
|5|失败|