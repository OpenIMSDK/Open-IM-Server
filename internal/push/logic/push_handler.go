/*
** description("").
** copyright('Open_IM,www.Open_IM.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/5/13 10:33).
 */
package logic

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/mq"
	kfk "Open_IM/pkg/common/mq/kafka"
	"Open_IM/pkg/common/mq/nsq"
	pbChat "Open_IM/pkg/proto/chat"
	pbRelay "Open_IM/pkg/proto/relay"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
)

type PushConsumerHandler struct {
	pushConsumerGroup mq.Consumer
}

func (ms *PushConsumerHandler) Init() {
	cfg:= config.Config.MQ.Ms2pschat
	switch cfg.Type {
	case "kafka":
		ms.pushConsumerGroup = kfk.NewMConsumerGroup(&kfk.MConsumerGroupConfig{KafkaVersion: sarama.V0_10_2_0,
			OffsetsInitial: sarama.OffsetNewest, IsReturnErr: false}, cfg.Addr,
			config.Config.MQ.ConsumerGroupID.MsgToPush)
	case "nsq":
		nc, err := nsq.NewNsqConsumer(cfg.Addr,cfg.Topic,cfg.Channel)
		if err != nil {
			panic(err)
		}
		ms.pushConsumerGroup = nc
	}

	ms.pushConsumerGroup.RegisterMessageHandler(cfg.Topic, mq.MessageHandleFunc(ms.handleMs2PsChat))
}
func (ms *PushConsumerHandler) handleMs2PsChat(message *mq.Message) error {
	msg := message.Value
	log.InfoByKv("msg come from mq  And push!!!", "", "msg", string(msg))
	pbData := pbChat.MsgSvrToPushSvrChatMsg{}
	if err := proto.Unmarshal(msg, &pbData); err != nil {
		log.ErrorByKv("push Unmarshal msg err", "", "msg", string(msg), "err", err.Error())
		return nil // not retry
	}
	sendPbData := pbRelay.MsgToUserReq{}
	sendPbData.SendTime = pbData.SendTime
	sendPbData.OperationID = pbData.OperationID
	sendPbData.ServerMsgID = pbData.MsgID
	sendPbData.MsgFrom = pbData.MsgFrom
	sendPbData.ContentType = pbData.ContentType
	sendPbData.SessionType = pbData.SessionType
	sendPbData.RecvID = pbData.RecvID
	sendPbData.Content = pbData.Content
	sendPbData.SendID = pbData.SendID
	sendPbData.SenderNickName = pbData.SenderNickName
	sendPbData.SenderFaceURL = pbData.SenderFaceURL
	sendPbData.ClientMsgID = pbData.ClientMsgID
	sendPbData.PlatformID = pbData.PlatformID
	sendPbData.RecvSeq = pbData.RecvSeq
	//Call push module to send message to the user
	MsgToUser(&sendPbData, pbData.OfflineInfo, pbData.Options)

	return nil
}
