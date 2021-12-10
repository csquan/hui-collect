import yaml, os, time, sys
import apollo_client


CONFIG_SERVER_URL_TEST   = "http://apollo-config.system-service.huobiapps.com:80"
CONFIG_SERVER_URL_ONLINE = "http://apollo-config.system-service.apne-1.huobiapps.com:80"
#NAMESPACE_TEST   = "remote_config.base.yaml"
#NAMESPACE_ONLINE = "remote_config.base.yaml"
CLUSTER_TEST   = 'test-1'
CLUSTER_ONLINE = 'prd8'

class Config():

    app_id ="re-allocation"
    use_remote_config = None
    config_origin = None
    config_server_url = None
    namespace = None
    cluster = None
    local_config_file = None
    mode = None

    @staticmethod
    def read_file_yaml(path):
        with open(path, 'rb') as f:
            cf= f.read()
            cf = yaml.load(cf, Loader=yaml.CLoader)
            return cf

    @staticmethod
    def get_local_origin():
        return Config.read_file_yaml(Config.local_config_file)

    @staticmethod
    def get_remote_origin():
        client = apollo_client.ApolloClient(app_id=Config.app_id, cluster=Config.cluster, config_server_url=Config.config_server_url)
        #client.start()
        return client

    @staticmethod
    def get(key, defaultvalue=None):
        if Config.use_remote_config is None:
            raise(ValueError("use_remote_config : ", Config.use_remote_config))
        if Config.config_origin is None:
            if Config.use_remote_config is True:
                Config.config_origin = Config.get_remote_origin()
            else:
                Config.config_origin = Config.get_local_origin()
        if Config.use_remote_config is True:
            return Config.config_origin.get_value(key, defaultvalue, Config.namespace)
        else:
            return Config.config_origin.get(key, defaultvalue)
            
    @staticmethod
    def get_list(key):
        res = []
        for i in range(100):
            newkey = key + '[%d]'%i
            val = Config.get(newkey)
            if val is not None:
                res.append(val)
            else:
                break
        return res

    @staticmethod
    def get_config(filename):
        if not filename.endswith('.yaml')  :
            raise(Exception("Config file must ends with .yaml !!"))

        if filename.startswith('remote'):
            if 'test' in filename:
                Config.use_remote_config = True
                Config.config_server_url = CONFIG_SERVER_URL_TEST
                Config.namespace = filename
                Config.cluster = 'test-1'
                Config.mode = 'test'

            else:
                Config.use_remote_config = True
                Config.config_server_url = CONFIG_SERVER_URL_ONLINE
                Config.namespace = filename
                Config.cluster = 'prd8'
                Config.mode = 'online'
        else:
            Config.use_remote_config = False
            Config.local_config_file = filename
            Config.mode = 'dev'
        
        if Config.use_remote_config is False:
            return Config.get_local_origin()

        for i in range(10):
            time.sleep(3.0)
            try:
                appname = Config.get("apollo", {}).get("appname", '')
                if not appname:
                    print("get nothing : " + str(i))
                    continue
                elif appname != Config.app_id:
                    raise ValueError("get apollo.appname wrong : ", appname)
                else:
                    print("get apollo.appname : ", appname)
                    return Config.get_remote_origin().get_all(Config.namespace)     
            except Exception as e :
                print(e)
                continue
        else:
            raise ValueError("init apollo config wrong : not get apollo.appname at all ! ")


if __name__ == '__main__':
    print('**********************************************************')
    print(Config.get_config("remote_conf_test.yaml"))
    print('**********************************************************')

