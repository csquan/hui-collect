from sqlalchemy.ext.declarative import *

from sqlalchemy import *

Base = declarative_base()


class Strategy(Base):
    __tablename__ = 't_strategy'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    chain = Column(String, name='f_chain')
    project = Column(String, name='f_project')
    currency0 = Column(String, name='f_currency0')
    currency1 = Column(String, name='f_currency1')

    def __str__(self):
        return "chain:{} project:{} currency0:{} currency1:{}".format(self.chain, self.project,
                                                                      self.currency0, self.currency1)


class Currency(Base):
    __tablename__ = 't_currency'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    name = Column(String, name='f_name')
    min = Column(Float, name='f_min')
    crossDecimal = Column(Integer, name='f_cross_scale')

    def __str__(self):
        return "name:{} min:{} crossScale:{} tokens:{}".format(self.name, self.min, self.crossDecimal, self.tokens)


class Token(Base):
    __tablename__ = 't_token'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    chain = Column(String, name='f_chain')
    currency = Column(String, name='f_currency')
    symbol = Column(String, name='f_symbol')
    address = Column(String, name='f_address')
    decimal = Column(Integer, name='f_decimal')
    crossSymbol = Column(String, name='f_cross_symbol')

    def __str__(self):
        return "chain:{} currency:{} symbol:{} address:{} decimal:{} crossSymbol:{}".format(self.chain, self.currency,
                                                                                            self.symbol, self.address,
                                                                                            self.decimal,
                                                                                            self.crossSymbol)

class FullReBalanceTask(Base):
    __tablename__ = 't_full_rebalance_task'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    state = Column(SmallInteger, name='f_state')
    params = Column(TEXT, name='f_params')
    message = Column(TEXT, name='f_message')


class PartReBalanceTask(Base):
    __tablename__ = 't_part_rebalance_task'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    full_rebalance_id = Column(Integer, name='f_full_rebalance_id')
    params = Column(TEXT, name='f_params')
    message = Column(TEXT, name='f_message')
    state = Column(SmallInteger, name='f_state')


def find_strategies_by_chain_and_currency(session, chain, currency):
    return session.query(Strategy).filter(Strategy.chain == chain).filter(
        or_(Strategy.currency0 == currency, Strategy.currency1 == currency))


def find_strategies_by_chain_project_and_currencies(session, chain, project, currency0, currency1):
    q = session.query(Strategy).filter(Strategy.chain == chain).filter(Strategy.project == project).filter(
        or_(and_(Strategy.currency0 == currency0, Strategy.currency1 == currency1),
            and_(Strategy.currency0 == currency1, Strategy.currency1 == currency0)))

    return [s for s in q]


def find_currency_by_address(session, address):
    token = [x for x in session.query(Token).filter(Token.address == address)]
    if len(token) == 0:
        return None

    if len(token) > 1:
        raise ValueError('find more than one token with address:{}', address)

    return token[0].currency


def find_part_re_balance_open_tasks(session):
    q = session.query(PartReBalanceTask).filter(PartReBalanceTask.state.not_in([5, 6]))
    tasks = [x for x in q]
    if len(tasks) == 0:
        return None

    return tasks

def find_full_re_balance_open_tasks(session):
    q = session.query(FullReBalanceTask).filter(FullReBalanceTask.state.in_([5]))
    tasks = [x for x in q]
    if len(tasks) == 0:
        return None

    return tasks

def create_part_re_balance_task(session, params):
    p = PartReBalanceTask()
    p.params = params
    p.message = ''
    p.state = 0

    session.add(p)

def create_part_re_balance_task_for_full(session, params, full_id):
    p = PartReBalanceTask()
    p.full_rebalance_id = full_id
    p.params = params
    p.message = ''
    p.state = 0
    session.add(p)
    q = session.query(FullReBalanceTask).filter(FullReBalanceTask.id.in_([full_id])).\
        update({FullReBalanceTask.state:6})

    
