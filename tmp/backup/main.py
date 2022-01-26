import numpy as np
import scipy.optimize as op

if __name__ == '__main__':
    price = {'cake': float('12.067314265704663'),
             'bnb': float('584.7946894465347'),
             'usd': float('0.9941719646158913'),
             'btc': float('50000'),
             'usdt': float('0.99500624'),
             'eth': float('4350.6')}

    amount = {
        'cake': 1000,
        'bnb': 100,
        'usd': 34774,
        'btc': 1,
        'usdt': 89378,
        'eth': 1
    }

    daily_reward_info = {
        'bsc_pancake_bnb_usd': float('90736.31622747983404655133031'),
        'bsc_pancake_usdt_bnb': float('49492.53612407991606210850692'),
        'bsc_pancake_cake_usd': float('20621.89005169996140568426484'),
        'bsc_pancake_cake_usdt': float('16497.51204135996719377712936'),
        'bsc_biswap_usdt_bnb': float('83901.13684877362992515739288'),
        'bsc_biswap_bnb_usd': float('81057.03051491689773334655095'),
        'bsc_biswap_usdt_eth': float('51193.91400942119963042019203'),
        'bsc_biswap_usdt_btc': float('51193.91400942119963042019203'),
    }
    tvl = {
        'bsc_pancake_bnb_usd': float('462321085.54743105'),
        'bsc_pancake_usdt_bnb': float('269084921.09558576'),
        'bsc_pancake_cake_usd': float('29102284.526785344'),
        'bsc_pancake_cake_usdt': float('16173141.362343531'),
        'bsc_biswap_usdt_bnb': float('33705976.15534866'),
        'bsc_biswap_bnb_usd': float('38612573.482310146'),
        'bsc_biswap_usdt_eth': float('52148265.805380374'),
        'bsc_biswap_usdt_btc': float('50239764.45369236'),
    }


    def function(x):
        res = daily_reward_info['bsc_pancake_bnb_usd'] * (price['bnb'] * x[0] + price['usd'] * x[1]) / (
                tvl['bsc_pancake_bnb_usd'] + price['bnb'] * x[0] + price['usd'] * x[1]) + \
              daily_reward_info['bsc_pancake_usdt_bnb'] * (price['bnb'] * x[2] + price['usdt'] * x[3]) / (
                      tvl['bsc_pancake_usdt_bnb'] + price['bnb'] * x[2] + price['usdt'] * x[3]) + \
              daily_reward_info['bsc_pancake_cake_usd'] * (price['cake'] * x[4] + price['usd'] * x[5]) / (
                      tvl['bsc_pancake_cake_usd'] + price['cake'] * x[4] + price['usd'] * x[5]) + \
              daily_reward_info['bsc_pancake_cake_usdt'] * (price['cake'] * x[6] + price['usdt'] * x[7]) / (
                      tvl['bsc_pancake_cake_usdt'] + price['cake'] * x[6] + price['usdt'] * x[7]) + \
              daily_reward_info['bsc_biswap_usdt_bnb'] * (price['bnb'] * x[8] + price['usdt'] * x[9]) / (
                      tvl['bsc_biswap_usdt_bnb'] + price['bnb'] * x[8] + price['usdt'] * x[9]) + \
              daily_reward_info['bsc_biswap_bnb_usd'] * (price['bnb'] * x[10] + price['usd'] * x[11]) / (
                      tvl['bsc_biswap_bnb_usd'] + price['bnb'] * x[10] + price['usd'] * x[11]) + \
              daily_reward_info['bsc_biswap_usdt_eth'] * (price['eth'] * x[12] + price['usdt'] * x[13]) / (
                      tvl['bsc_biswap_usdt_eth'] + price['eth'] * x[12] + price['usdt'] * x[13]) + \
              daily_reward_info['bsc_biswap_usdt_btc'] * (price['btc'] * x[14] + price['usdt'] * x[15]) / (
                      tvl['bsc_biswap_usdt_btc'] + price['btc'] * x[14] + price['usdt'] * x[15])
        return res * -1000


    initx = [0] * 16

    cons = (
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

        {'type': 'ineq', 'fun': lambda x: amount['bnb'] - x[0] - x[2] - x[8] - x[10]},
        {'type': 'ineq', 'fun': lambda x: amount['cake'] - x[4] - x[6]},
        {'type': 'ineq', 'fun': lambda x: amount['eth'] - x[12]},
        {'type': 'ineq', 'fun': lambda x: amount['btc'] - x[14]},
        {'type': 'ineq', 'fun': lambda x: amount['usdt'] - x[3] - x[7] - x[9] - x[13] - x[15]},
        {'type': 'ineq', 'fun': lambda x: amount['usd'] - x[1] - x[5] - x[11]},
        {'type': 'eq', 'fun': lambda x: price['bnb'] * x[0] - price['usd'] * x[1]},
        {'type': 'eq', 'fun': lambda x: price['bnb'] * x[2] - price['usdt'] * x[3]},
        {'type': 'eq', 'fun': lambda x: price['cake'] * x[4] - price['usd'] * x[5]},
        {'type': 'eq', 'fun': lambda x: price['cake'] * x[6] - price['usdt'] * x[7]},
        {'type': 'eq', 'fun': lambda x: price['bnb'] * x[8] - price['usdt'] * x[9]},
        {'type': 'eq', 'fun': lambda x: price['bnb'] * x[10] - price['usd'] * x[11]},
        {'type': 'eq', 'fun': lambda x: price['eth'] * x[12] - price['usdt'] * x[13]},
        {'type': 'eq', 'fun': lambda x: price['btc'] * x[14] - price['usdt'] * x[15]},
    )
    res = op.minimize(fun=function,
                      x0=initx,
                      constraints=cons,
                      options={'maxiter': 1000})

    x = res.x
    print(res)

    print("*************************")

    print('bsc_pancake_bnb_usd bnb:{:.4f}  usd:{:.4f}'.format(x[0], x[1]))
    print('bsc_pancake_usdt_bnb bnb:{:.4f}  usdt:{:.4f}'.format(x[2], x[3]))
    print('bsc_pancake_cake_usd cake:{:.4f}  usd:{:.4f}'.format(x[4], x[5]))
    print('bsc_pancake_cake_usdt cake:{:.4f}  usdt:{:.4f}'.format(x[6], x[7]))
    print('bsc_biswap_usdt_bnb bnb:{:.4f}  usdt:{:.4f}'.format(x[8], x[9]))
    print('bsc_biswap_bnb_usd bnb:{:.4f}  usd:{:.4f}'.format(x[10], x[11]))
    print('bsc_biswap_usdt_eth eth:{:.4f}  usdt:{:.4f}'.format(x[12], x[13]))
    print('bsc_biswap_usdt_btc btc:{:.4f}  usdt:{:.4f}'.format(x[14], x[15]))
