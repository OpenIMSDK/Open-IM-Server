package logic

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbPush "Open_IM/pkg/proto/push"
	"Open_IM/pkg/utils"
	"context"
	"net"
	"strings"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type RPCServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func (r *RPCServer) Init(rpcPort int) {
	r.rpcPort = rpcPort
	r.rpcRegisterName = config.Config.RpcRegisterName.OpenImPushName
	r.etcdSchema = config.Config.Etcd.EtcdSchema
	r.etcdAddr = config.Config.Etcd.EtcdAddr
}
func (r *RPCServer) run() {
	ip := utils.ServerIP
	registerAddress := ip + ":" + utils.IntToString(r.rpcPort)
	listener, err := net.Listen("tcp", registerAddress)
	if err != nil {
		log.ErrorByKv("push module rpc listening port err", "", "err", err.Error())
		return
	}
	defer listener.Close()
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	pbPush.RegisterPushMsgServiceServer(srv, r)
	host := viper.GetString("endpoints.push")
	err = getcdv3.RegisterEtcd(r.etcdSchema, strings.Join(r.etcdAddr, ","), host, r.rpcPort, r.rpcRegisterName, 10)
	if err != nil {
		log.ErrorByKv("register push module  rpc to etcd err", "", "err", err.Error())
	}
	err = srv.Serve(listener)
	if err != nil {
		log.ErrorByKv("push module rpc start err", "", "err", err.Error())
		return
	}
}
func (r *RPCServer) PushMsg(_ context.Context, pbData *pbPush.PushMsgReq) (*pbPush.PushMsgResp, error) {
	//Call push module to send message to the user
	MsgToUser(pbData)
	return &pbPush.PushMsgResp{
		ResultCode: 0,
	}, nil

}
