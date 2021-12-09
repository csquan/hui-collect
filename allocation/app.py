import yaml, time, json, logging
from argparse import ArgumentParser
from sqlalchemy.orm import sessionmaker
from orm import *
from format import calc_re_balance_params
import utils
import traceback

def run(config):
    
    # to set default log
    logging.basicConfig(level=logging.INFO,
                        format='%(asctime)s %(filename)s[line:%(lineno)d] %(levelname)s %(message)s',
                        datefmt='%Y-%m-%d %H:%M:%S')
    # to config kafka log
    mode = 'dev'

    kafka_enable = config.get("kafka.enable")
    if isinstance(kafka_enable, str):
        kafka_enable = (kafka_enable.lower() == 'true')
    if kafka_enable:
        print("kafka logs starts...")
        from kafkalog import KafkaHandler
        logger = logging.getLogger()
        logger.setLevel(int(config.get("kafka.level")))
        try:
            hosts = config.get_list("kafka.hosts")
        except:
            hosts = config.get("kafka.hosts")
        #print(hosts)
        logger.addHandler(KafkaHandler(
            hosts = hosts,
            topic = config.get("kafka.topic"),
            env_name = mode))
    
    # to config database
    db = create_engine(config.get('db'),
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

    logging.info('start re-allocation program ... ')
    while True:
        time.sleep(3)
        try:
            session = sessionmaker(db)()

            # 有大re任务，需要拆解成小re的任务
            tasks = find_full_re_balance_open_tasks(session)
            session.commit()

            if tasks is not None:
                for task in tasks :
                    params = calc_re_balance_params(conf, session, currencies)
                    session.begin()
                    create_part_re_balance_task_for_full(session, json.dumps(params, cls=utils.DecimalEncoder), task.id)
                    try:
                        session.commit()
                    except Exception as e:
                        session.rollback()
                        logging.error("db except :{}".format(e) + '\n' + traceback.format_exc())
                        
            # 已经有小re了
            #tasks = find_part_re_balance_open_tasks(session)
            #if tasks is not None:
            #   continue

            params = calc_re_balance_params(config, session, currencies)
            session.commit()
            if params is None:
                continue
            #print('params:', params)
            print('params_json:', json.dumps(params, cls=utils.DecimalEncoder))
            
            session.begin()
            create_part_re_balance_task(session, json.dumps(params, cls=utils.DecimalEncoder))
            try:
                session.commit()
            except Exception as e:
                session.rollback()  
                logging.error("db except :{}".format(e) + '\n' + traceback.format_exc())
                  
        except Exception as e:
            logging.error("except happens:{}".format(e) + '\n' + traceback.format_exc())
        finally:
            session.close()



if __name__ == '__main__':

    p = ArgumentParser(usage='a re-allocation program, for investing', description='need a config file ')   
    p.add_argument('-c', '--config',type=str, required=True, help='the config file')  
    args = p.parse_args() 

    filename = args.config
    conf = yaml.load(open(filename), Loader=yaml.FullLoader)

    run(conf)

