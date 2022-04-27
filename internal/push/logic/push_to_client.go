/*
** description("").
** copyright('open-im,www.open-im.io').
** author("fg,Gordon@open-im.io").
** time(2021/3/5 14:31).
 */
package logic

import (
	pusher "Open_IM/internal/push"
	"Open_IM/internal/push/getui"
	jpush "Open_IM/internal/push/jpush"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	pbPush "Open_IM/pkg/proto/push"
	pbRelay "Open_IM/pkg/proto/relay"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
)

type OpenIMContent struct {
	SessionType int    `json:"sessionType"`
	From        string `json:"from"`
	To          string `json:"to"`
	Seq         uint32 `json:"seq"`
}
type AtContent struct {
	Text       string   `json:"text"`
	AtUserList []string `json:"atUserList"`
	IsAtSelf   bool     `json:"isAtSelf"`
}

func MsgToUser(pushMsg *pbPush.PushMsgReq) {
	var wsResult []*pbRelay.SingleMsgToUser
	isOfflinePush := utils.GetSwitchFromOptions(pushMsg.MsgData.Options, constant.IsOfflinePush)
	log.Debug("Get msg from msg_transfer And push msg", pushMsg.OperationID, "PushData", pushMsg.String())
	//grpcCons := getcdv3.GetConn4Unique(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOnlineMessageRelayName)
	grpcCons := watcher.GetAllConns()
	//Online push message
	log.Debug("test", pushMsg.OperationID, "len  grpc", len(grpcCons), "data", pushMsg.String())
	for _, v := range grpcCons {
		msgClient := pbRelay.NewOnlineMessageRelayServiceClient(v)
		reply, err := msgClient.OnlinePushMsg(context.Background(), &pbRelay.OnlinePushMsgReq{OperationID: pushMsg.OperationID, MsgData: pushMsg.MsgData, PushToUserID: pushMsg.PushToUserID})
		if err != nil {
			log.InfoByKv("push data to client rpc err", pushMsg.OperationID, "err", err)
			continue
		}
		if reply != nil && reply.Resp != nil {
			wsResult = append(wsResult, reply.Resp...)
		}
	}
	log.InfoByKv("push_result", pushMsg.OperationID, "result", wsResult, "sendData", pushMsg.MsgData)
	count++
	if isOfflinePush && pushMsg.PushToUserID != pushMsg.MsgData.SendID {
		for _, v := range wsResult {
			if v.ResultCode == 0 {
				continue
			}
			if utils.IsContainInt32(v.RecvPlatFormID, pushTerminal) {
				//Use offline push messaging
				var UIDList []string
				UIDList = append(UIDList, v.RecvID)
				customContent := OpenIMContent{
					SessionType: int(pushMsg.MsgData.SessionType),
					From:        pushMsg.MsgData.SendID,
					To:          pushMsg.MsgData.RecvID,
					Seq:         pushMsg.MsgData.Seq,
				}
				bCustomContent, _ := json.Marshal(customContent)
				jsonCustomContent := string(bCustomContent)
				var content string
				if pushMsg.MsgData.OfflinePushInfo != nil {
					content = pushMsg.MsgData.OfflinePushInfo.Title

				} else {
					switch pushMsg.MsgData.ContentType {
					case constant.Text:
						content = constant.ContentType2PushContent[constant.Text]
					case constant.Picture:
						content = constant.ContentType2PushContent[constant.Picture]
					case constant.Voice:
						content = constant.ContentType2PushContent[constant.Voice]
					case constant.Video:
						content = constant.ContentType2PushContent[constant.Video]
					case constant.File:
						content = constant.ContentType2PushContent[constant.File]
					case constant.AtText:
						a := AtContent{}
						_ = utils.JsonStringToStruct(string(pushMsg.MsgData.Content), &a)
						if utils.IsContain(v.RecvID, a.AtUserList) {
							content = constant.ContentType2PushContent[constant.AtText] + constant.ContentType2PushContent[constant.Common]
						} else {
							content = constant.ContentType2PushContent[constant.GroupMsg]
						}
					default:
						content = constant.ContentType2PushContent[constant.Common]
					}
				}
				var offlinePusher pusher.OfflinePusher
				if config.Config.Push.Getui.Enable {
					log.NewInfo(pushMsg.OperationID, utils.GetSelfFuncName(), config.Config.Push.Getui)
					offlinePusher = getui.GetuiClient
				}
				if config.Config.Push.Jpns.Enable {
					offlinePusher = jpush.JPushClient
				}
				if offlinePusher == nil {
					offlinePusher = jpush.JPushClient
				}
				pushResult, err := offlinePusher.Push(UIDList, content, jsonCustomContent, pushMsg.OperationID)
				if err != nil {
					log.NewError(pushMsg.OperationID, "offline push error", pushMsg.String(), err.Error())
				} else {
					log.NewDebug(pushMsg.OperationID, "offline push return result is ", pushResult, pushMsg.MsgData)
				}
				break
			}

		}

	}
}

//func SendMsgByWS(m *pbChat.WSToMsgSvrChatMsg) {
//	m.MsgID = rpcChat.GetMsgID(m.SendID)
//	m.ClientMsgID = m.MsgID
//	switch m.SessionType {
//	case constant.SingleChatType:
//		sendMsgToKafka(m, m.SendID, "msgKey--sendID")
//		sendMsgToKafka(m, m.RecvID, "msgKey--recvID")
//	case constant.GroupChatType:
//		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName)
//		client := pbGroup.NewGroupClient(etcdConn)
//		req := &pbGroup.Req{
//			GroupID:     m.RecvID,
//			Token:       config.Config.Secret,
//			OperationID: m.OperationID,
//		}
//		reply, err := client.(context.Background(), req)
//		if err != nil {
//			log.Error(m.Token, m.OperationID, "rpc  getGroupInfo failed, err = %s", err.Error())
//			return
//		}
//		if reply.ErrorCode != 0 {
//			log.Error(m.Token, m.OperationID, "rpc  getGroupInfo failed, err = %s", reply.ErrorMsg)
//			return
//		}
//		groupID := m.RecvID
//		for i, v := range reply.MemberList {
//			m.RecvID = v.UserId + " " + groupID
//			sendMsgToKafka(m, utils.IntToString(i), "msgKey--recvID+\" \"+groupID")
//		}
//	default:
//
//	}
//}
//
//func sendMsgToKafka(m *pbChat.WSToMsgSvrChatMsg, key string, flag string) {
//	pid, offset, err := producer.SendMessage(m, key)
//	if err != nil {
//		log.ErrorByKv("kafka send failed", m.OperationID, "send data", m.String(), "pid", pid, "offset", offset, "err", err.Error(), flag, key)
//	}
//
//}
