import json
import logging
import sys
from kafka import KafkaProducer
import socket

APP = 'reba-compute'
APPNAME = 'applog-heco'

class KafkaHandler(logging.Handler):
    """Class to instantiate the kafka logging facility."""

    def __init__(self, hosts,  topic=APPNAME, env_name='test', tls=None):
        """Initialize an instance of the kafka handler."""
        logging.Handler.__init__(self)
        self.producer = KafkaProducer(bootstrap_servers=hosts,
                                      value_serializer=lambda v: json.dumps(v).encode('utf-8'),
                                      linger_ms=10)
        self.topic = topic
        self.env_name = env_name

    def emit(self, record):
        """Emit the provided record to the kafka_client producer."""
        # drop kafka logging to avoid infinite recursion
        if 'kafka.' in record.name:
            return

        try:
            # apply the logger formatter
            #print(record, dir(record))
            msg = self.format(record)

            self.producer.send(self.topic, {'message': msg, 
                                            'app':APP, 
                                            'app_name':APP, 
                                            'hostname': socket.gethostname(),
                                            'env_name': self.env_name,
                                            'level'   : record.levelname,
                                            'filename': record.filename,
                                            'lineno'  :  record.lineno,
                                            })
            self.flush(timeout=1.0)
        except Exception:
            logging.Handler.handleError(self, record)

    def flush(self, timeout=None):
        """Flush the objects."""
        self.producer.flush(timeout=timeout)


    def close(self):
        """Close the producer and clean up."""
        self.acquire()
        try:
            if self.producer:
                self.producer.close()

            logging.Handler.close(self)
        finally:
            self.release()
            
if __name__ == '__main__':

    logger = logging.getLogger(__name__)
    # enable the debug logger if you want to see ALL of the lines
    #logging.basicConfig(level=logging.DEBUG)
    logger.setLevel(logging.ERROR)

    kh = KafkaHandler(['kafka-01.sinnet.huobiidc.com:9092'], APPNAME, 'dev')
    logger.addHandler(kh)

    logger.info("this is info")
    logger.debug("this is debug")
    logger.warning("this is warning")
    logger.error("this is error")
    logger.critical("this is critical")


