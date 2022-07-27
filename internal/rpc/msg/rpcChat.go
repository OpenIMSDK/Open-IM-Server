package msg

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/kafka"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/proto/msg"
	"Open_IM/pkg/utils"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
)

type rpcChat struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
	onlineProducer  *kafka.Producer
	//offlineProducer *kafka.Producer
	delMsgCh chan deleteMsg
}

type deleteMsg struct {
	UserID      string
	OpUserID    string
	SeqList     []uint32
	OperationID string
}

func NewRpcChatServer(port int) *rpcChat {
	log.NewPrivateLog(constant.LogFileName)
	rc := rpcChat{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImMsgName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
	rc.onlineProducer = kafka.NewKafkaProducer(config.Config.Kafka.Ws2mschat.Addr, config.Config.Kafka.Ws2mschat.Topic)
	//rc.offlineProducer = kafka.NewKafkaProducer(config.Config.Kafka.Ws2mschatOffline.Addr, config.Config.Kafka.Ws2mschatOffline.Topic)
	rc.delMsgCh = make(chan deleteMsg, 1000)
	return &rc
}

func (rpc *rpcChat) Run() {
	log.Info("", "rpcChat init...")
	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(rpc.rpcPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + rpc.rpcRegisterName)
	}
	log.Info("", "listen network success, address ", address)

	srv := grpc.NewServer()
	defer srv.GracefulStop()

	rpcRegisterIP := config.Config.RpcRegisterIP
	msg.RegisterMsgServer(srv, rpc)
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	err = getcdv3.RegisterEtcd(rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName, 10)
	if err != nil {
		log.Error("", "register rpcChat to etcd failed ", err.Error())
		return
	}
	go rpc.runCh()
	err = srv.Serve(listener)
	if err != nil {
		log.Error("", "rpc rpcChat failed ", err.Error())
		return
	}
	log.Info("", "rpc rpcChat init success")
}

func (rpc *rpcChat) runCh() {
	log.NewInfo("", "start del msg chan ")
	for {
		select {
		case msg := <-rpc.delMsgCh:
			log.NewInfo(msg.OperationID, utils.GetSelfFuncName(), "delmsgch recv new: ", msg)
			db.DB.DelMsgFromCache(msg.UserID, msg.SeqList, msg.OperationID)
			unexistSeqList, err := db.DB.DelMsgBySeqList(msg.UserID, msg.SeqList, msg.OperationID)
			if err != nil {
				log.NewError(msg.OperationID, utils.GetSelfFuncName(), "DelMsgBySeqList args: ", msg.UserID, msg.SeqList, msg.OperationID, err.Error())
				continue
			}
			if len(unexistSeqList) > 0 {
				DeleteMessageNotification(msg.OpUserID, msg.UserID, unexistSeqList, msg.OperationID)
			}
		}
	}
}
