package msg

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/kafka"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbChat "Open_IM/pkg/proto/chat"
	"Open_IM/pkg/utils"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type rpcChat struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
	producer        *kafka.Producer
}

func NewRpcChatServer(port int) *rpcChat {
	log.NewPrivateLog("msg")
	rc := rpcChat{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImOfflineMessageName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
	rc.producer = kafka.NewKafkaProducer(config.Config.Kafka.Ws2mschat.Addr, config.Config.Kafka.Ws2mschat.Topic)
	return &rc
}

func (rpc *rpcChat) Run() {
	log.Info("", "", "rpc get_token init...")

	address := utils.ServerIP + ":" + strconv.Itoa(rpc.rpcPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("", "", "listen network failed, err = %s, address = %s", err.Error(), address)
		return
	}
	log.Info("", "", "listen network success, address = %s", address)

	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	//service registers with etcd

	pbChat.RegisterChatServer(srv, rpc)
	host := viper.GetString("endpoints.rpc_msg")
	err = getcdv3.RegisterEtcd(rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), host, rpc.rpcPort, rpc.rpcRegisterName, 10)
	if err != nil {
		log.Error("", "", "register rpc get_token to etcd failed, err = %s", err.Error())
		return
	}

	err = srv.Serve(listener)
	if err != nil {
		log.Info("", "", "rpc get_token fail, err = %s", err.Error())
		return
	}
	log.Info("", "", "rpc get_token init success")
}
