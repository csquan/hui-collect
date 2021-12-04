# coding=utf-8
from scipy.optimize import minimize
import numpy as np
import time, random
#token price
bnb  = 18.0
cake = 13.37
btcb = 1000.98
eth  = 200.8
busd = 60.9
usdt = 1.0
#usdc = 1.1
#matic= 25.0

#token quantity（total）
'''
bnb_total  = 100000000000
cake_total = 200000000000
btc_total  = 2000000000
eth_total  = 20000000000
busd_total = 100000000000 
usdt_total = 100000000000 
usdc_total = 100000000000
matic_total= 10000000000000
'''
'''
#token quantity (小re分配) 结果要返回这个！！
btc_bsc  = btc quantity   
eth_bsc  = eth quantity
usdt_bsc = usdt quantity
btc_poly = btc quantity
eth_poly = eth quantity
usdt_poly= usdt quantity

#token quantity（lp存入 + 小re分配）？？？？？？
Matic_Q = matic lp存matic
Wbtc_Q = matic lp存btc + btc_poly
Eth_Q = matic lp存eth + eth_poly
Usdt_Q = matic lp存usdt + usdt_poly
Usdc_Q = matic lp存usdc
'''
# pool daily reward               daily? yearly?
A = 0.012 #bnb_busd_biswap_daily_rewawrd 
B = 0.016 #bnb_busd_pancake_daily_rewawrd 
C = 0.019 #bnb_usdt_biswap_daily_reward
D = 0.02  #bnb_usdt_pancake_daily_reward
E = 0.016 #cake_busd_pancake_daily_reward
F = 0.026 #cake_usdt_pancake_daily_reward
G = 0.032 #btcb_usdt_biswap_daily_rewawrd
H = 0.001 #eth_usdt_biswap_daily_reward
I = 0.019 #eth_usdc_quickswap_daily_reward
J = 0.030 #eth_usdt_quickswap_daily_reward
K = 0.01  #wbtc_usdc_quickswap_daily_reward
L = 0.005 #matic_usdc_quickswap_daily_reward
M = 0.016 #matic_usdt_quickswap_daily_reward


N = 0.016 #bnb_bsolotop_daily_reward
O = 0.020 #cake_bsolotop_daily_reward
P = 0.046 #btcb_bsolotop_daily_reward
Q = 0.017 #eth_bsolotop_daily_reward
R = 0.011 #busd_bsolotop_daily_reward
S = 0.010 #usdt_bsolotop_daily_reward
T = 0.012 #btc_psolotop_daily_reward
U = 0.013 #eth_psolotop__daily_reward
V = 0.008 #matic_psolotop_daily_reward
W = 0.016 #usdt_psolotop_daily_reward
X = 0.002 #usdc_psolotop_daily_reward

# root amount 要求计算的目标
'''
X(0) = bnb_amt_biswap_busd
X(1) = bnb_amt_pancake_busd
X(2) = busd_amt_pancake_bnb
X(3) = busd_amt_biswap_bnb
X(4) = busd_amt_pancake_cake
X(5) = bnb_amt_biswap_usdt
X(6) = bnb_amt_pancake_usdt
X(7) = btcb_amt_biswap_usdt
X(8) = eth_amt_biswap_usdt
X(9) = usdt_amt_biswap_eth
X(10) = usdt_amt_biswap_btcb
X(11) = usdt_amt_biswap_bnb
X(12) = usdt_amt_pancake_bnb
X(13) = usdt_amt_pancake_cake
X(14) = cake_amt_pancake_usdt
X(15) = cake_amt_pancake_busd
X(16) = eth_amt_quickswap_usdc  
X(17) = eth_amt_quickswap_usdt
X(18) = btc_amt_quickswap_usdc
X(19) = matic_amt_quickswap_usdc
X(20) = matic_amt_quickswap_usdt
X(21) = usdc_amt_quickswap_eth
X(22) = usdt_amt_quickswap_eth
X(23) = usdc_amt_quickswap_btc
X(24) = usdc_amt_quickswap_matic
X(25) =usdt_amt_quickswap_matic
X(26) = bnb_bsolotop_amt
X(27) = cake_bsolotop_amt
X(28) = btcb_bsolotop_amt
X(29) = eth_bsolotop_amt
X(30) = busd_bsolotop_amt
X(31) = usdt_bsolotop_amt
X(32) = btc_psolotop_amt
X(33) = eth_psolotop_amt
X(34) = matic_psolotop_amt
X(35) = usdt_psolotop_amt
X(36) = usdc_psolotop_amt
'''


# pool tvl (不包括我们提供的部分)
a = 332300 #bnb_busd_biswap_pool_tvl
b = 4099766 #bnb_busd_pancake_pool_tvl
c = 3234325#bnb_usdt_biswap_pool_tvl
d = 46464#bnb_usdt_pancake_pool_tvl
e = 4326577#cake_busd_pancake_pool_tvl
f = 4545#cake_usdt_pancake_pool_tvl
g = 342646#btcb_usdt_biswap_pool_tvl
h = 345456#eth_usdt_biswap_pool_tvl
'''
i = eth_usdc_quickswap_pool_tvl
j = eth_usdt_quickswap_pool_tvl
k = wbtc_usdc_quickswap_pool_tvl
l = matic_usdc_quickswap_pool_tvl
m = matic_usdt_quickswap_pool_tvl
n = bnb_bsolotop_amt
o = cake_bsolotop_amt
p = btcb_bsolotop_amt
q = eth_bsolotop_amt
r = busd_bsolotop_amt
s = usdt_bsolotop_amt
t = btc_psolotop_amt
u = eth_psolotop__amt
v = matic_psolotop_amt
w = usdt_psolotop_amt
x = usdc_psolotop_amt
'''
# pool tvl (别人的+我们的)
aa = 5325#bnb_busd_biswap_pool_tvl
bb = 579890#bnb_busd_pancake_pool_tvl
cc = 431321#bnb_usdt_biswap_pool_tvl
dd = 5436#bnb_usdt_pancake_pool_tvl
ee = 97949#cake_busd_pancake_pool_tvl
ff = 1235#cake_usdt_pancake_pool_tvl
gg = 135#btcb_usdt_biswap_pool_tvl
hh = 6557899#eth_usdt_biswap_pool_tvl
ii = 23#eth_usdc_quickswap_pool_tvl
jj = 32513#eth_usdt_quickswap_pool_tvl
kk = 314765482#wbtc_usdc_quickswap_pool_tvl
ll = 3547#matic_usdc_quickswap_pool_tvl
mm = 769#matic_usdt_quickswap_pool_tvl
'''
nn = bnb_bsolotop_amt
oo = cake_bsolotop_amt
pp = btcb_bsolotop_amt
qq = eth_bsolotop_amt
rr = busd_bsolotop_amt
ss = usdt_bsolotop_amt
tt = btc_psolotop_amt
uu = eth_psolotop__amt
vv = matic_psolotop_amt
ww = usdt_psolotop_amt
xx = usdc_psolotop_amt
'''
to_btc  = 2000     #待跨链btc
to_eth  = 300000   #待跨链eth
to_usdt = 40000000 #待跨链usdt

btc_bsc = to_btc * ((365* G / 0.18)- gg) / ((365* G / 0.18) - gg  +  (365* K / 0.18)- kk )
eth_bsc = to_eth * ((365* H / 0.18)- hh) / ((365* H / 0.18)- hh + (365* I / 0.18)- ii  + (365* J / 0.18)- jj )
usdt_bsc= to_usdt *((365* C / 0.18)- cc + (365* D / 0.18)- dd +(365* F / 0.18)- ff  +(365* G / 0.18)- gg +(365* H / 0.18)- hh ) / ((365* C / 0.18)- cc+(365 * D / 0.18)- dd +(365* F / 0.18)- ff  +(365* G / 0.18)- gg +(365* H / 0.18)- hh +(365* J / 0.18)- jj +(365* M / 0.18)- mm)

#token quantity-bsc ( lp存入 + 小re分配)
bnb_q  =  3000
cake_q =  400
btcb_q =  500000 + btc_bsc
eth_q  =  6000 + eth_bsc
busd_q =  700
usdt_q =  8000 + usdt_bsc 

def total_reward(argsp, argsr, argst):
    bnb, busd, usdt, cake, btcb, eth = argsp
    A, B ,C, D, E, F, G , H = argsr
    a, b, c, d, e, f, g , h = argst
    v = lambda x: (
        (bnb*x[0] + busd*x[3])/(bnb*x[0] + busd*x[3] + a )*A + \
        (bnb*x[1] + busd*x[2])/(bnb*x[1] + busd*x[2] + b )*B + \
        (bnb*x[5] + usdt*x[11])/(bnb*x[5] + usdt*x[11] + c )*C + \
        (bnb*x[6] + usdt*x[12])/(bnb*x[6] + usdt*x[12] + d )*D + \
        (cake*x[15] + busd*x[4])/(cake*x[15] + busd*x[4] + e )*E + \
        (cake*x[14] + usdt*x[13])/(cake*x[14] + usdt*x[13] + f )*F + \
        (btcb*x[7] + usdt*x[10])/(btcb*x[7] + usdt*x[10] + g )*G + \
        (eth*x[8] + usdt*x[9])/(eth*x[8] + usdt*x[9] + h )*H 
          )*(-1)
    return v

def con_samllre(argsq, argsp):
    bnb_q, busd_q, btcb_q, eth_q, usdt_q, cake_q = argsq
    bnb, busd, usdt, cake, btcb, eth = argsp

    # 约束条件 分为eq 和ineq
    # eq表示 函数结果等于0 ； ineq 表示 表达式大于等于0  
    cons = ({'type': 'ineq', 'fun': lambda x: x[0]},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[0]},
            {'type': 'ineq', 'fun': lambda x: x[1]},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[1]},
            {'type': 'ineq', 'fun': lambda x: x[2]},
            {'type': 'ineq', 'fun': lambda x: busd_q - x[2]},
            {'type': 'ineq', 'fun': lambda x: x[3]},
            {'type': 'ineq', 'fun': lambda x: busd_q - x[3]},
            {'type': 'ineq', 'fun': lambda x: x[4]},
            {'type': 'ineq', 'fun': lambda x: busd_q - x[4]},
            {'type': 'ineq', 'fun': lambda x: x[5]},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[5]},
            {'type': 'ineq', 'fun': lambda x: x[6]},
            {'type': 'ineq', 'fun': lambda x: bnb_q - x[6]},
            {'type': 'ineq', 'fun': lambda x: x[7]},
            {'type': 'ineq', 'fun': lambda x: btcb_q - x[7]},
            {'type': 'ineq', 'fun': lambda x: x[8]},
            {'type': 'ineq', 'fun': lambda x: eth_q - x[8]},
            {'type': 'ineq', 'fun': lambda x: x[9]},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[9]},
            {'type': 'ineq', 'fun': lambda x: x[10]},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[10]},
            {'type': 'ineq', 'fun': lambda x: x[11]},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[11]},
            {'type': 'ineq', 'fun': lambda x: x[12]},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[12]},
            {'type': 'ineq', 'fun': lambda x: x[13]},
            {'type': 'ineq', 'fun': lambda x: usdt_q - x[13]},
            {'type': 'ineq', 'fun': lambda x: x[14]},
            {'type': 'ineq', 'fun': lambda x: cake_q - x[14]}, 
            {'type': 'ineq', 'fun': lambda x: x[15]},
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

def get_rand_x0(bounds):
    x = []
    for b in bounds:
        x.append( random.random() * b )
    return x

def minimize_wrapper(x0, argsq, argsp, argsr, argst):
    res_list = []
    cons = con_samllre(argsq, argsp)
    for m in ('Nelder-Mead',
              'Powell', 
              'CG', 
              'BFGS',
              'Newton-CG',
              'L-BFGS-B',
              'TNC',
              'COBYLA', 
              'dogleg', 
              'trust-ncg',
              'SLSQP'
            ) :
        ss = time.time()
        result = minimize(total_reward(argsp, argsr, argst), x0, method='SLSQP', constraints=cons)
        res_tup = (result.fun, m, x0, result.x)
        if result.success:
            res_list.append(res_tup)
        else:
            print('not success:', res_tup)
    res_list.sort(key=lambda k:k[0], reverse=True)
    if len(res_list) > 1 :
        if res_list[-1][0]-res_list[0][0] > 0.000000000001:
            print('max diff val of all methods:', res_list[-1][0]-res_list[0][0] )
    return res_tup if len(res_list) > 0 else None

if __name__ == "__main__":
    LOOP_COUNT = 100
    #x0 = np.asarray((1,1,2,2,3,3,4,4,5,5,6,6,7,7,8,8))  # 初始猜测值 from  x[0] to x[15]
    #x0 = np.asarray((0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0))
    #x0 = np.asarray( tuple([random.random() for i in range(16)]))
    argsq = (bnb_q, busd_q, btcb_q, eth_q, usdt_q, cake_q)
    argsp = (bnb, busd, usdt, cake, btcb, eth) 
    argsr = (A, B ,C, D, E, F, G ,H)
    argst = (a, b, c, d, e, f, g, h)
    boundary = (bnb_q, bnb_q, busd_q, busd_q, busd_q, bnb_q, bnb_q, btcb_q,
                eth_q, usdt_q, usdt_q, usdt_q, usdt_q, usdt_q, cake_q, cake_q)
    res_list = []

    for i in range(LOOP_COUNT):
        res = minimize_wrapper(get_rand_x0(boundary), argsq, argsp, argsr, argst)
        if res:
            res_list.append(res)

    res_list.sort(key=lambda k:k[0], reverse=True)
    for res in res_list:
        #print('Time cost: %0.6f'%(time.time()-ss))
        #print('Method:',m)
        print('Value: %0.6f'%res[0])
        print('Init  X:',res[2])
        print('Final X:',res[3])
        print('================================================================')

   
