|error|log|line|file|排查|
|:----|:--|:---|:----|:---|
|getLpData err|claim get lpData err|274|claim_lp_handler.go|检查lpdata接口
|根据tokenSymbol和chain获取vault地址失败|vault addr not found chain|145|claim_lp_handler.go|检查lpdata接口返回vaultInfoList|
|baseTokenAmount/quoteTokenAmount空值|str to decimal err|126|claim_lp_handler.go|检查接口返回值liquidityProviderList|
|获取base/quote decimal失败|unexpectd decimal|197|claim_lp_handler.go|检查tokens table和查询的chain symbol参数是否匹配|
|abi pack claimAll参数失败|claim pack err|216|claim_lp_handler.go|检查claimAll合约参数和实际的传参,参数日志包含claimAll tid|
|调用claimAll获取from地址失败|get from addr empty chainName|222|claim_lp_handler.go|检查chainName与配置文件中的chains中的key是否匹配|
|调用claimAll失败|call claim fail|308|claim_lp_handler.go|查询链上调用合约的交易，根据txtask的id获取tx_hash|