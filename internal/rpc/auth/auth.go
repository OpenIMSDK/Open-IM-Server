package auth

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAuth "Open_IM/pkg/proto/auth"
	"Open_IM/pkg/utils"
	"context"
	"net"
	"strconv"
	"strings"

	"Open_IM/pkg/common/config"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func (rpc *rpcAuth) UserRegister(_ context.Context, req *pbAuth.UserRegisterReq) (*pbAuth.UserRegisterResp, error) {
	log.NewInfo(req.OperationID, "UserRegister args ", req.String())
	var user db.User
	utils.CopyStructFields(&user, req.UserInfo)
	if req.UserInfo.Birth != 0 {
		user.Birth = utils.UnixSecondToTime(int64(req.UserInfo.Birth))
	}
	err := imdb.UserRegister(user)
	if err != nil {
		log.NewError(req.OperationID, "UserRegister failed ", err.Error(), user)
		return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	log.NewInfo(req.OperationID, "rpc UserRegister return")
	return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{}}, nil
}

func (rpc *rpcAuth) UserToken(_ context.Context, req *pbAuth.UserTokenReq) (*pbAuth.UserTokenResp, error) {
	log.NewInfo(req.OperationID, "UserToken args ", req.String())

	_, err := imdb.GetUserByUserID(req.FromUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.FromUserID)
		return &pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	tokens, expTime, err := token_verify.CreateToken(req.FromUserID, req.Platform)
	if err != nil {
		log.NewError(req.OperationID, "CreateToken failed ", err.Error(), req.FromUserID, req.Platform)
		return &pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	log.NewInfo(req.OperationID, "rpc UserToken return ")
	return &pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{}, Token: tokens, ExpiredTime: expTime}, nil
}

type rpcAuth struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewRpcAuthServer(port int) *rpcAuth {
	log.NewPrivateLog("auth")
	return &rpcAuth{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImAuthName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (rpc *rpcAuth) Run() {
	log.NewInfo("0", "rpc auth start...")

	address := utils.ServerIP + ":" + strconv.Itoa(rpc.rpcPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.NewError("0", "listen network failed ", err.Error(), address)
		return
	}
	log.NewInfo("0", "listen network success, ", address, listener)
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	//service registers with etcd
	pbAuth.RegisterAuthServer(srv, rpc)
	host := viper.GetString("endpoints.rpc_auth")
	err = getcdv3.RegisterEtcd(rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), host, rpc.rpcPort, rpc.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error(),
			rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), host, rpc.rpcPort, rpc.rpcRegisterName)
		return
	}
	log.NewInfo("0", "RegisterAuthServer ok ", rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), utils.ServerIP, rpc.rpcPort, rpc.rpcRegisterName)
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "rpc auth ok")
}
