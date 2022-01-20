from scipy.optimize import minimize


def calc_invest(session, chain, balance_info_dict, price_dict, daily_reward_dict, apr_dict, tvl_dict):
    # 确定最终的资产分布情况
    # X(0) = bnb_amt_biswap_busd
    # X(3) = busd_amt_biswap_bnb

    # X(1) = bnb_amt_pancake_busd
    # X(2) = busd_amt_pancake_bnb

    # X(15) = cake_amt_pancake_busd
    # X(4) = busd_amt_pancake_cake

    # X(5) = bnb_amt_biswap_usdt
    # X(11) = usdt_amt_biswap_bnb

    # X(6) = bnb_amt_pancake_usdt
    # X(12) = usdt_amt_pancake_bnb

    # X(7) = btcb_amt_biswap_usdt
    # X(10) = usdt_amt_biswap_btcb

    # X(8) = eth_amt_biswap_usdt
    # X(9) = usdt_amt_biswap_eth

    # X(14) = cake_amt_pancake_usdt
    # X(13) = usdt_amt_pancake_cake
    def get_amount(currency, chain):
        if currency not in balance_info_dict:
            return 0

        if chain not in balance_info_dict[currency]:
            return 0

        if 'amount' not in balance_info_dict[currency][chain]:
            return 0

        return float(balance_info_dict[currency][chain]['amount'])

    def get_price(currency):
        if currency not in price_dict:
            return 0

        return float(price_dict[currency])

    def get_reward(key):
        if key not in daily_reward_dict:
            return 0

        return float(daily_reward_dict[key])

    def get_apr(key):
        if key not in apr_dict:
            return 0

        return float(apr_dict[key])

    def get_tvl(key):
        if key not in tvl_dict:
            return 0

        return float(tvl_dict[key])

    def function(x):
        input = [
            ['bsc', 'biswap', 'bnb', 'usd', x[0], x[3]],
            ['bsc', 'pancake', 'bnb', 'usd', x[1], x[2]],
            ['bsc', 'pancake', 'cake', 'usd', x[15], x[4]],
            ['bsc', 'biswap', 'bnb', 'usdt', x[5], x[11]],
            ['bsc', 'pancake', 'bnb', 'usdt', x[6], x[12]],
            ['bsc', 'biswap', 'btc', 'usdt', x[7], x[10]],
            ['bsc', 'biswap', 'eth', 'usdt', x[8], x[9]],
            ['bsc', 'pancake', 'cake', 'usdt', x[14], x[13]],
        ]

        total = 0

        for [chain, project, c1, c2, x1, x2] in input:
            key = '{}_{}_{}_{}'.format(chain, project, c1, c2)

            base = get_price(c1) * 2 + get_tvl(key)

            if base <= 0.01:
                continue

            r = get_reward(key) * (get_price(c1) * x1 * 2) / base
            total += r

        return total * -1

    cons = [
        {'type': 'ineq', 'fun': lambda x: x[0]},
        {'type': 'ineq', 'fun': lambda x: x[1]},
        {'type': 'ineq', 'fun': lambda x: x[2]},
        {'type': 'ineq', 'fun': lambda x: x[3]},
        {'type': 'ineq', 'fun': lambda x: x[4]},
        {'type': 'ineq', 'fun': lambda x: x[5]},
        {'type': 'ineq', 'fun': lambda x: x[6]},
        {'type': 'ineq', 'fun': lambda x: x[7]},
        {'type': 'ineq', 'fun': lambda x: x[8]},
        {'type': 'ineq', 'fun': lambda x: x[9]},
        {'type': 'ineq', 'fun': lambda x: x[10]},
        {'type': 'ineq', 'fun': lambda x: x[11]},
        {'type': 'ineq', 'fun': lambda x: x[12]},
        {'type': 'ineq', 'fun': lambda x: x[13]},
        {'type': 'ineq', 'fun': lambda x: x[14]},
        {'type': 'ineq', 'fun': lambda x: x[15]},

        {'type': 'ineq', 'fun': lambda x: get_amount('btc', 'bsc') - x[7]},
        {'type': 'ineq', 'fun': lambda x: get_amount('eth', 'bsc') - x[8]},
        {'type': 'ineq', 'fun': lambda x: get_amount('cake', 'bsc') - x[14] - x[15]},
        {'type': 'ineq', 'fun': lambda x: get_amount('bnb', 'bsc') - x[0] - x[1] - x[5] - x[6]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usd', 'bsc') - x[2] - x[3] - x[4]},
        {'type': 'ineq', 'fun': lambda x: get_amount('usdt', 'bsc') - x[9] - x[10] - x[11] - x[12] - x[13]},

        {'type': 'eq', 'fun': lambda x: x[0] * get_price('bnb') - x[3] * get_price('usd')},
        {'type': 'eq', 'fun': lambda x: x[1] * get_price('bnb') - x[2] * get_price('usd')},
        {'type': 'eq', 'fun': lambda x: x[15] * get_price('cake') - x[4] * get_price('usd')},
        {'type': 'eq', 'fun': lambda x: x[5] * get_price('bnb') - x[11] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[6] * get_price('bnb') - x[12] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[7] * get_price('btc') - x[10] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[8] * get_price('eth') - x[9] * get_price('usdt')},
        {'type': 'eq', 'fun': lambda x: x[14] * get_price('cake') - x[13] * get_price('usdt')},
    ]

    initx = [0] * 16

    res = minimize(function,
                   x0=initx,
                   constraints=cons,
                   )

    for x in res.x:
        print("{:.4f}".format(x))
    print("""
        success: {}
        message: {}
        value: {}

        tvl:{:.4f} apr:{:.4f} biswap btc:{:.4f} usdt:{:.4f}
        tvl:{:.4f} apr:{:.4f} biswap eth:{:.4f} usdt:{:.4f}
        tvl:{:.4f} apr:{:.4f} biswap bnb:{:.4f} usdt:{:.4f}
        tvl:{:.4f} apr:{:.4f} pancake bnb:{:.4f} usdt:{:.4f}
        tvl:{:.4f} apr:{:.4f} biswap bnb:{:.4f} usd:{:.4f}
        tvl:{:.4f} apr:{:.4f} pancake bnb:{:.4f} usd:{:.4f}
        tvl:{:.4f} apr:{:.4f} pancake cake:{:.4f} usdt:{:.4f}
        tvl:{:.4f} apr:{:.4f} pancake cake:{:.4f} usd:{:.4f}

        total info: btc:{} eth:{}, cake:{} bnb:{} usdt:{}
        price btc:{:.4f} eth:{:.4f} cake:{:.4f} bnb:{:.4f} usdt:{:.4f}
        total invest btc:{:.4f} eth:{:.4f} cake:{:.4f} bnb:{:.4f} usdt:{:.4f} usd:{:.4f}
        left btc:{:.4f} eth:{:.4f} cake:{:.4f} bnb:{:.4f} usdt:{:.4f} usd:{:.4f}
        """.format(
        res.success,
        res.message,
        res.fun,

        tvl_dict['bsc_biswap_btc_usdt'], apr_dict['bsc_biswap_btc_usdt'], res.x[7], res.x[10],
        tvl_dict['bsc_biswap_eth_usdt'], apr_dict['bsc_biswap_eth_usdt'], res.x[8], res.x[9],
        tvl_dict['bsc_biswap_bnb_usdt'], apr_dict['bsc_biswap_bnb_usdt'], res.x[5], res.x[11],
        tvl_dict['bsc_pancake_bnb_usdt'], apr_dict['bsc_pancake_bnb_usdt'], res.x[6], res.x[12],
        tvl_dict['bsc_biswap_bnb_usd'], apr_dict['bsc_biswap_bnb_usd'], res.x[0], res.x[3],
        tvl_dict['bsc_pancake_bnb_usd'], apr_dict['bsc_pancake_bnb_usd'], res.x[1], res.x[2],
        tvl_dict['bsc_pancake_cake_usdt'], apr_dict['bsc_pancake_cake_usdt'], res.x[14], res.x[13],
        tvl_dict['bsc_pancake_cake_usd'], apr_dict['bsc_pancake_cake_usd'], res.x[15], res.x[4],

        get_amount('btc', 'bsc'), get_amount('eth', 'bsc'), get_amount('cake', 'bsc'),
        get_amount('bnb', 'bsc'),
        get_amount('usdt', 'bsc'),
        get_price('btc'), get_price('eth'), get_price('cake'), get_price('bnb'), get_price('usdt'),

        res.x[7], res.x[8], res.x[14] + res.x[15], res.x[5] + res.x[6],
                            res.x[9] + res.x[10] + res.x[11] + res.x[12] + res.x[13],
                            res.x[2] + res.x[3] + res.x[4],
                            get_amount('btc', 'bsc') - res.x[7],
                            get_amount('eth', 'bsc') - res.x[8],
                            get_amount('cake', 'bsc') - res.x[14] - res.x[15],
                            get_amount('bnb', 'bsc') - res.x[5] - res.x[6] - res.x[0] - res.x[1],
                            get_amount('usdt', 'bsc') - (
                                    res.x[9] + res.x[10] + res.x[11] + res.x[12] + res.x[13]),
                            get_amount('usd', 'bsc') - (res.x[2] + res.x[3] + res.x[4])
    ))
