# -*- coding:utf-8 -*-
import traceback
import requests
import yaml
import json
import utils
import time
import logging
from functools import reduce
from decimal import *
from sqlalchemy.orm import sessionmaker
from orm import *

dest_chains = ['bsc', 'polygon']


def get_pool_info(url):
    # 存储从api获取的poolinfo
    ret = requests.get(url)
    string = str(ret.content, 'utf-8')
    #print(string)
    e = json.loads(string)

    return e['data']


def read_yaml(path):
    with open(path, 'r', encoding='utf8') as f:
        return yaml.safe_load(f.read())


def format_addr(addr):
    if addr.startswith('0x'):
        return addr.lower()
    else:
        return ('0x' + addr).lower()


def format_currency_name(currency_name_set, name):
    for k in currency_name_set:
        if name.lower().find(k) >= 0:
            return True, k

    return False, name


class Project:
    def __init__(self, chain, name, url):
        self.chain = chain
        self.name = name
        self.url = url

    def get_info(self):
        res = requests.get(self.url)
        if res.status_code != 200:
            raise (Exception('connection error : %s' % str(res.content, 'utf-8')))
        string = str(res.content, 'utf-8')
        print(string)
        e = json.loads(string)
        return e['data']


def generate_strategy_key(chain, project, currencies):
    currencies = list(filter(lambda x: x is not None, currencies))
    currencies.sort(key=len, reverse=True)
    return "{}_{}_{}".format(chain, project, '_'.join(currencies))


def calc_cross_params(conf, session, currencies, account_info, daily_reward, apr, tvl):
    cross_balances = []
    send_to_bridge = []
    receive_from_bridge = []

    # 计算跨链的最终状态
    after_balance_info = {}
    for currency in account_info:
        caps = {}
        for chain in dest_chains:
            strategies = find_strategies_by_chain_and_currency(session, chain, currency)
            caps[chain] = Decimal(0)

            for s in strategies:
                # 先忽略单币
                if s.currency1 is None:
                    continue

                key = generate_strategy_key(s.chain, s.project, [s.currency0, s.currency1])
                if key not in apr or apr[key] < Decimal(0.18):
                    continue

                caps[chain] += (daily_reward[key] * Decimal(365) / Decimal(0.18) - tvl[key])

        total = Decimal(0)
        for item in account_info[currency].values():
            total += item['amount']

        caps_total = sum(caps.values())
        for k, v in caps.items():
            if v > 0:
                after_balance_info[currency] = {
                    k: total * v / caps_total
                }

    logging.info("calc final state: {}".format(after_balance_info))

    # 跨链信息
    balance_diff_map = {}

    # 生成跨链参数, 需要考虑最小值
    for currency in after_balance_info:
        for chain in dest_chains:
            if chain not in after_balance_info[currency] or currencies[currency].min is None:
                continue

            diff = after_balance_info[currency][chain] - account_info[currency][chain]['amount']
            if abs(diff) < Decimal(currencies[currency].min):
                continue

            if currency not in balance_diff_map:
                balance_diff_map[currency] = {}
            balance_diff_map[currency][chain] = diff.quantize(
                Decimal(10) ** (-1 * currencies[currency].crossDecimal),
                ROUND_DOWN)  # format to min scale

    logging.info("diff map:{}".format(balance_diff_map))

    def add_cross_item(conf, currency, from_chain, to_chain, amount):
        if amount > Decimal(currencies[currency].min):
            token_decimal = currencies[currency].tokens[from_chain].decimal

            account_info[currency][from_chain]['amount'] -= amount
            account_info[currency][to_chain]['amount'] += amount

            cross_balances.append({
                'from_chain': from_chain,
                'to_chain': to_chain,
                'from_addr': conf['bridge_port'][from_chain],
                'to_addr': conf['bridge_port'][to_chain],
                'from_currency': currencies[currency].tokens[from_chain].crossSymbol,
                'to_currency': currencies[currency].tokens[to_chain].crossSymbol,
                'amount': amount,
            })

            task_id = '{}'.format(time.time_ns() * 100)
            send_to_bridge.append({
                'chain_name': from_chain,
                'chain_id': conf['chain'][from_chain],
                'from': conf['bridge_port'][from_chain],
                'to': account_info[currency][from_chain]['controller'],
                'bridge_address': conf['bridge_port'][from_chain],
                'amount': amount * (Decimal(10) ** token_decimal),
                'task_id': task_id
            })
            receive_from_bridge.append({
                'chain_name': to_chain,
                'chain_id': conf['chain'][to_chain],
                'from': conf['bridge_port'][to_chain],
                'to': account_info[currency][to_chain]['controller'],
                "erc20_contract_addr": currencies[currency].tokens[from_chain].address,
                'amount': amount * (Decimal(10) ** token_decimal),
                'task_id': task_id,
            })

    for currency in balance_diff_map:
        target_chain = {
            'bsc': 'polygon',
            'polygon': 'bsc',
        }

        for chain in balance_diff_map[currency]:
            diff = balance_diff_map[currency][chain]

            # 向其他链进行跨链操作
            if diff < 0:
                add_cross_item(conf, currency, chain, target_chain[chain], diff * -1)
                # 先从heco向目标链转移，然后再从其他链向目标链转移
            elif account_info[currency]['heco']['amount'] >= diff:
                add_cross_item(conf, currency, 'heco', chain, diff)
            else:
                heco_amount = account_info[currency]['heco']['amount']
                add_cross_item(conf, currency, 'heco', chain, account_info[currency]['heco']['amount'].quantize(
                    Decimal(10) ** (-1 * currencies[currency].crossDecimal),
                    ROUND_DOWN))

                add_cross_item(conf, currency, target_chain[chain], chain,
                               (diff - heco_amount).quantize(
                                   Decimal(10) ** (-1 * currencies[currency].crossDecimal),
                                   ROUND_DOWN))

    return account_info, cross_balances, send_to_bridge, receive_from_bridge


def calc_re_balance_params(conf, session, currencies):
    res = {}

    # 注意usdt与usd的区分，别弄混了
    currency_names = [k for k in sorted(currencies.keys(), key=len, reverse=True)]

    logging.info("currencies:{}".format(currencies))

    # 获取rebalance所需业务信息
    re_balance_input_info = get_pool_info(conf['pool']['url'])

    threshold_org = re_balance_input_info['threshold']
    vault_info_list = re_balance_input_info['vaultInfoList']

    # 整理出阈值，当前值 进行比较 {usdt:{bsc:{amount:"1", controllerAddress:""}}}
    threshold_info = {}
    account_info = {}
    strategy_addresses = {}

    # 计算跨链的初始状态
    for vault in vault_info_list:
        (ok, currency) = format_currency_name(currency_names, vault['tokenSymbol'])
        if ok:
            for chain in vault['activeAmount']:
                if currency not in account_info:
                    account_info[currency] = {}

                amt_info = vault['activeAmount'][chain]
                if 'amount' not in amt_info:
                    amt_info['amount'] = '0'
                if 'controllerAddress' not in amt_info:
                    amt_info['controllerAddress'] = None

                account_info[currency][chain.lower()] = amt_info
                account_info[currency][chain.lower()]['amount'] = Decimal(
                    account_info[currency][chain.lower()]['amount'])
                account_info[currency][chain.lower()]['controller'] = account_info[currency][chain.lower()][
                    'controllerAddress']

            for chain, proj_dict in vault['strategies'].items():
                for project, st_list in proj_dict.items():
                    for st in st_list:
                        tokens = [format_currency_name(currency_names, c)[1] for c in st['tokenSymbol'].split('-')]
                        strategy_addresses[generate_strategy_key(chain.lower(), project.lower(), tokens)] = st[
                            'strategyAddress']

    account_info['usdt'] = {
        'bsc': {
            'amount': Decimal(0),
            'controller': ''
        },
        'heco': {
            'amount': Decimal(10000000),
            'controller': ''
        },
        'polygon': {
            'amount': Decimal(0),
            'controller': ''
        }
    }

    account_info['eth'] = {
        'bsc': {
            'amount': Decimal(100),
            "controller": ""
        },
        'heco': {
            'amount': Decimal(0),
            "controller": ""
        },
        'polygon': {
            'amount': Decimal(0),
            "controller": ""
        }
    }
    account_info['btc'] = {
        'bsc': {
            'amount': Decimal(10),
            "controller": ""
        },
        'heco': {
            'amount': Decimal(0),
            "controller": ""
        },
        'polygon': {
            'amount': Decimal(0),
            "controller": ""
        }
    }
    #
    # account_info['cake'] = {
    #     'bsc': {
    #         'amount': Decimal(0),
    #     }
    # }
    #
    # account_info['bnb'] = {
    #     'bsc': {
    #         'amount': Decimal(2000),
    #     }
    # }
    logging.info("init balance info for cross:{}".format(account_info) + \
                 "strategy address info:{}".format(strategy_addresses))

    # 计算阈值
    for threshold in threshold_org:
        (ok, currency) = format_currency_name(currency_names, threshold['tokenSymbol'])
        if ok:
            threshold_info[currency] = threshold['thresholdAmount']
    logging.info("threshold info after format:{}".format(threshold_info))

    # 比较阈值
    need_re_balance = False
    for currency in threshold_info:
        # 没有发现相关资产
        if currency not in account_info:
            continue

        total = Decimal(0)
        for item in account_info[currency].values():
            total += item['amount']

        need_re_balance = total > Decimal(threshold_info[currency])
        if need_re_balance:
            break

    # 没超过阈值
    if not need_re_balance:
        logging.info("deposit amount not large enough")
        return

    # 获取apr等信息
    projects = [Project(p['chain'], p['name'], p['url']) for p in conf['project']]

    # chain_project_coin1_coin2
    apr = {}
    price = {}
    daily_reward = {}
    tvl = {}

    for p in projects:
        info = p.get_info()

        for pool in info:
            for token in pool['rewardTokenList'] + pool['depositTokenList']:
                currency = find_currency_by_address(session, format_addr(token['tokenAddress']))
                if currency is not None and currency not in price:
                    price[currency] = Decimal(token['tokenPrice'])

                names = [format_currency_name(currency_names, c)[1] for c in pool['poolName'].split("/")]
                key = generate_strategy_key(p.chain, p.name, names)

                apr[key] = Decimal(pool['apr'])
                tvl[key] = Decimal(pool['tvl'])
                daily_reward[key] = reduce(lambda x, y: x + y,
                                           map(lambda t: Decimal(t['tokenPrice']) * Decimal(t['dayAmount']),
                                               pool['rewardTokenList']))
    logging.info(         
        "apr info:{}"\
    "    price info:{}"\
    "    daily reward info:{}"\
    "    tvl info:{}".format(apr, price, daily_reward, tvl)
    )


    account_info, cross_balances, send_to_bridge, receive_from_bridge = calc_cross_params(conf, session, currencies, account_info, 
                                                                                          daily_reward, apr, tvl)
    res['send_to_bridge_params'] = send_to_bridge
    res['receive_from_bridge_params'] = receive_from_bridge
    res['cross_balances'] = cross_balances

    res['invest_params'] = []
    for chain in dest_chains:
        invest_result = calc_invest(session, chain, account_info, price, daily_reward, apr, tvl)
        invest_param_list = generate_invest_params(conf, session, currencies, account_info, chain, strategy_addresses, invest_result)
        res['invest_params'].extend(invest_param_list)

    return res


def get_info_by_strategy_str(lp):
    data = lp.split('_')
    if len(data) < 3:
        return None, None
    elif len(data) == 3:
        return data, data[2], None
    else:
        return data, data[2], data[3]


def get_counter_currency(session, lp):
    data, token0, token1 = get_info_by_strategy_str(lp)
    st = find_strategies_by_chain_project_and_currencies(session, data[0], data[1], token0, token1)
    return st[0].currency1


def get_base_currency(session, lp):
    data, token0, token1 = get_info_by_strategy_str(lp)
    st = find_strategies_by_chain_project_and_currencies(session, data[0], data[1], token0, token1)
    return st[0].currency0


def generate_invest_params(conf, session, currencies, account_info, chain, strategy_addresses, invest_calc_result):
    res = []
    st_by_base = {}
    # 根据base token 进行分组
    for st, st_amounts in invest_calc_result.items():

        base = get_base_currency(session, st)
        if base not in st_by_base:
            st_by_base[base] = []

        st_by_base[base].append((st, st_amounts))

    for base, st_info in st_by_base.items():

        invest_addr = []
        invest_base = []
        invest_counter = []

        for (st, st_amounts) in st_info:
            if st not in strategy_addresses:
                continue

            invest_addr.append(strategy_addresses[st])
            base_decimal = currencies[base].tokens[chain].decimal

            base_amount = (st_amounts[base] * (Decimal(10) ** base_decimal)).quantize(Decimal(1), ROUND_DOWN)
            invest_base.append(base_amount)

            counter = get_counter_currency(session, st)
            if counter is None:
                invest_counter.append(0)
            else:
                counter_decimal = currencies[counter].tokens[chain].decimal
                counter_amount = (st_amounts[counter] * (Decimal(10) ** counter_decimal)).quantize(Decimal(1),
                                                                                                   ROUND_DOWN)
                invest_counter.append(counter_amount)

        if len(invest_addr) > 0:
            res.append({
                'chain_name': chain,
                'chain_id': conf['chain'][chain],
                'from': conf['bridge_port'][chain],
                'to': account_info[base][chain]['controller'],
                "strategy_addresses": invest_addr,
                'base_token_amount': invest_base,
                'counter_token_amount': invest_counter,
            })

    return res


"""
此策略对于单稳定币的情况符合预期，对于多稳定币的情况，可能求出的解不是最优
"""


def calc_invest(session, chain, balance_info_dict, price_dict, daily_reward_dict, apr_dict, tvl_dict):
    def f(key):
        infos = get_info_by_strategy_str(key)
        strategies = find_strategies_by_chain_project_and_currencies(session, chain, infos[0][1], infos[1], infos[2])
        return len(strategies) > 0

    invest_calc_result = {}

    detla = Decimal(0.005)

    while True:
        lpKeys = [k for k in sorted(apr_dict, key=apr_dict.get, reverse=True)]
        lpKeys = list(filter(f, lpKeys))
        if len(lpKeys) <= 0:
            break

        # 找到top1 与top2
        top = []
        apr1 = apr_dict[lpKeys[0]]
        target_apr_down_limit = Decimal(0.01)
        for key in lpKeys:

            if abs(apr_dict[key] - apr1) < detla:
                top.append(key)
            else:
                target_apr_down_limit = apr_dict[key]
                break

        """如果只有top只有一个值，那么直接下降到排名第二的apr，如果top有多个，就每次进行小量尝试"""
        target_apr = max(target_apr_down_limit, apr1 - detla) if len(top) > 1 else target_apr_down_limit

        for key in top:
            filled, vol, changes = fill_cap(chain, key, daily_reward_dict, tvl_dict, balance_info_dict, price_dict,
                                            target_apr)

            if key not in invest_calc_result:
                invest_calc_result[key] = {}

            for k, v in changes.items():
                if k not in invest_calc_result[key]:
                    invest_calc_result[key][k] = Decimal(0)

                invest_calc_result[key][k] += v * -1
                balance_info_dict[k][chain]['amount'] += v

            tvl_dict[key] += vol
            apr_dict[key] = daily_reward_dict[key] * 365 / tvl_dict[key]

            # 说明没有对应资产了
            if not filled:
                lpKeys.remove(key)
                apr_dict.pop(key)

    for k in [key for key in invest_calc_result.keys()]:
        if len(invest_calc_result[k].keys()) == 0:
            invest_calc_result.pop(k)

    """
    这个时候如果有稳定币和非稳定币同时剩下了，我们要考虑进行调整，比如 bnb busd usdt btc，在上述运行完成后，剩下了busd btc，然而我们没有对应的投资标的
    那么我们这时候可以考虑用busd 替换已经确定的投资组合中的usdt，然后将usdt和btc进行组合，这种情况下，只要新增reward多于 下降的reward就可以
    """

    logging.info('invest info:{}'.format(json.dumps(invest_calc_result, cls=utils.DecimalEncoder)))

    return invest_calc_result


def get_balance(balance_dict, currency, chain):
    if currency not in balance_dict:
        return Decimal(0)
    if chain not in balance_dict[currency]:
        return Decimal(0)
    return balance_dict[currency][chain]['amount']


def get_price(price_dict, currency):
    if currency not in price_dict:
        return Decimal(0)

    return price_dict[currency]


# 返回值 1. 是否填满了，2 填充量是多少 3.资产余额变化
def fill_cap(chain, strategy, daily_reward_dict, tvl_dict, balance_dict, price_dict, target_apr):
    cap = (daily_reward_dict[strategy] * Decimal(365) / target_apr - tvl_dict[strategy])
    data, c0, c1 = get_info_by_strategy_str(strategy)
    if c0 is None:
        return False, 0, {}

    # 单币
    if c1 is None:
        if get_price(price_dict, c0) == Decimal(0):
            return False, 0, {}

        vol = min(cap / get_price(price_dict, c0), get_balance(balance_dict, c0, chain))
        if vol <= 0:
            return False, 0, {}

        return cap == vol, vol, {c0: -1 * vol}

    # 双币
    vol = min(get_balance(balance_dict, c0, chain) * get_price(price_dict, c0),
              get_balance(balance_dict, c1, chain) * get_price(price_dict, c1), cap / 2)

    if vol <= 0:
        return False, 0, {}

    amount0 = vol / get_price(price_dict, c0)
    amount1 = vol / get_price(price_dict, c1)
    if amount0 > 0 and amount1 > 0:
        return cap == vol * 2, vol * 2, {c0: - amount0, c1: -amount1}

    return False, 0, {}


if __name__ == '__main__':
    # 读取config
    conf = read_yaml("./wang.yaml")

    db = create_engine(conf['db'],
                       encoding='utf-8',  # 编码格式
                       echo=False,  # 是否开启sql执行语句的日志输出
                       pool_recycle=-1  # 多久之后对线程池中的线程进行一次连接的回收（重置） （默认为-1）,其实session并不会被close
                       )
    session = sessionmaker(db)()

    currencies = {x.name: x for x in session.query(Currency).all()}

    for t in session.query(Token).all():
        curr = currencies[t.currency]
        if not hasattr(curr, 'tokens'):
            curr.tokens = {}
        curr.tokens[t.chain] = t

    session.close()

    while True:
        time.sleep(3)
        try:
            session = sessionmaker(db)()

            # 已经有小re了
            # tasks = find_part_re_balance_open_tasks(session)
            # if tasks is not None:
            #    continue

            params = calc_re_balance_params(conf, session, currencies)
            if params is None:
                continue
            # print('params:', params)
            print('params_json:', json.dumps(params, cls=utils.DecimalEncoder))

            create_part_re_balance_task(session, json.dumps(params, cls=utils.DecimalEncoder))
            session.commit()

        except Exception as e:
            print("except happens:{}".format(e))
            print(traceback.format_exc())
        finally:
            session.close()
