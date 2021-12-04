from scipy.optimize import minimize
import numpy as np
import time, random

_RATE = 0.18

def computeTarget(todo_btc, todo_eth, todo_usdt, argsr, argstt):
    #print(repr(argsr),repr(argstt))
    A, B ,C, D, E, F, G ,H, I, J, K, L, M = argsr
    aa, bb, cc, dd, ee, ff, gg , hh, ii, jj, kk, ll, mm = argstt

    btc_bsc = todo_btc * ((365* G / _RATE )- gg) / ((365* G / _RATE)- gg  +  (365* K / _RATE )- kk )
    eth_bsc = todo_eth * ((365* H / _RATE )- hh) / ((365* H / _RATE)- hh + (365* I / _RATE )- ii  + (365* J / _RATE )- jj )
    _sub = (365* C / _RATE )- cc + (365* D / _RATE)- dd +(365* F / _RATE )- ff  +(365* G / _RATE )- gg +(365* H / _RATE )- hh
    usdt_bsc= todo_usdt *( _sub ) / ( _sub +(365* J / _RATE )- jj +(365* M / _RATE )- mm)
 
    return btc_bsc, eth_bsc, usdt_bsc


def totalReward(argsp, argsr, argst):

    bnb, busd, usdt, cake, btcb, eth = argsp
    A, B ,C, D, E, F, G , H ,*_ = argsr
    a, b, c, d, e, f, g , h ,*_ = argst

    func = lambda x: (-1)*(
            (bnb*x[0] + busd*x[3])/(bnb*x[0] + busd*x[3] + a )*A + \
            (bnb*x[1] + busd*x[2])/(bnb*x[1] + busd*x[2] + b )*B + \
            (bnb*x[5] + usdt*x[11])/(bnb*x[5] + usdt*x[11] + c )*C + \
            (bnb*x[6] + usdt*x[12])/(bnb*x[6] + usdt*x[12] + d )*D + \
            (cake*x[15] + busd*x[4])/(cake*x[15] + busd*x[4] + e )*E + \
            (cake*x[14] + usdt*x[13])/(cake*x[14] + usdt*x[13] + f )*F + \
            (btcb*x[7] + usdt*x[10])/(btcb*x[7] + usdt*x[10] + g )*G + \
            (eth*x[8] + usdt*x[9])/(eth*x[8] + usdt*x[9] + h )*H 
          )
    return func

LOW_BOUND = 0

def conSamllRe(argsq, argsp):
    bnb_q, busd_q, btcb_q, eth_q, usdt_q, cake_q = argsq
    bnb, busd, usdt, cake, btcb, eth = argsp
    # 约束条件 分为eq 和ineq
    # eq表示 函数结果等于0 ； ineq 表示 表达式大于等于0  
    cons = ({'type': 'ineq', 'fun': lambda x: x[0] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[0]},
            {'type': 'ineq', 'fun': lambda x: x[1]},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[1]},
            {'type': 'ineq', 'fun': lambda x: x[2] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: busd_q - x[2]},
            {'type': 'ineq', 'fun': lambda x: x[3] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: busd_q - x[3]},
            {'type': 'ineq', 'fun': lambda x: x[4] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: busd_q - x[4]},
            {'type': 'ineq', 'fun': lambda x: x[5] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[5]},
            {'type': 'ineq', 'fun': lambda x: x[6] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[6]},
            {'type': 'ineq', 'fun': lambda x: x[7] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: btcb_q - x[7]},
            {'type': 'ineq', 'fun': lambda x: x[8] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: eth_q - x[8]},
            {'type': 'ineq', 'fun': lambda x: x[9] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[9]},
            {'type': 'ineq', 'fun': lambda x: x[10] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[10]},
            {'type': 'ineq', 'fun': lambda x: x[11] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[11]},
            {'type': 'ineq', 'fun': lambda x: x[12] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[12]},
            {'type': 'ineq', 'fun': lambda x: x[13] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[13]},
            {'type': 'ineq', 'fun': lambda x: x[14] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: cake_q - x[14]}, 
            {'type': 'ineq', 'fun': lambda x: x[15] - LOW_BOUND},
            {'type': 'ineq', 'fun': lambda x: cake_q - x[15]}, 
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[10] - x[11] - x[12] - x[13]}, 
            {'type': 'ineq', 'fun': lambda x: busd_q - x[2] - x[3] - x[4]}, 
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[0] - x[1]- x[5] - x[6]}, 
            {'type': 'ineq', 'fun': lambda x: cake_q - x[14] - x[15]}, 
            {'type': 'ineq', 'fun': lambda x: btcb_q - x[7]}, 
            {'type': 'ineq', 'fun': lambda x: eth_q - x[8]}, 
            {'type': 'eq',   'fun': lambda x: bnb*x[0] - busd*x[3]}, 
            {'type': 'eq',   'fun': lambda x: bnb*x[1] - busd*x[2]},
            {'type': 'eq',   'fun': lambda x: bnb*x[5] - usdt*x[11]},
            {'type': 'eq',   'fun': lambda x: bnb*x[6] - usdt*x[12]},
            {'type': 'eq',   'fun': lambda x: cake*x[15] - busd*x[4]},
            {'type': 'eq',   'fun': lambda x: cake*x[14] - usdt*x[13]},
            {'type': 'eq',   'fun': lambda x: btcb*x[7] - usdt*x[10]},
            {'type': 'eq',   'fun': lambda x: eth*x[8] - usdt*x[9]},
            )
    return cons

def getRandX0(bounds):
    x = []
    for b in bounds:
        x.append( random.random() * b )
    return x

def minimizeWrapper(x0, argsq, argsp, argsr, argst):
    res_list = []
    cons = conSamllRe(argsq, argsp)
    for m in (#'Nelder-Mead',
              #'Powell', 
              'CG', 
              'BFGS',
              'L-BFGS-B',
              #'TNC',
              'SLSQP',
            ) :
        ss = time.time()
        result = minimize(totalReward(argsp, argsr, argst), x0, method=m, constraints=cons)
        res_tup = (result.fun, m, x0, result.x)
        if result.success:
            res_list.append(res_tup)
        else:
            print('not success:', res_tup)
    res_list.sort(key=lambda k:k[0], reverse=True)
    if len(res_list) > 1 :
        if res_list[-1][0]-res_list[0][0] > 0.000000000001:
            print('max diff val of all methods:', res_list[-1][0]-res_list[0][0] )
    return res_list[-1] if len(res_list) > 0 else None

LOOP_COUNT = 10
def doCompute(argsq, argsp, argsr, argst):
    bnb_q, busd_q, btcb_q, eth_q, usdt_q, cake_q = argsq
    #argsp = (bnb, busd, usdt, cake, btcb, eth) 
    #argsr = (A, B ,C, D, E, F, G ,H)
    #argst = (a, b, c, d, e, f, g, h)
    boundary = (bnb_q, bnb_q, busd_q, busd_q, busd_q, bnb_q, bnb_q, btcb_q,
                eth_q, usdt_q, usdt_q, usdt_q, usdt_q, usdt_q, cake_q, cake_q)
    res_list = []

    for i in range(LOOP_COUNT):
        res = minimizeWrapper(getRandX0(boundary), argsq, argsp, argsr, argst)
        if res:
            res_list.append(res)
    print("len of res list:",len(res_list))
    res_list.sort(key=lambda k:k[0], reverse=True)
    for res in res_list:
        #print('Time cost: %0.6f'%(time.time()-ss))
        #print('Method:',m)
        print('Value: %0.6f'%res[0])
        print('Method:', res[1])
        print('Init  X:',res[2])
        print('Final X:',res[3])
        print('================================================================')
    if len(res_list)==0 :
        return []
    return list(map(float, res_list[-1][3]))

if __name__ == '__main__':
    pass


