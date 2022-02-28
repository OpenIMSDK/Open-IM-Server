package gate

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbRelay "Open_IM/pkg/proto/relay"
	"Open_IM/pkg/utils"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

type RPCServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func (r *RPCServer) onInit(rpcPort int) {
	r.rpcPort = rpcPort
	r.rpcRegisterName = config.Config.RpcRegisterName.OpenImOnlineMessageRelayName
	r.etcdSchema = config.Config.Etcd.EtcdSchema
	r.etcdAddr = config.Config.Etcd.EtcdAddr
}
func (r *RPCServer) run() {
	ip := utils.ServerIP
	registerAddress := ip + ":" + utils.IntToString(r.rpcPort)
	listener, err := net.Listen("tcp", registerAddress)
	if err != nil {
		log.ErrorByArgs(fmt.Sprintf("fail to listening consumer, err:%v\n", err))
		return
	}
	defer listener.Close()
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	pbRelay.RegisterOnlineMessageRelayServiceServer(srv, r)
	host := viper.GetString("endpoints.msg_gateway")
	err = getcdv3.RegisterEtcd4Unique(r.etcdSchema, strings.Join(r.etcdAddr, ","), host, r.rpcPort, r.rpcRegisterName, 10)
	if err != nil {
		log.ErrorByKv("register push message rpc to etcd err", "", "err", err.Error())
	}
	err = srv.Serve(listener)
	if err != nil {
		log.ErrorByKv("push message rpc listening err", "", "err", err.Error())
		return
	}
}
func (r *RPCServer) OnlinePushMsg(_ context.Context, in *pbRelay.OnlinePushMsgReq) (*pbRelay.OnlinePushMsgResp, error) {
	log.InfoByKv("PushMsgToUser is arriving", in.OperationID, "args", in.String())
	var resp []*pbRelay.SingleMsgToUser
	msgBytes, _ := proto.Marshal(in.MsgData)
	mReply := Resp{
		ReqIdentifier: constant.WSPushMsg,
		OperationID:   in.OperationID,
		Data:          msgBytes,
	}
	var replyBytes bytes.Buffer
	enc := gob.NewEncoder(&replyBytes)
	err := enc.Encode(mReply)
	if err != nil {
		log.NewError(in.OperationID, "data encode err", err.Error())
	}
	var tag bool
	recvID := in.PushToUserID
	platformList := genPlatformArray()
	for _, v := range platformList {
		if conn := ws.getUserConn(recvID, v); conn != nil {
			tag = true
			resultCode := sendMsgToUser(conn, replyBytes.Bytes(), in, v, recvID)
			temp := &pbRelay.SingleMsgToUser{
				ResultCode:     resultCode,
				RecvID:         recvID,
				RecvPlatFormID: constant.PlatformNameToID(v),
			}
			resp = append(resp, temp)
		} else {
			temp := &pbRelay.SingleMsgToUser{
				ResultCode:     -1,
				RecvID:         recvID,
				RecvPlatFormID: constant.PlatformNameToID(v),
			}
			resp = append(resp, temp)
		}
	}
	if !tag {
		log.NewError(in.OperationID, "push err ,no matched ws conn not in map", in.String())
	}
	return &pbRelay.OnlinePushMsgResp{
		Resp: resp,
	}, nil
}
func (r *RPCServer) GetUsersOnlineStatus(_ context.Context, req *pbRelay.GetUsersOnlineStatusReq) (*pbRelay.GetUsersOnlineStatusResp, error) {
	log.NewInfo(req.OperationID, "rpc GetUsersOnlineStatus arrived server", req.String())
	if !token_verify.IsMangerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "no permission GetUsersOnlineStatus ", req.OpUserID)
		return &pbRelay.GetUsersOnlineStatusResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}
	var resp pbRelay.GetUsersOnlineStatusResp
	for _, userID := range req.UserIDList {
		platformList := genPlatformArray()
		temp := new(pbRelay.GetUsersOnlineStatusResp_SuccessResult)
		temp.UserID = userID
		for _, platform := range platformList {
			if conn := ws.getUserConn(userID, platform); conn != nil {
				ps := new(pbRelay.GetUsersOnlineStatusResp_SuccessDetail)
				ps.Platform = platform
				ps.Status = constant.OnlineStatus
				temp.Status = constant.OnlineStatus
				temp.DetailPlatformStatus = append(temp.DetailPlatformStatus, ps)

			}
		}
		if temp.Status == constant.OnlineStatus {
			resp.SuccessResult = append(resp.SuccessResult, temp)
		}
	}
	log.NewInfo(req.OperationID, "GetUsersOnlineStatus rpc return ", resp.String())
	return &resp, nil
}
func sendMsgToUser(conn *UserConn, bMsg []byte, in *pbRelay.OnlinePushMsgReq, RecvPlatForm, RecvID string) (ResultCode int64) {
	err := ws.writeMsg(conn, websocket.BinaryMessage, bMsg)
	if err != nil {
		log.ErrorByKv("PushMsgToUser is failed By Ws", "", "Addr", conn.RemoteAddr().String(),
			"error", err, "senderPlatform", constant.PlatformIDToName(in.MsgData.SenderPlatformID), "recvPlatform", RecvPlatForm, "args", in.String(), "recvID", RecvID)
		ResultCode = -2
		return ResultCode
	} else {
		log.InfoByKv("PushMsgToUser is success By Ws", in.OperationID, "args", in.String(), "recvPlatForm", RecvPlatForm, "recvID", RecvID)
		ResultCode = 0
		return ResultCode
	}

}
func genPlatformArray() (array []string) {
	for i := 1; i <= constant.LinuxPlatformID; i++ {
		array = append(array, constant.PlatformIDToName(int32(i)))
	}
	return array
}
